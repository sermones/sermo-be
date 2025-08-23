package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"sermo-be/internal/core/status"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"
	"sermo-be/pkg/openai"
	"sermo-be/pkg/prompt"
)

// BotMessage 봇 메시지 구조
type BotMessage struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	SessionID string `json:"session_id"`
}

// UserMessage 사용자 메시지 구조
type UserMessage struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	SessionID string `json:"session_id"`
}

// OnKeyboardMessage 키보드 입력 이벤트 구조
type OnKeyboardMessage struct {
	Type      string `json:"type"`
	SessionID string `json:"session_id"`
	Timestamp string `json:"timestamp"`
}

// BotGoroutine 봇 고루틴 관리자
type BotGoroutine struct {
	messageService *MessageService
	statusService  *status.StatusService
	openaiClient   *openai.Client
}

// NewBotGoroutine 새로운 봇 고루틴 관리자 생성
func NewBotGoroutine() *BotGoroutine {
	return &BotGoroutine{
		messageService: GetMessageService(),
		statusService:  status.GetStatusService(),
		openaiClient:   nil, // StartBotGoroutine에서 설정됨
	}
}

// StartBotGoroutine 봇 고루틴 시작
func (bg *BotGoroutine) StartBotGoroutine(session *middleware.SSESession, openaiClient *openai.Client) chan string {
	// OpenAI 클라이언트 설정
	bg.openaiClient = openaiClient

	// 봇 고루틴 전용 채널 생성
	botChannel := make(chan string, 100)

	go func() {
		log.Printf("봇 고루틴 시작 - 세션: %s", session.SessionID)
		bg.runBotGoroutine(session, botChannel, openaiClient)
	}()

	return botChannel
}

// runBotGoroutine 봇 고루틴 메인 로직
func (bg *BotGoroutine) runBotGoroutine(session *middleware.SSESession, botChannel chan string, openaiClient *openai.Client) {
	// 메시지 버퍼와 타이머
	var messageBuffer []string
	var responseTimer *time.Timer

	// 봇 채널에서 메시지 수신
	for session.IsActive {
		select {
		case message := <-botChannel:
			// 사용자 메시지 처리
			bg.handleIncomingMessage(session, message, &messageBuffer, &responseTimer, openaiClient)

		case <-session.Done:
			// 세션 종료 신호
			bg.handleSessionDone(session, responseTimer)
			return
		}
	}

	log.Printf("봇 고루틴 종료 - 세션: %s", session.SessionID)
}

// handleIncomingMessage 들어오는 메시지 처리
func (bg *BotGoroutine) handleIncomingMessage(session *middleware.SSESession, message string, messageBuffer *[]string, responseTimer **time.Timer, openaiClient *openai.Client) {
	// 메시지 파싱
	sseMessage, err := bg.parseUserMessage(message)
	if err != nil {
		log.Printf("사용자 메시지 파싱 실패: %v", err)
		return
	}

	// 메시지 타입별로 처리
	switch sseMessage.Type {
	case "user":
		bg.processUserMessage(session, sseMessage, messageBuffer, responseTimer, openaiClient)
	case "onkeyboard":
		bg.processOnKeyboardEvent(session, responseTimer)
	default:
		log.Printf("알 수 없는 메시지 타입: %s", sseMessage.Type)
	}
}

// parseUserMessage 사용자 메시지 파싱
func (bg *BotGoroutine) parseUserMessage(message string) (*UserMessage, error) {
	// SSE 메시지에서 "data: " 접두사 제거
	jsonData := message
	if len(message) > 6 && message[:6] == "data: " {
		jsonData = message[6:]
	}

	var userMessage UserMessage
	if err := json.Unmarshal([]byte(jsonData), &userMessage); err != nil {
		return nil, err
	}
	return &userMessage, nil
}

// processUserMessage 사용자 메시지 처리
func (bg *BotGoroutine) processUserMessage(session *middleware.SSESession, sseMessage *UserMessage, messageBuffer *[]string, responseTimer **time.Timer, openaiClient *openai.Client) {
	log.Printf("사용자 메시지 수신 - 세션: %s, 내용: %s", session.SessionID, sseMessage.Content)

	// 기존 타이머가 있다면 취소
	if *responseTimer != nil {
		(*responseTimer).Stop()
	}

	// 메시지를 버퍼에 추가
	*messageBuffer = append(*messageBuffer, sseMessage.Content)

	// 5초 후 버퍼에 쌓인 모든 메시지로 봇 응답 생성
	*responseTimer = time.AfterFunc(5*time.Second, func() {
		bg.generateAIResponse(session, *messageBuffer, openaiClient)
	})
}

