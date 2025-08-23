package models

import (
	"time"

	"github.com/google/uuid"
)

type SentenceBookmark struct {
	UUID      uuid.UUID `json:"uuid" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserUUID  uuid.UUID `json:"user_uuid" gorm:"type:uuid;not null;index"`
	Sentence  string    `json:"sentence" gorm:"type:text;not null"`
	Meaning   string    `json:"meaning" gorm:"type:text;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func NewSentenceBookmark(userUUID uuid.UUID, sentence string, meaning string) *SentenceBookmark {
	now := time.Now()
	return &SentenceBookmark{
		UUID:      uuid.New(),
		UserUUID:  userUUID,
		Sentence:  sentence,
		Meaning:   meaning,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type WordBookmark struct {
	UUID      uuid.UUID `json:"uuid" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserUUID  uuid.UUID `json:"user_uuid" gorm:"type:uuid;not null;index"`
	Word      string    `json:"word" gorm:"type:varchar(100);not null"`
	Meaning   string    `json:"meaning" gorm:"type:text;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func NewWordBookmark(userUUID uuid.UUID, word string, meaning string) *WordBookmark {
	now := time.Now()
	return &WordBookmark{
		UUID:      uuid.New(),
		UserUUID:  userUUID,
		Word:      word,
		Meaning:   meaning,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
