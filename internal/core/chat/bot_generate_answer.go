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
	log.Printf("=== AI 응답 생성 시작 - 세션: %s ===", session.SessionID)
	log.Printf("사용자 메시지: %s", combinedMessage)

	// 1. 고루틴으로 필요한 데이터 병렬 조회
	dataResult := ag.collectDataParallel(session.UserUUID, session.ChatbotUUID, combinedMessage)
	if dataResult.Err != nil {
		log.Printf("데이터 수집 실패 - 세션: %s, 에러: %v", session.SessionID, dataResult.Err)
		return nil
	}

	log.Printf("데이터 수집 완료 - 채팅봇: %s, 히스토리 수: %d, 상태정보: %v",
		dataResult.ChatbotInfo.Name, len(dataResult.History), dataResult.UserStatus != nil)

	// 2. 가중치 기반 대화 히스토리 구성
	weightedHistory := ag.buildWeightedHistory(dataResult.History)
	log.Printf("가중치 히스토리 구성 완료 - 메시지 수: %d", len(weightedHistory))

	// 3. 초기 프롬프팅으로 응답 생성
	log.Printf("초기 AI 응답 생성 시작...")
	initialResponse, err := ag.generateInitialResponse(dataResult.ChatbotInfo, weightedHistory, dataResult.UserStatus, combinedMessage, openaiClient)
	if err != nil {
		log.Printf("초기 응답 생성 실패 - 세션: %s, 에러: %v", session.SessionID, err)
		return nil
	}
	log.Printf("초기 AI 응답 생성 완료 - 응답 길이: %d, 내용: %s", len(initialResponse), initialResponse)

	// 4. 응답 검증 및 재조정 (1번)
	log.Printf("응답 검증 및 재조정 시작...")
	finalResponse, err := ag.validateAndAdjustResponse(dataResult.ChatbotInfo, dataResult.UserStatus, combinedMessage, initialResponse, openaiClient)
	if err != nil {
		log.Printf("응답 검증 및 재조정 실패 - 세션: %s, 에러: %v", session.SessionID, err)
		// 재조정 실패시 초기 응답 사용
		finalResponse = initialResponse
		log.Printf("초기 응답을 최종 응답으로 사용 - 응답: %s", finalResponse)
	} else {
		log.Printf("응답 검증 및 재조정 완료 - 최종 응답 길이: %d, 내용: %s", len(finalResponse), finalResponse)
	}

	// 최종 응답 검증 - 빈 응답인 경우 초기 응답 사용
	if strings.TrimSpace(finalResponse) == "" {
		log.Printf("재조정된 응답이 빈 값입니다. 초기 응답을 사용합니다.")
		finalResponse = initialResponse
		log.Printf("빈 응답 대체 - 최종 응답: %s", finalResponse)
	}

	// 최종 응답 길이 검증
	if len(strings.TrimSpace(finalResponse)) < 1 {
		log.Printf("재조정된 응답이 너무 짧습니다. 초기 응답을 사용합니다.")
		finalResponse = initialResponse
		log.Printf("짧은 응답 대체 - 최종 응답: %s", finalResponse)
	}

	// 6. 문법 및 맞춤법 검증 및 수정
	log.Printf("문법 및 맞춤법 검증 시작...")
	grammarFixedResponse := ag.validateAndFixGrammar(finalResponse, openaiClient)
	if grammarFixedResponse != finalResponse {
		log.Printf("문법 수정 완료 - 수정 전: %s, 수정 후: %s", finalResponse, grammarFixedResponse)
		finalResponse = grammarFixedResponse
	} else {
		log.Printf("문법 검증 완료 - 수정 불필요")
	}

	// 7. 봇 메시지 저장
	log.Printf("봇 메시지 저장 시작...")
	botChatMessage, err := ag.messageService.CreateBotMessage(
		session.SessionID,
		session.UserUUID,
		session.ChatbotUUID,
		finalResponse,
	)

	if err != nil {
		log.Printf("봇 메시지 저장 실패 - 세션: %s, 에러: %v", session.SessionID, err)
		return nil
	}

	log.Printf("봇 메시지 저장 완료 - UUID: %s, 내용: %s", botChatMessage.UUID, botChatMessage.Content)
	log.Printf("=== AI 응답 생성 완료 - 세션: %s ===", session.SessionID)

	return botChatMessage
}