// processOnKeyboardEvent 키보드 입력 이벤트 처리
func (bg *BotGoroutine) processOnKeyboardEvent(session *middleware.SSESession, responseTimer **time.Timer) {
	log.Printf("키보드 입력 이벤트 수신 - 세션: %s", session.SessionID)

	// 기존 타이머가 있다면 취소
	if *responseTimer != nil {
		(*responseTimer).Stop()
	}

	// 키보드 입력 중일 때는 최소 5초 대기
	*responseTimer = time.AfterFunc(5*time.Second, func() {
		// 5초 후에 타이머를 nil로 설정하여 추가 대기 없이 즉시 응답 가능하게 함
		*responseTimer = nil
		log.Printf("키보드 입력 대기 완료 - 세션: %s", session.SessionID)
	})
}

// generateAIResponse AI 응답 생성
func (bg *BotGoroutine) generateAIResponse(session *middleware.SSESession, messageBuffer []string, openaiClient *openai.Client) {
	if !session.IsActive {
		return
	}

	// 버퍼에 메시지가 없으면 리턴
	if len(messageBuffer) == 0 {
		return
	}

	log.Printf("봇 응답 생성 시작 - 세션: %s, 버퍼 메시지 수: %d", session.SessionID, len(messageBuffer))

	// 버퍼의 모든 메시지를 하나의 컨텍스트로 결합
	combinedMessage := bg.combineMessages(messageBuffer)
	log.Printf("결합된 메시지: %s", combinedMessage)

	// 이전 대화 내용 가져오기 (최근 20개)
	history, err := bg.messageService.GetChatHistory(session.UserUUID, session.ChatbotUUID, 20)
	if err != nil {
		log.Printf("대화 히스토리 조회 실패 - 세션: %s, 에러: %v", session.SessionID, err)
		return
	}

	// OpenAI 클라이언트가 nil인지 확인
	if openaiClient == nil {
		log.Printf("OpenAI 클라이언트가 nil임 - 세션: %s", session.SessionID)
		return
	}

	// 대화 컨텍스트 구성
	var messages []openai.ChatMessage
	messages = append(messages, openai.ChatMessage{
		Role:    "system",
		Content: "당신은 친근하고 도움이 되는 AI 어시스턴트입니다.",
	})

	// 이전 대화 내용 추가
	for _, msg := range history {
		role := "user"
		if msg.MessageType == models.MessageTypeChatbot {
			role = "assistant"
		}
		messages = append(messages, openai.ChatMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	// 결합된 사용자 메시지 추가
	messages = append(messages, openai.ChatMessage{
		Role:    "user",
		Content: combinedMessage,
	})

	// 동시에 두 개의 고루틴으로 처리
	responseChan := make(chan *openai.ChatResponse, 1)
	statusChan := make(chan *status.StatusExtractionResult, 1)

	// 고루틴 1: AI 답변 생성
	go func() {
		response, err := openaiClient.ChatCompletion(context.Background(), messages)
		if err != nil {
			log.Printf("AI 응답 생성 실패 - 세션: %s, 에러: %v", session.SessionID, err)
			responseChan <- nil
			return
		}
		responseChan <- response
	}()

	// 고루틴 2: 상태 정보 추출
	go func() {
		statusResult := bg.extractUserStatus(combinedMessage, openaiClient)
		statusChan <- statusResult
	}()

	// AI 응답 대기
	response := <-responseChan
	if response == nil {
		log.Printf("AI 응답 생성 실패 - 세션: %s", session.SessionID)
		return
	}

	// 봇 메시지 저장
	botChatMessage, err := bg.messageService.CreateBotMessage(
		session.SessionID,
		session.UserUUID,
		session.ChatbotUUID,
		response.Message.Content,
	)

	if err != nil {
		log.Printf("봇 메시지 저장 실패 - 세션: %s, 에러: %v", session.SessionID, err)
		return
	}

	// 봇 응답을 session.Channel로 전송
	bg.sendBotMessage(session, botChatMessage)

	// 상태 정보 결과 대기 및 저장
	statusResult := <-statusChan
	if statusResult != nil && statusResult.NeedsSave {
		log.Printf("상태 정보 저장 시작 - 이벤트: %s, 유효시간: %s", statusResult.Event, statusResult.ValidUntil)
		go bg.saveUserStatus(session, statusResult.Event, statusResult.ValidUntil, statusResult.Context)
	} else {
		log.Printf("상태 정보 저장 불필요 - 일반 대화 응답")
	}
}

// combineMessages 버퍼에 쌓인 메시지들을 하나의 컨텍스트로 결합
func (bg *BotGoroutine) combineMessages(messages []string) string {
	if len(messages) == 0 {
		return ""
	}

	if len(messages) == 1 {
		return messages[0]
	}

	// 여러 메시지가 있을 경우 연결
	combined := ""
	for i, msg := range messages {
		if i > 0 {
			combined += "\n"
		}
		combined += msg
	}

	return combined
}

// sendBotMessage 봇 메시지를 session.Channel로 전송
func (bg *BotGoroutine) sendBotMessage(session *middleware.SSESession, botChatMessage *models.ChatMessage) {
	// 봇 응답을 SSE로 전송
	botSSEMessage := BotMessage{
		Type:      "bot",
		Content:   botChatMessage.Content,
		Timestamp: botChatMessage.CreatedAt.Format(time.RFC3339),
		SessionID: session.SessionID,
	}

	botSSEData, _ := json.Marshal(botSSEMessage)
	botMessageData := fmt.Sprintf("data: %s\n\n", string(botSSEData))

	// session.Channel로 전송 (클라이언트에 전달)
	select {
	case session.Channel <- botMessageData:
		log.Printf("봇 응답 전송 완료 - 세션: %s", session.SessionID)
	default:
		log.Printf("세션 채널이 가득 참 - 세션: %s", session.SessionID)
	}
}

// extractUserStatus 사용자 메시지에서 상태 정보 추출
func (bg *BotGoroutine) extractUserStatus(userMessage string, openaiClient *openai.Client) *status.StatusExtractionResult {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 상태 정보 추출을 위한 프롬프트 구성
	statusPrompt := prompt.GetStatusExtractionPrompt()
	statusMessages := []openai.ChatMessage{
		{
			Role:    "system",
			Content: statusPrompt,
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}

	log.Printf("상태 정보 추출 요청 시작 - 사용자 메시지: %s", userMessage)

	// 상태 정보 추출 요청
	statusResponse, err := openaiClient.ChatCompletion(ctx, statusMessages)
	if err != nil {
		log.Printf("상태 정보 추출 실패: %v", err)
		return nil
	}

	log.Printf("상태 정보 추출 응답 수신 - 응답: %s", statusResponse.Message.Content)

	// 상태 정보 파싱
	statusResult, err := bg.statusService.ParseStatusExtractionResult(statusResponse.Message.Content)
	if err != nil {
		log.Printf("상태 정보 파싱 실패: %v, 원본 응답: %s", err, statusResponse.Message.Content)
		return nil
	}

	log.Printf("상태 정보 파싱 완료 - needs_save: %t, event: %s, valid_until: %s",
		statusResult.NeedsSave, statusResult.Event, statusResult.ValidUntil)

	return statusResult
}

// saveUserStatus 사용자 상태 정보 저장
func (bg *BotGoroutine) saveUserStatus(session *middleware.SSESession, event string, validUntil time.Time, context string) {
	err := bg.statusService.SaveUserStatus(
		session.UserUUID,
		session.ChatbotUUID,
		event,
		validUntil,
		context,
	)
	if err != nil {
		log.Printf("상태 정보 저장 실패 - 세션: %s, 에러: %v", session.SessionID, err)
	} else {
		log.Printf("상태 정보 저장 완료 - 세션: %s, 이벤트: %s", session.SessionID, event)
	}
}

// handleSessionDone 세션 종료 처리
func (bg *BotGoroutine) handleSessionDone(session *middleware.SSESession, responseTimer *time.Timer) {
	log.Printf("봇 고루틴 종료 신호 수신 - 세션: %s", session.SessionID)
	if responseTimer != nil {
		responseTimer.Stop()
	}
}

// 전역 BotGoroutine 인스턴스
var globalBotGoroutine = NewBotGoroutine()

// GetBotGoroutine 전역 BotGoroutine 반환
func GetBotGoroutine() *BotGoroutine {
	return globalBotGoroutine
}
