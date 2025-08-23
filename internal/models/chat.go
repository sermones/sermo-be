package models

import (
	"time"

	"github.com/google/uuid"
)

// MessageType 메시지 타입 enum
type MessageType string

const (
	MessageTypeUser    MessageType = "user"    // 사용자 메시지
	MessageTypeChatbot MessageType = "chatbot" // 채팅봇 메시지
)

// ChatMessage 채팅 메시지 모델
type ChatMessage struct {
	UUID        uuid.UUID   `json:"uuid" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SessionID   string      `json:"session_id" gorm:"type:varchar(36);not null;index"`
	UserUUID    string      `json:"user_uuid" gorm:"type:varchar(36);not null;index"`
	ChatbotUUID string      `json:"chatbot_uuid" gorm:"type:varchar(36);not null;index"`
	MessageType MessageType `json:"message_type" gorm:"type:varchar(20);not null"`
	Content     string      `json:"content" gorm:"type:text;not null"`
	CreatedAt   time.Time   `json:"created_at" gorm:"autoCreateTime"`
}

// NewChatMessage 새로운 채팅 메시지 인스턴스 생성
func NewChatMessage(sessionID, userUUID, chatbotUUID string, messageType MessageType, content string) *ChatMessage {
	return &ChatMessage{
		UUID:        uuid.New(),
		SessionID:   sessionID,
		UserUUID:    userUUID,
		ChatbotUUID: chatbotUUID,
		MessageType: messageType,
		Content:     content,
	}
}

// TableName GORM 테이블명 지정
func (ChatMessage) TableName() string {
	return "chat_messages"
}
