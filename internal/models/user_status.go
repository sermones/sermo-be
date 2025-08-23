package models

import (
	"time"

	"github.com/google/uuid"
)

// UserStatus 사용자 상태 정보 모델
type UserStatus struct {
	UUID        uuid.UUID `json:"uuid" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserUUID    string    `json:"user_uuid" gorm:"type:varchar(36);not null;index"`
	ChatbotUUID string    `json:"chatbot_uuid" gorm:"type:varchar(36);not null;index"`
	Event       string    `json:"event" gorm:"type:varchar(255);not null"`           // 이벤트 (예: 시험보기, 회의, 약속 등)
	ValidUntil  time.Time `json:"valid_until" gorm:"type:timestamp;not null"`       // 유효한 시간
	Context     string    `json:"context" gorm:"type:text"`                          // 추가 컨텍스트 정보
	IsActive    bool      `json:"is_active" gorm:"type:boolean;default:true"`        // 활성 상태 여부
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// NewUserStatus 새로운 사용자 상태 정보 인스턴스 생성
func NewUserStatus(userUUID, chatbotUUID, event string, validUntil time.Time, context string) *UserStatus {
	now := time.Now()
	return &UserStatus{
		UUID:        uuid.New(),
		UserUUID:    userUUID,
		ChatbotUUID: chatbotUUID,
		Event:       event,
		ValidUntil:  validUntil,
		Context:     context,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// TableName GORM 테이블명 지정
func (UserStatus) TableName() string {
	return "user_statuses"
}
