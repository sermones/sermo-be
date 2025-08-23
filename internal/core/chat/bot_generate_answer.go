package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"sermo-be/internal/middleware"
	"sermo-be/internal/models"
	"sermo-be/pkg/database"
	"sermo-be/pkg/openai"
	"sermo-be/pkg/prompt"
)

// AnswerGenerator AI 응답 생성을 담당하는 구조체
type AnswerGenerator struct {
	messageService *MessageService
}

// NewAnswerGenerator 새로운 AnswerGenerator 생성
func NewAnswerGenerator() *AnswerGenerator {
	return &AnswerGenerator{
		messageService: GetMessageService(),
	}
}

// ChatbotInfo 채팅봇 정보
type ChatbotInfo struct {
	Name     string          `json:"name"`
	Gender   string          `json:"gender"`
	Details  string          `json:"details"`
	Hashtags json.RawMessage `json:"hashtags"`
	Summary  *string         `json:"summary"` // 추가: 요약 정보
}

// WeightedMessage 가중치가 적용된 메시지
type WeightedMessage struct {
	Content  string  `json:"content"`
	Weight   float64 `json:"weight"`
	Role     string  `json:"role"`
	IsRecent bool    `json:"is_recent"`
}

// DataCollectionResult 고루틴으로 수집된 데이터 결과
type DataCollectionResult struct {
	ChatbotInfo *ChatbotInfo
	History     []models.ChatMessage
	UserStatus  *models.UserStatus
	Err         error
}

// GenerateAnswer AI 응답 생성 및 저장
func (ag *AnswerGenerator) GenerateAnswer(session *middleware.SSESession, combinedMessage string, openaiClient *openai.Client) *models.ChatMessage {
	// 타이핑 이벤트 시작 전송
	ag.sendTypingEvent(session, true)

	// 1. 고루틴으로 필요한 데이터 병렬 조회
	dataResult := ag.collectDataParallel(session.UserUUID, session.ChatbotUUID, combinedMessage, openaiClient)
	if dataResult.Err != nil {
		ag.sendTypingEvent(session, false) // 타이핑 이벤트 종료
		return nil
	}

	// 2. 가중치 기반 대화 히스토리 구성
	weightedHistory := ag.buildWeightedHistory(dataResult.History)

	// 3. 초기 프롬프팅으로 응답 생성
	initialResponse, err := ag.generateInitialResponse(dataResult.ChatbotInfo, weightedHistory, dataResult.UserStatus, combinedMessage, openaiClient)
	if err != nil {
		ag.sendTypingEvent(session, false) // 타이핑 이벤트 종료
		return nil
	}

	// 4. 응답 검증 및 재조정 (2단계)
	finalResponse := ag.validateAndAdjustResponse(dataResult.ChatbotInfo, dataResult.UserStatus, initialResponse, combinedMessage, openaiClient)

	// 최종 응답 검증 - 빈 응답인 경우 처리
	if strings.TrimSpace(finalResponse) == "" {
		ag.sendTypingEvent(session, false) // 타이핑 이벤트 종료
		return nil
	}

	// 최종 응답 길이 검증
	if len(strings.TrimSpace(finalResponse)) < 1 {
		ag.sendTypingEvent(session, false) // 타이핑 이벤트 종료
		return nil
	}

	// 4. 봇 메시지 저장
	botChatMessage, err := ag.messageService.CreateBotMessage(
		session.SessionID,
		session.UserUUID,
		session.ChatbotUUID,
		finalResponse,
	)

	if err != nil {
		ag.sendTypingEvent(session, false) // 타이핑 이벤트 종료
		return nil
	}

	// 타이핑 이벤트 종료 전송
	ag.sendTypingEvent(session, false)

	return botChatMessage
}

