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

// GetChatHistory 사용자와 채팅봇의 대화 히스토리 조회
func (s *MessageService) GetChatHistory(userUUID, chatbotUUID string, limit int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage

	query := database.DB.Where("user_uuid = ? AND chatbot_uuid = ?", userUUID, chatbotUUID)

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch chat history: %w", err)
	}

	return messages, nil
}

// CreateBotMessage 봇 메시지 생성 및 저장
func (s *MessageService) CreateBotMessage(sessionID, userUUID, chatbotUUID, content string) (*models.ChatMessage, error) {
	botMessage := models.NewChatMessage(
		sessionID,
		userUUID,
		chatbotUUID,
		models.MessageTypeChatbot,
		content,
	)

	if err := database.DB.Create(botMessage).Error; err != nil {
		return nil, fmt.Errorf("failed to save bot message: %w", err)
	}

	return botMessage, nil
}

// GetChatHistoryCount 사용자와 채팅봇의 총 메시지 수 조회
func (s *MessageService) GetChatHistoryCount(userUUID, chatbotUUID string) (int64, error) {
	var count int64

	if err := database.DB.Model(&models.ChatMessage{}).
		Where("user_uuid = ? AND chatbot_uuid = ?", userUUID, chatbotUUID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return count, nil
}

// 전역 MessageService 인스턴스
var globalMessageService = NewMessageService()

// GetMessageService 전역 MessageService 반환
func GetMessageService() *MessageService {
	return globalMessageService
}