// collectDataParallel 고루틴으로 필요한 데이터를 병렬로 수집
func (ag *AnswerGenerator) collectDataParallel(userUUID, chatbotUUID, currentMessage string) *DataCollectionResult {
	log.Printf("=== 데이터 병렬 수집 시작 ===")
	log.Printf("사용자 UUID: %s, 채팅봇 UUID: %s", userUUID, chatbotUUID)

	result := &DataCollectionResult{}
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 채팅봇 정보 조회
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("채팅봇 정보 조회 시작...")
		chatbotInfo, err := ag.getChatbotInfo(chatbotUUID)
		mu.Lock()
		if err != nil {
			log.Printf("채팅봇 정보 조회 실패: %v", err)
			result.Err = fmt.Errorf("채팅봇 정보 조회 실패: %w", err)
		} else {
			log.Printf("채팅봇 정보 조회 성공 - 이름: %s, 성별: %s", chatbotInfo.Name, chatbotInfo.Gender)
			result.ChatbotInfo = chatbotInfo
		}
		mu.Unlock()
	}()

	// 대화 히스토리 조회
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("대화 히스토리 조회 시작...")
		history, err := ag.messageService.GetChatHistory(userUUID, chatbotUUID, 20)
		mu.Lock()
		if err != nil {
			log.Printf("대화 히스토리 조회 실패: %v", err)
			result.Err = fmt.Errorf("대화 히스토리 조회 실패: %w", err)
		} else {
			log.Printf("대화 히스토리 조회 성공 - 메시지 수: %d", len(history))
			result.History = history
		}
		mu.Unlock()
	}()

	// 사용자 상태 정보 조회
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("사용자 상태 정보 조회 시작...")
		userStatus, err := ag.getRelevantUserStatus(userUUID, chatbotUUID, currentMessage)
		mu.Lock()
		if err != nil {
			// 상태 정보 없어도 계속 진행 (에러가 아님)
			log.Printf("사용자 상태 정보 없음 또는 맥락 불일치: %v", err)
		} else {
			log.Printf("사용자 상태 정보 조회 성공 - 이벤트: %s, 컨텍스트: %s", userStatus.Event, userStatus.Context)
			result.UserStatus = userStatus
		}
		mu.Unlock()
	}()

	wg.Wait()
	log.Printf("=== 데이터 병렬 수집 완료 ===")

	// 에러가 있으면 반환
	if result.Err != nil {
		log.Printf("데이터 수집 중 에러 발생: %v", result.Err)
		return result
	}

	// 필수 데이터 검증
	if result.ChatbotInfo == nil {
		log.Printf("채팅봇 정보가 없습니다")
		result.Err = fmt.Errorf("채팅봇 정보가 없습니다")
		return result
	}

	if len(result.History) == 0 {
		log.Printf("대화 히스토리가 없습니다")
		result.Err = fmt.Errorf("대화 히스토리가 없습니다")
		return result
	}

	log.Printf("데이터 검증 통과 - 모든 필수 데이터 수집 완료")
	return result
}

// getChatbotInfo 채팅봇 정보 조회
func (ag *AnswerGenerator) getChatbotInfo(chatbotUUID string) (*ChatbotInfo, error) {
	var chatbot models.Chatbot
	if err := database.DB.Where("uuid = ?", chatbotUUID).First(&chatbot).Error; err != nil {
		return nil, fmt.Errorf("채팅봇 조회 실패: %w", err)
	}

	return &ChatbotInfo{
		Name:     chatbot.Name,
		Gender:   string(chatbot.Gender),
		Details:  chatbot.Details,
		Hashtags: chatbot.Hashtags,
	}, nil
}