// sendTypingEvent 타이핑 이벤트 전송
func (ag *AnswerGenerator) sendTypingEvent(session *middleware.SSESession, isTyping bool) {
	// 타이핑 이벤트 메시지 생성
	typingEvent := map[string]interface{}{
		"type":       "bot_typing",
		"is_typing":  isTyping,
		"session_id": session.SessionID,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	// JSON으로 직렬화
	eventData, err := json.Marshal(typingEvent)
	if err != nil {
		log.Printf("타이핑 이벤트 직렬화 실패: %v", err)
		return
	}

	// SSE 형식으로 변환
	sseMessage := "data: " + string(eventData) + "\n\n"

	// 세션 채널로 이벤트 전송
	select {
	case session.Channel <- sseMessage:
		if isTyping {
			log.Printf("봇 타이핑 시작 이벤트 전송 - 세션: %s", session.SessionID)
		} else {
			log.Printf("봇 타이핑 종료 이벤트 전송 - 세션: %s", session.SessionID)
		}
	default:
		log.Printf("세션 채널이 가득 참 - 타이핑 이벤트 전송 실패 - 세션: %s", session.SessionID)
	}
}

// collectDataParallel 고루틴으로 필요한 데이터를 병렬로 수집
func (ag *AnswerGenerator) collectDataParallel(userUUID, chatbotUUID, currentMessage string, openaiClient *openai.Client) *DataCollectionResult {
	result := &DataCollectionResult{}
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 채팅봇 정보 조회
	wg.Add(1)
	go func() {
		defer wg.Done()
		chatbotInfo, err := ag.getChatbotInfo(chatbotUUID, openaiClient)
		mu.Lock()
		if err != nil {
			result.Err = fmt.Errorf("채팅봇 정보 조회 실패: %w", err)
		} else {
			result.ChatbotInfo = chatbotInfo
		}
		mu.Unlock()
	}()

	// 대화 히스토리 조회
	wg.Add(1)
	go func() {
		defer wg.Done()
		history, err := ag.messageService.GetChatHistory(userUUID, chatbotUUID, 10) // 20개에서 10개로 줄임
		mu.Lock()
		if err != nil {
			result.Err = fmt.Errorf("대화 히스토리 조회 실패: %w", err)
		} else {
			result.History = history
		}
		mu.Unlock()
	}()

	// 사용자 상태 정보 조회
	wg.Add(1)
	go func() {
		defer wg.Done()
		userStatus, err := ag.getRelevantUserStatus(userUUID, chatbotUUID, currentMessage)
		mu.Lock()
		if err != nil {
			// 상태 정보 없어도 계속 진행 (에러가 아님)
		} else {
			result.UserStatus = userStatus
		}
		mu.Unlock()
	}()

	wg.Wait()

	// 에러가 있으면 반환
	if result.Err != nil {
		return result
	}

	// 필수 데이터 검증
	if result.ChatbotInfo == nil {
		result.Err = fmt.Errorf("채팅봇 정보가 없습니다")
		return result
	}

	if len(result.History) == 0 {
		result.Err = fmt.Errorf("대화 히스토리가 없습니다")
		return result
	}

	return result
}

// getChatbotInfo 채팅봇 정보 조회
func (ag *AnswerGenerator) getChatbotInfo(chatbotUUID string, openaiClient *openai.Client) (*ChatbotInfo, error) {
	var chatbot models.Chatbot
	if err := database.DB.Where("uuid = ?", chatbotUUID).First(&chatbot).Error; err != nil {
		return nil, fmt.Errorf("채팅봇 조회 실패: %w", err)
	}

	// 요약이 없으면 AI로 생성
	if chatbot.GetSummary() == nil {
		summary := ag.generateCharacterSummary(&chatbot, openaiClient)

		// 생성된 요약을 Chatbot 모델에 설정
		chatbot.SetSummary(summary)

		// 생성된 요약을 DB에 저장
		if err := database.DB.Model(&chatbot).Update("summary", summary).Error; err != nil {
			log.Printf("요약 저장 실패: %v", err)
		} else {
			log.Printf("요약 생성 및 저장 완료 - 길이: %d", len(summary))
		}
	}

	return &ChatbotInfo{
		Name:     chatbot.Name,
		Gender:   string(chatbot.Gender),
		Details:  chatbot.Details,
		Hashtags: chatbot.Hashtags,
		Summary:  chatbot.Summary, // 요약 정보 추가
	}, nil
}

// buildWeightedHistory 가중치 기반 대화 히스토리 구성
func (ag *AnswerGenerator) buildWeightedHistory(history []models.ChatMessage) []WeightedMessage {
	var weightedMessages []WeightedMessage
	totalMessages := len(history)

	for i, msg := range history {
		weight := 0.2 // 기본 가중치를 더 낮게 설정
		isRecent := false

		// 최근 메시지일수록 높은 가중치 (최근 메시지 우선)
		if i == totalMessages-1 {
			weight = 1.0 // 가장 최근 메시지는 최고 가중치
			isRecent = true
		} else if i >= totalMessages-2 {
			weight = 0.9 // 최근 2개 메시지는 높은 가중치
			isRecent = true
		} else if i >= totalMessages-5 {
			weight = 0.6 // 중간 가중치
		} else {
			weight = 0.2 // 오래된 메시지는 낮은 가중치 (문맥 파악용)
		}

		role := "user"
		if msg.MessageType == models.MessageTypeChatbot {
			role = "assistant"
		}

		weightedMessages = append(weightedMessages, WeightedMessage{
			Content:  msg.Content,
			Weight:   weight,
			Role:     role,
			IsRecent: isRecent,
		})
	}

	return weightedMessages
}

// getRelevantUserStatus 맥락에 맞는 사용자 상태 정보 조회
func (ag *AnswerGenerator) getRelevantUserStatus(userUUID, chatbotUUID, currentMessage string) (*models.UserStatus, error) {
	var userStatus models.UserStatus
	if err := database.DB.Where("user_uuid = ? AND chatbot_uuid = ? AND is_active = ?",
		userUUID, chatbotUUID, true).First(&userStatus).Error; err != nil {
		return nil, err
	}

	// 현재 메시지와 상태 정보의 맥락 일치성 검증
	if ag.isContextRelevant(currentMessage, userStatus.Event, userStatus.Context) {
		return &userStatus, nil
	}

	return nil, fmt.Errorf("맥락이 일치하지 않음")
}

// isContextRelevant 현재 메시지와 상태 정보의 맥락 일치성 검증
func (ag *AnswerGenerator) isContextRelevant(currentMessage, event, context string) bool {
	currentMessage = strings.ToLower(currentMessage)
	event = strings.ToLower(event)
	context = strings.ToLower(context)

	// 키워드 기반 맥락 일치성 검증
	keywords := []string{event, context}
	for _, keyword := range keywords {
		if keyword != "" && strings.Contains(currentMessage, keyword) {
			return true
		}
	}

	return false
}

// generateCharacterSummary AI를 이용해 캐릭터 상세 정보를 요약
func (ag *AnswerGenerator) generateCharacterSummary(chatbotInfo *models.Chatbot, openaiClient *openai.Client) string {
	// 상세 정보가 짧으면 요약 불필요
	if len(chatbotInfo.Details) < 200 {
		return chatbotInfo.Details
	}

	// 요약 프롬프트 구성 (pkg/prompt 사용)
	summaryPrompt := prompt.BuildCharacterSummaryPrompt(chatbotInfo.Name, string(chatbotInfo.Gender), chatbotInfo.Details)

	var messages []openai.ChatMessage
	messages = append(messages, openai.ChatMessage{
		Role:    "system",
		Content: "You are an expert at analyzing and summarizing AI chatbot personalities. Focus on capturing the character's unique speech patterns, vocabulary choices, and communication style. Create summaries that highlight what makes each character distinct in how they talk and express themselves.",
	})
	messages = append(messages, openai.ChatMessage{
		Role:    "user",
		Content: summaryPrompt,
	})

	// OpenAI API 호출하여 요약 생성
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := openaiClient.ChatCompletion(ctx, messages)
	if err != nil {
		return chatbotInfo.Details
	}

	summary := strings.TrimSpace(response.Message.Content)

	// 요약이 실패하거나 빈 문자열인 경우 원본 사용
	if summary == "" || len(summary) < 5 {
		return chatbotInfo.Details
	}

	return summary
}

// convertToPromptChatbotInfo ChatbotInfo를 prompt.ChatbotInfo로 변환
func convertToPromptChatbotInfo(chatbotInfo *ChatbotInfo, openaiClient *openai.Client) *prompt.ChatbotInfo {
	// 요약이 있으면 요약 사용, 없으면 상세 정보 사용
	details := chatbotInfo.Details
	if chatbotInfo.Summary != nil && *chatbotInfo.Summary != "" {
		details = *chatbotInfo.Summary
	}

	return &prompt.ChatbotInfo{
		Name:     chatbotInfo.Name,
		Gender:   chatbotInfo.Gender,
		Details:  details,
		Hashtags: chatbotInfo.Hashtags,
	}
}

// validateAndAdjustResponse 2단계: 응답 검증 및 재조정
func (ag *AnswerGenerator) validateAndAdjustResponse(chatbotInfo *ChatbotInfo, userStatus *models.UserStatus, initialResponse, currentMessage string, openaiClient *openai.Client) string {
	// ChatbotInfo를 prompt.ChatbotInfo로 변환
	promptChatbotInfo := convertToPromptChatbotInfo(chatbotInfo, openaiClient)

	// 검증 프롬프트 구성
	validationPrompt := prompt.BuildValidationPrompt(promptChatbotInfo, userStatus, currentMessage, initialResponse)

	// 영어 응답 강제 지시 추가
	validationPrompt += "\n\nCRITICAL: The response MUST be in English only. If the response is in Korean or any other language, convert it to natural English while maintaining the same meaning and tone."

	// 캐릭터의 고유한 말투와 성격 유지 강조
	validationPrompt += "\n\nCHARACTER AUTHENTICITY CHECK: Ensure the response maintains the character's unique speech patterns, vocabulary, and personality. If the response feels generic or doesn't match the character's established traits, adjust it to be more authentic to this specific character. Preserve any catchphrases or unique expressions that make the character distinct."

	// 검증 요청을 위한 메시지 구성
	validationMessages := []openai.ChatMessage{
		{
			Role:    "system",
			Content: "당신은 AI 응답을 검증하고 재조정하는 전문가입니다. 친구다운 자연스러운 대화가 되도록 검증하고 필요시 수정해주세요.",
		},
		{
			Role:    "user",
			Content: validationPrompt,
		},
	}

	// 검증 API 호출
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := openaiClient.ChatCompletion(ctx, validationMessages)
	if err != nil {
		return initialResponse // 실패시 원본 응답 사용
	}

	finalResponse := strings.TrimSpace(response.Message.Content)

	// 검증 결과가 빈 값이거나 너무 짧으면 원본 사용
	if finalResponse == "" || len(finalResponse) < 5 {
		return initialResponse
	}

	return finalResponse
}

// generateInitialResponse 초기 프롬프팅으로 응답 생성
func (ag *AnswerGenerator) generateInitialResponse(chatbotInfo *ChatbotInfo, weightedHistory []WeightedMessage,
	userStatus *models.UserStatus, currentMessage string, openaiClient *openai.Client) (string, error) {

	// 시스템 프롬프트 구성 (pkg/prompt 사용)
	systemPrompt := prompt.BuildSystemPrompt(convertToPromptChatbotInfo(chatbotInfo, openaiClient), userStatus)

	// 영어 응답 강제 프롬프트 추가
	systemPrompt += "\n\nIMPORTANT INSTRUCTION: You MUST respond in English only. Do not use Korean, Japanese, or any other language. Always use natural, conversational English that matches your character's personality."

	// 캐릭터의 고유한 말투와 성격 유지 강조
	systemPrompt += "\n\nCHARACTER CONSISTENCY: Stay true to your character's unique speech patterns, vocabulary, and personality. If you have specific catchphrases, speaking habits, or unique expressions, use them naturally. Avoid generic responses - make every response feel authentic to your specific character. Maintain your character's background, age, and personality traits throughout the conversation."

	// 대화 컨텍스트 구성
	var messages []openai.ChatMessage
	messages = append(messages, openai.ChatMessage{
		Role:    "system",
		Content: systemPrompt,
	})

	// 가중치가 높은 메시지부터 추가 (최근 메시지 우선)
	// 최근 메시지부터 역순으로 추가하여 최신 메시지가 마지막에 오도록 함
	for i := len(weightedHistory) - 1; i >= 0; i-- {
		msg := weightedHistory[i]
		// 가중치 기준을 더 엄격하게 설정하여 메시지 수 줄임
		if msg.Weight >= 0.7 && strings.TrimSpace(msg.Content) != "" && len(strings.TrimSpace(msg.Content)) > 2 {
			messages = append(messages, openai.ChatMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// 현재 사용자 메시지 추가 (가장 최근 메시지)
	messages = append(messages, openai.ChatMessage{
		Role:    "user",
		Content: currentMessage,
	})

	// 메시지 검증 - 모든 content가 유효한지 확인
	for i, msg := range messages {
		if strings.TrimSpace(msg.Content) == "" {
			return "", fmt.Errorf("메시지 %d의 content가 빈 값입니다", i)
		}
		if len(strings.TrimSpace(msg.Content)) < 1 {
			return "", fmt.Errorf("메시지 %d의 content가 너무 짧습니다: %s", msg.Content)
		}
	}

	// AI 응답 생성
	response, err := openaiClient.ChatCompletion(context.Background(), messages)
	if err != nil {
		return "", fmt.Errorf("AI 응답 생성 실패: %w", err)
	}

	return response.Message.Content, nil
}

// convertToPromptChatbotInfo ChatbotInfo를 prompt.ChatbotInfo로 변환
