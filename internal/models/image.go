package models

import (
	"time"

	"github.com/google/uuid"
)

// Image 사용자 이미지 정보
type Image struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    string    `json:"user_id" gorm:"type:varchar(255);not null;index"`
	FileName  string    `json:"file_name" gorm:"type:varchar(255);not null"`
	FileSize  int64     `json:"file_size" gorm:"not null"`
	MimeType  string    `json:"mime_type" gorm:"type:varchar(100);not null"`
	URL       string    `json:"url" gorm:"type:text;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func NewImage(userID, fileName, mimeType, url string, fileSize int64) *Image {
	now := time.Now()
	return &Image{
		ID:        uuid.New(),
		UserID:    userID,
		FileName:  fileName,
		FileSize:  fileSize,
		MimeType:  mimeType,
		URL:       url,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
