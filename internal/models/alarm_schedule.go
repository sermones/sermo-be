package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// AlarmSchedule 알람 스케줄 모델
type AlarmSchedule struct {
	ID            uint            `json:"id" gorm:"primaryKey"`
	UserUUID      string          `json:"user_uuid" gorm:"not null;index"`
	ChatbotUUID   string          `json:"chatbot_uuid" gorm:"not null;index"`
	ChatbotName   string          `json:"chatbot_name" gorm:"not null"`
	ChatbotAvatar string          `json:"chatbot_avatar"`
	Message       string          `json:"message" gorm:"not null"`
	SendTime      time.Time       `json:"send_time" gorm:"not null;index"`
	Keywords      json.RawMessage `json:"keywords" gorm:"type:json"`
	Context       string          `json:"context"`
	Sended        bool            `json:"sended" gorm:"default:false;index"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `json:"deleted_at" gorm:"index"`
}

// TableName 테이블명 지정
func (AlarmSchedule) TableName() string {
	return "alarm_schedules"
}
