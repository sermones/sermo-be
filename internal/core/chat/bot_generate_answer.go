package chat

import (
	"context"
	"log"

	"sermo-be/internal/middleware"
	"sermo-be/internal/models"
	"sermo-be/pkg/openai"
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

// GenerateAnswer AI 응답 생성 및 저장
func (ag *AnswerGenerator) GenerateAnswer(session *middleware.SSESession, combinedMessage string, openaiClient *openai.Client) *models.ChatMessage {
	// 이전 대화 내용 가져오기 (최근 20개)
	history, err := ag.messageService.GetChatHistory(session.UserUUID, session.ChatbotUUID, 20)
	if err != nil {
		log.Printf("대화 히스토리 조회 실패 - 세션: %s, 에러: %v", session.SessionID, err)
		return nil
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

	// AI 응답 생성
	response, err := openaiClient.ChatCompletion(context.Background(), messages)
	if err != nil {
		log.Printf("AI 응답 생성 실패 - 세션: %s, 에러: %v", session.SessionID, err)
		return nil
	}

	log.Printf("AI 응답 생성 완료 - 세션: %s, 내용 길이: %d", session.SessionID, len(response.Message.Content))

	// 봇 메시지 저장
	botChatMessage, err := ag.messageService.CreateBotMessage(
		session.SessionID,
		session.UserUUID,
		session.ChatbotUUID,
		response.Message.Content,
	)

	if err != nil {
		log.Printf("봇 메시지 저장 실패 - 세션: %s, 에러: %v", session.SessionID, err)
		return nil
	}

	return botChatMessage
}
