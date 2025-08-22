package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UUID      uuid.UUID `json:"uuid" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ID        string    `json:"id" gorm:"type:varchar(20);uniqueIndex;not null"`
	Nickname  string    `json:"nickname" gorm:"type:varchar(100);not null"`
	Password  string    `json:"-" gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func NewUser(id, nickname, password string) *User {
	now := time.Now()
	return &User{
		UUID:      uuid.New(),
		ID:        id,
		Nickname:  nickname,
		Password:  password,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