// buildWeightedHistory 가중치 기반 대화 히스토리 구성
func (ag *AnswerGenerator) buildWeightedHistory(history []models.ChatMessage) []WeightedMessage {
	var weightedMessages []WeightedMessage
	totalMessages := len(history)

	for i, msg := range history {
		weight := 0.3 // 기본 가중치를 낮게 설정
		isRecent := false

		// 최근 메시지일수록 높은 가중치 (최근 메시지 우선)
		if i == totalMessages-1 {
			weight = 1.0 // 가장 최근 메시지는 최고 가중치
			isRecent = true
		} else if i >= totalMessages-3 {
			weight = 0.8 // 최근 3개 메시지는 높은 가중치
			isRecent = true
		} else if i >= totalMessages-8 {
			weight = 0.5 // 중간 가중치
		} else {
			weight = 0.3 // 오래된 메시지는 낮은 가중치 (문맥 파악용)
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

// summarizeCharacterDetails 캐릭터 상세 정보를 핵심 특징으로 요약
func (ag *AnswerGenerator) summarizeCharacterDetails(chatbotInfo *ChatbotInfo, openaiClient *openai.Client) string {
	// 상세 정보가 짧으면 요약 불필요
	if len(chatbotInfo.Details) < 200 {
		return chatbotInfo.Details
	}

	log.Printf("캐릭터 상세 정보 요약 시작 - 길이: %d", len(chatbotInfo.Details))

	// 요약 프롬프트 구성
	summaryPrompt := fmt.Sprintf(`다음은 AI 채팅봇의 상세한 성격 설명입니다. 
이를 3-4개의 핵심 특징과 간단한 세계관 설정으로 요약해주세요.
요약은 100자 이내로 작성하고, 한국어로 답변해주세요.

채팅봇 이름: %s
성별: %s
상세 설명: %s

핵심 특징과 세계관:`, chatbotInfo.Name, chatbotInfo.Gender, chatbotInfo.Details)

	var messages []openai.ChatMessage
	messages = append(messages, openai.ChatMessage{
		Role:    "system",
		Content: "당신은 AI 채팅봇의 성격을 간결하게 요약하는 전문가입니다.",
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
		log.Printf("캐릭터 요약 생성 실패: %v, 원본 상세 정보 사용", err)
		return chatbotInfo.Details
	}

	summary := strings.TrimSpace(response.Message.Content)
	log.Printf("캐릭터 요약 완료 - 원본 길이: %d, 요약 길이: %d", len(chatbotInfo.Details), len(summary))
	log.Printf("요약된 내용: %s", summary)

	// 요약이 실패하거나 빈 문자열인 경우 원본 사용
	if summary == "" || len(summary) < 10 {
		log.Printf("캐릭터 요약 실패 또는 너무 짧음 - 원본 상세 정보 사용")
		return chatbotInfo.Details
	}

	return summary
}

// convertToPromptChatbotInfo ChatbotInfo를 prompt.ChatbotInfo로 변환
func convertToPromptChatbotInfo(chatbotInfo *ChatbotInfo, openaiClient *openai.Client) *prompt.ChatbotInfo {
	// 캐릭터 상세 정보 요약
	summarizedDetails := chatbotInfo.Details
	if openaiClient != nil {
		// AnswerGenerator 인스턴스 생성하여 요약 함수 호출
		ag := &AnswerGenerator{}
		summarizedDetails = ag.summarizeCharacterDetails(chatbotInfo, openaiClient)
	}

	return &prompt.ChatbotInfo{
		Name:     chatbotInfo.Name,
		Gender:   chatbotInfo.Gender,
		Details:  summarizedDetails,
		Hashtags: chatbotInfo.Hashtags,
	}
}

// generateInitialResponse 초기 프롬프팅으로 응답 생성
func (ag *AnswerGenerator) generateInitialResponse(chatbotInfo *ChatbotInfo, weightedHistory []WeightedMessage,
	userStatus *models.UserStatus, currentMessage string, openaiClient *openai.Client) (string, error) {

	log.Printf("=== 초기 AI 응답 생성 상세 과정 ===")

	// 시스템 프롬프트 구성 (pkg/prompt 사용)
	systemPrompt := prompt.BuildSystemPrompt(convertToPromptChatbotInfo(chatbotInfo, openaiClient), userStatus)
	log.Printf("시스템 프롬프트 구성 완료 - 길이: %d", len(systemPrompt))
	log.Printf("시스템 프롬프트 내용: %s", systemPrompt)

	// 대화 컨텍스트 구성
	var messages []openai.ChatMessage
	messages = append(messages, openai.ChatMessage{
		Role:    "system",
		Content: systemPrompt,
	})

	// 가중치가 높은 메시지부터 추가 (최근 메시지 우선)
	log.Printf("대화 히스토리 메시지 구성 시작...")
	// 최근 메시지부터 역순으로 추가하여 최신 메시지가 마지막에 오도록 함
	for i := len(weightedHistory) - 1; i >= 0; i-- {
		msg := weightedHistory[i]
		// 빈 내용이나 너무 짧은 메시지는 제외
		if msg.Weight >= 0.5 && strings.TrimSpace(msg.Content) != "" && len(strings.TrimSpace(msg.Content)) > 2 {
			messages = append(messages, openai.ChatMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
			log.Printf("히스토리 메시지 추가 - Role: %s, Weight: %.1f, Content: %s", msg.Role, msg.Weight, msg.Content)
		} else {
			log.Printf("히스토리 메시지 제외 - Role: %s, Weight: %.1f, Content: %s (빈 내용 또는 가중치 부족)", msg.Role, msg.Weight, msg.Content)
		}
	}

	// 현재 사용자 메시지 추가 (가장 최근 메시지)
	messages = append(messages, openai.ChatMessage{
		Role:    "user",
		Content: currentMessage,
	})
	log.Printf("현재 사용자 메시지 추가 (최우선) - Content: %s", currentMessage)

	log.Printf("최종 메시지 구성 완료 - 총 메시지 수: %d", len(messages))
	log.Printf("메시지 순서: 시스템 프롬프트 -> 히스토리 -> 최근 사용자 메시지")

	// 메시지 검증 - 모든 content가 유효한지 확인
	for i, msg := range messages {
		if strings.TrimSpace(msg.Content) == "" {
			log.Printf("경고: 메시지 %d의 content가 빈 값입니다 - Role: %s", i, msg.Role)
			return "", fmt.Errorf("메시지 %d의 content가 빈 값입니다", i)
		}
		if len(strings.TrimSpace(msg.Content)) < 1 {
			log.Printf("경고: 메시지 %d의 content가 너무 짧습니다 - Role: %s, Content: %s", i, msg.Role, msg.Content)
			return "", fmt.Errorf("메시지 %d의 content가 너무 짧습니다: %s", msg.Content)
		}
	}
	log.Printf("모든 메시지 검증 통과")

	// AI 응답 생성
	log.Printf("OpenAI API 호출 시작...")
	response, err := openaiClient.ChatCompletion(context.Background(), messages)
	if err != nil {
		log.Printf("OpenAI API 호출 실패: %v", err)
		return "", fmt.Errorf("AI 응답 생성 실패: %w", err)
	}

	log.Printf("OpenAI API 응답 수신 - 응답 길이: %d", len(response.Message.Content))
	log.Printf("OpenAI API 응답 내용: %s", response.Message.Content)

	return response.Message.Content, nil
}

// validateAndFixGrammar 문법과 맞춤법을 검증하고 수정
func (ag *AnswerGenerator) validateAndFixGrammar(response string, openaiClient *openai.Client) string {
	// 응답이 너무 짧으면 검증 불필요
	if len(strings.TrimSpace(response)) < 5 {
		return response
	}

	log.Printf("문법 및 맞춤법 검증 시작 - 응답 길이: %d", len(response))

	// 문법 검증 프롬프트 구성
	grammarPrompt := fmt.Sprintf(`다음 한국어 응답의 문법과 맞춤법을 검증하고 수정해주세요.
잘못된 부분이 있다면 수정하고, 맞다면 원본을 그대로 반환하세요.
수정 시에는 자연스럽게 수정하고, 원본의 의미와 톤을 유지해주세요.

원본 응답:
%s

수정된 응답:`, response)

	var messages []openai.ChatMessage
	messages = append(messages, openai.ChatMessage{
		Role:    "system",
		Content: "당신은 한국어 문법과 맞춤법을 검증하고 수정하는 전문가입니다. 자연스럽게 수정하고, 원본의 의미를 유지해주세요.",
	})
	messages = append(messages, openai.ChatMessage{
		Role:    "user",
		Content: grammarPrompt,
	})

	// OpenAI API 호출하여 문법 검증 및 수정
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	grammarResponse, err := openaiClient.ChatCompletion(ctx, messages)
	if err != nil {
		log.Printf("문법 검증 실패: %v, 원본 응답 사용", err)
		return response
	}

	fixedResponse := strings.TrimSpace(grammarResponse.Message.Content)

	// 수정된 응답이 빈 값이거나 너무 짧으면 원본 사용
	if fixedResponse == "" || len(fixedResponse) < 3 {
		log.Printf("문법 검증 결과가 유효하지 않음 - 원본 응답 사용")
		return response
	}

	// 원본과 동일하면 수정 불필요
	if fixedResponse == response {
		log.Printf("문법 검증 완료 - 수정 불필요")
		return response
	}

	log.Printf("문법 검증 완료 - 수정됨")
	log.Printf("원본: %s", response)
	log.Printf("수정: %s", fixedResponse)

	return fixedResponse
}

// validateAndAdjustResponse 응답 검증 및 재조정 (1번)
func (ag *AnswerGenerator) validateAndAdjustResponse(chatbotInfo *ChatbotInfo, userStatus *models.UserStatus,
	currentMessage, initialResponse string, openaiClient *openai.Client) (string, error) {

	log.Printf("=== 응답 검증 및 재조정 상세 과정 ===")

	// 검증 프롬프트 구성 (pkg/prompt 사용)
	validationPrompt := prompt.BuildValidationPrompt(convertToPromptChatbotInfo(chatbotInfo, openaiClient), userStatus, currentMessage, initialResponse)
	log.Printf("검증 프롬프트 구성 완료 - 길이: %d", len(validationPrompt))
	log.Printf("검증 프롬프트 내용: %s", validationPrompt)

	var messages []openai.ChatMessage
	messages = append(messages, openai.ChatMessage{
		Role:    "system",
		Content: validationPrompt,
	})

	log.Printf("검증용 메시지 구성 완료 - 메시지 수: %d", len(messages))

	// 메시지 검증 - 모든 content가 유효한지 확인
	for i, msg := range messages {
		if strings.TrimSpace(msg.Content) == "" {
			log.Printf("경고: 검증 메시지 %d의 content가 빈 값입니다 - Role: %s", i, msg.Role)
			return "", fmt.Errorf("검증 메시지 %d의 content가 빈 값입니다", i)
		}
		if len(strings.TrimSpace(msg.Content)) < 1 {
			log.Printf("경고: 검증 메시지 %d의 content가 너무 짧습니다 - Role: %s, Content: %s", i, msg.Role, msg.Content)
			return "", fmt.Errorf("검증 메시지 %d의 content가 너무 짧습니다: %s", msg.Content)
		}
	}
	log.Printf("모든 검증 메시지 검증 통과")

	// AI 응답 재조정
	log.Printf("OpenAI API 재조정 호출 시작...")
	response, err := openaiClient.ChatCompletion(context.Background(), messages)
	if err != nil {
		log.Printf("OpenAI API 재조정 호출 실패: %v", err)
		return "", fmt.Errorf("응답 재조정 실패: %w", err)
	}

	log.Printf("OpenAI API 재조정 응답 수신 - 응답 길이: %d", len(response.Message.Content))
	log.Printf("OpenAI API 재조정 응답 내용: %s", response.Message.Content)

	// 재조정된 응답 검증
	adjustedResponse := strings.TrimSpace(response.Message.Content)
	if adjustedResponse == "" {
		log.Printf("경고: 재조정된 응답이 빈 값입니다!")
		return "", fmt.Errorf("재조정된 응답이 빈 값입니다")
	}

	if len(adjustedResponse) < 5 {
		log.Printf("경고: 재조정된 응답이 너무 짧습니다! 길이: %d", len(adjustedResponse))
		return "", fmt.Errorf("재조정된 응답이 너무 짧습니다: %s", adjustedResponse)
	}

	log.Printf("재조정된 응답 검증 통과 - 길이: %d", len(adjustedResponse))
	return adjustedResponse, nil
}
