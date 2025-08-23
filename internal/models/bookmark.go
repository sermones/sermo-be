package models

import (
	"time"

	"github.com/google/uuid"
)

type SentenceBookmark struct {
	UUID      uuid.UUID `json:"uuid" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserUUID  uuid.UUID `json:"user_uuid" gorm:"type:uuid;not null;index"`
	Sentence  string    `json:"sentence" gorm:"type:text;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func NewSentenceBookmark(userUUID uuid.UUID, sentence string) *SentenceBookmark {
	now := time.Now()
	return &SentenceBookmark{
		UUID:      uuid.New(),
		UserUUID:  userUUID,
		Sentence:  sentence,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
