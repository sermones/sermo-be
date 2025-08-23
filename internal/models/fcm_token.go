package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FCMToken struct {
	ID         uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	UserUUID   *uuid.UUID     `json:"user_uuid" gorm:"type:uuid;index;default:null"`
	FCMToken   string         `json:"fcm_token" gorm:"type:varchar(255);not null;uniqueIndex"`
	DeviceInfo string         `json:"device_info" gorm:"type:varchar(100);default:''"`
	CreatedAt  time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func NewFCMToken(fcmToken, deviceInfo string) *FCMToken {
	now := time.Now()
	return &FCMToken{
		FCMToken:   fcmToken,
		DeviceInfo: deviceInfo,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func NewFCMTokenWithUser(userUUID uuid.UUID, fcmToken, deviceInfo string) *FCMToken {
	now := time.Now()
	return &FCMToken{
		UserUUID:   &userUUID,
		FCMToken:   fcmToken,
		DeviceInfo: deviceInfo,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
