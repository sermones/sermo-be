package chat

import (
	"fmt"

	"sermo-be/internal/models"
	"sermo-be/pkg/database"
)

// MessageService 채팅 메시지 관련 비즈니스 로직을 담당하는 서비스
type MessageService struct{}

// NewMessageService 새로운 MessageService 인스턴스 생성
func NewMessageService() *MessageService {
	return &MessageService{}
}

// CreateUserMessage 사용자 메시지를 생성하고 DB에 저장
func (s *MessageService) CreateUserMessage(sessionID, userUUID, chatbotUUID, userMessage string) (*models.ChatMessage, error) {
	// 사용자 메시지 생성 및 저장
	userChatMessage := models.NewChatMessage(
		sessionID,
		userUUID,
		chatbotUUID,
		models.MessageTypeUser,
		userMessage,
	)

	if err := database.DB.Create(userChatMessage).Error; err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	return userChatMessage, nil
}

// 전역 MessageService 인스턴스
var globalMessageService = NewMessageService()

// GetMessageService 전역 MessageService 반환
func GetMessageService() *MessageService {
	return globalMessageService
}
