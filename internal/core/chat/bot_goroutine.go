package chat

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"sermo-be/internal/middleware"
	"sermo-be/internal/models"
	"sermo-be/pkg/openai"
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
	messageService  *MessageService
	answerGenerator *AnswerGenerator
	statusGenerator *StatusGenerator
	openaiClient    *openai.Client
}

// NewBotGoroutine 새로운 봇 고루틴 관리자 생성
func NewBotGoroutine() *BotGoroutine {
	return &BotGoroutine{
		messageService:  GetMessageService(),
		answerGenerator: NewAnswerGenerator(),
		statusGenerator: NewStatusGenerator(),
		openaiClient:    nil, // StartBotGoroutine에서 설정됨
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

	log.Printf("봇 고루틴 시작 - 세션: %s", session.SessionID)

	// 봇 채널에서 메시지 수신
	for session.IsActive {
		select {
		case message := <-botChannel:
			log.Printf("봇 채널에서 메시지 수신 - 세션: %s, 메시지: %s", session.SessionID, message)
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
	log.Printf("handleIncomingMessage 시작 - 세션: %s", session.SessionID)

	// 메시지 파싱
	sseMessage, err := bg.parseUserMessage(message)
	if err != nil {
		log.Printf("사용자 메시지 파싱 실패: %v", err)
		return
	}

	log.Printf("메시지 파싱 성공 - 타입: %s, 내용: %s - 세션: %s", sseMessage.Type, sseMessage.Content, session.SessionID)

	// 메시지 타입별로 처리
	switch sseMessage.Type {
	case "user":
		log.Printf("사용자 메시지 처리 시작 - 세션: %s", session.SessionID)
		bg.processUserMessage(session, sseMessage, messageBuffer, responseTimer, openaiClient)
	case "onkeyboard":
		log.Printf("onkeyboard 이벤트 처리 시작 - 세션: %s", session.SessionID)
		bg.processOnKeyboardEvent(session, responseTimer)
	case "bot_typing":
		// 타이핑 이벤트는 봇 채널로 전달하지 않음 (클라이언트에만 전송)
		log.Printf("봇 타이핑 이벤트 수신 - 봇 채널로 전달하지 않음 - 세션: %s", session.SessionID)
	default:
		log.Printf("알 수 없는 메시지 타입: %s", sseMessage.Type)
	}

	log.Printf("handleIncomingMessage 완료 - 세션: %s", session.SessionID)
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
	log.Printf("processUserMessage 시작 - 세션: %s, 메시지: %s", session.SessionID, sseMessage.Content)

	// 기존 타이머가 있다면 취소
	if *responseTimer != nil {
		log.Printf("기존 타이머 취소 - 세션: %s", session.SessionID)
		(*responseTimer).Stop()
	}

	// 메시지를 버퍼에 추가
	*messageBuffer = append(*messageBuffer, sseMessage.Content)
	log.Printf("메시지 버퍼에 추가됨 - 버퍼 크기: %d, 세션: %s", len(*messageBuffer), session.SessionID)

	// 4초 후 버퍼에 쌓인 모든 메시지로 봇 응답 생성 (버퍼링 구현)
	*responseTimer = time.AfterFunc(4*time.Second, func() {
		log.Printf("타이머 만료 - AI 응답 생성 시작 - 세션: %s", session.SessionID)
		bg.generateAIResponse(session, *messageBuffer, openaiClient)
		// 응답 생성 후 버퍼 초기화
		*messageBuffer = (*messageBuffer)[:0]
		log.Printf("메시지 버퍼 초기화 완료 - 세션: %s", session.SessionID)
	})

	log.Printf("processUserMessage 완료 - 타이머 설정됨 (4초 후 실행) - 세션: %s", session.SessionID)
}

// processOnKeyboardEvent 키보드 입력 이벤트 처리
func (bg *BotGoroutine) processOnKeyboardEvent(session *middleware.SSESession, responseTimer **time.Timer) {
	// 기존 타이머가 있다면 취소
	// if *responseTimer != nil {
	// 	(*responseTimer).Stop()
	// }

	// // 키보드 입력 중일 때는 최소 5초 대기
	// *responseTimer = time.AfterFunc(5*time.Second, func() {
	// 	// 5초 후에 타이머를 nil로 설정하여 추가 대기 없이 즉시 응답 가능하게 함
	// 	*responseTimer = nil
	// })
}

// generateAIResponse AI 응답 생성
func (bg *BotGoroutine) generateAIResponse(session *middleware.SSESession, messageBuffer []string, openaiClient *openai.Client) {
	log.Printf("generateAIResponse 시작 - 세션: %s, 버퍼 크기: %d", session.SessionID, len(messageBuffer))

	if !session.IsActive {
		log.Printf("세션이 비활성 상태 - AI 응답 생성 중단 - 세션: %s", session.SessionID)
		return
	}

	// 버퍼에 메시지가 없으면 리턴
	if len(messageBuffer) == 0 {
		log.Printf("메시지 버퍼가 비어있음 - AI 응답 생성 중단 - 세션: %s", session.SessionID)
		return
	}

	// 버퍼의 모든 메시지를 하나의 컨텍스트로 결합
	combinedMessage := bg.combineMessages(messageBuffer)
	log.Printf("메시지 결합 완료 - 결합된 메시지: %s - 세션: %s", combinedMessage, session.SessionID)

	// 동시에 두 개의 고루틴으로 처리
	responseChan := make(chan *models.ChatMessage, 1)
	statusChan := make(chan bool, 1)

	log.Printf("AI 응답 생성 고루틴 시작 - 세션: %s", session.SessionID)
	// 고루틴 1: AI 답변 생성
	go func() {
		log.Printf("AI 답변 생성 시작 - 세션: %s", session.SessionID)
		botChatMessage := bg.answerGenerator.GenerateAnswer(session, combinedMessage, openaiClient)
		log.Printf("AI 답변 생성 완료 - 세션: %s, 응답: %v", session.SessionID, botChatMessage)
		responseChan <- botChatMessage
	}()

	log.Printf("상태 정보 추출 고루틴 시작 - 세션: %s", session.SessionID)
	// 고루틴 2: 상태 정보 추출 및 저장
	go func() {
		bg.statusGenerator.ExtractAndSaveStatus(session, combinedMessage, openaiClient)
		statusChan <- true
	}()

	// AI 응답 대기
	log.Printf("AI 응답 대기 중 - 세션: %s", session.SessionID)
	botChatMessage := <-responseChan
	if botChatMessage == nil {
		log.Printf("AI 응답 생성 실패 - 세션: %s", session.SessionID)
		return
	}

	log.Printf("AI 응답 생성 성공 - 세션: %s, 응답 내용: %s", session.SessionID, botChatMessage.Content)

	// 봇 응답을 session.Channel로 전송
	bg.sendBotMessage(session, botChatMessage)

	// 상태 정보 처리 완료 대기
	<-statusChan
	log.Printf("generateAIResponse 완료 - 세션: %s", session.SessionID)
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
	log.Printf("sendBotMessage 시작 - 세션: %s, 응답 내용: %s", session.SessionID, botChatMessage.Content)

	// 봇 응답을 SSE로 전송
	botSSEMessage := BotMessage{
		Type:      "bot",
		Content:   botChatMessage.Content,
		Timestamp: botChatMessage.CreatedAt.Format(time.RFC3339),
		SessionID: session.SessionID,
	}

	botSSEData, _ := json.Marshal(botSSEMessage)
	botMessageData := fmt.Sprintf("data: %s\n\n", string(botSSEData))
	log.Printf("봇 SSE 메시지 생성 완료 - 세션: %s, SSE 메시지: %s", session.SessionID, botMessageData)

	// session.Channel로 전송 (클라이언트에 전달)
	select {
	case session.Channel <- botMessageData:
		log.Printf("봇 응답 전송 성공 - 세션: %s", session.SessionID)
	default:
		log.Printf("세션 채널이 가득 참 - 세션: %s", session.SessionID)
	}

	log.Printf("sendBotMessage 완료 - 세션: %s", session.SessionID)
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
