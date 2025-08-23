package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Gender 성별 enum
type Gender string

const (
	GenderMale        Gender = "male"        // 남성
	GenderFemale      Gender = "female"      // 여성
	GenderUnspecified Gender = "unspecified" // 지정안함
)

// Chatbot 채팅봇 모델
type Chatbot struct {
	UUID      uuid.UUID       `json:"uuid" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string          `json:"name" gorm:"type:varchar(100);not null"`
	ImageID   string          `json:"image_id" gorm:"type:varchar(255);not null"`
	Hashtags  json.RawMessage `json:"hashtags" gorm:"type:jsonb"`
	Gender    Gender          `json:"gender" gorm:"type:varchar(20);not null;default:'unspecified'"`
	Details   string          `json:"details" gorm:"type:text"`
	Summary   *string         `json:"summary" gorm:"type:text"`                   // AI가 생성한 캐릭터 요약 (optional)
	UserUUID  string          `json:"user_uuid" gorm:"type:varchar(36);not null"` // FK 없이 문자열로 저장
	CreatedAt time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

// NewChatbot 새로운 채팅봇 인스턴스 생성
func NewChatbot(name, imageID string, hashtags json.RawMessage, gender Gender, details, userUUID string) *Chatbot {
	now := time.Now()
	return &Chatbot{
		UUID:      uuid.New(),
		Name:      name,
		ImageID:   imageID,
		Hashtags:  hashtags,
		Gender:    gender,
		Details:   details,
		Summary:   nil, // 초기에는 요약 없음
		UserUUID:  userUUID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// TableName GORM 테이블명 지정
func (Chatbot) TableName() string {
	return "chatbots"
}

// GetSummary 기존 요약 반환 (없으면 nil)
func (c *Chatbot) GetSummary() *string {
	return c.Summary

}

// SetSummary 요약 정보 설정
func (c *Chatbot) SetSummary(summary string) {
	c.Summary = &summary
}
