package chat

import (
	"context"
	"fmt"
	"log"
	"time"

	"sermo-be/internal/models"
	"sermo-be/pkg/database"
	"sermo-be/pkg/firebase"
)

// FCMSender FCM 전송을 담당하는 구조체
type FCMSender struct {
	firebaseClient *firebase.Client
}

// NewFCMSender 새로운 FCMSender 생성
func NewFCMSender(firebaseClient *firebase.Client) *FCMSender {
	return &FCMSender{
		firebaseClient: firebaseClient,
	}
}

// SendImmediateFCM 즉시 FCM 전송 (테스트용)
func (fs *FCMSender) SendImmediateFCM(userUUID, message string) error {
	ctx := context.Background()

	// 사용자의 FCM 토큰 조회
	fcmTokens, err := fs.getUserFCMTokens(userUUID)
	if err != nil {
		return fmt.Errorf("FCM 토큰 조회 실패: %w", err)
	}

	if len(fcmTokens) == 0 {
		return fmt.Errorf("사용자의 FCM 토큰이 없음")
	}

	// FCM 메시지 생성
	fcmMessage := fs.createSimpleFCMMessage(message)

	// 각 FCM 토큰으로 전송
	for _, token := range fcmTokens {
		if err := fs.sendSingleFCM(ctx, fcmMessage, token); err != nil {
			log.Printf("⚠️ 개별 FCM 전송 실패 - 토큰: %s, 에러: %v", token, err)
			continue
		}
	}

	return nil
}

// SendImmediateFCMNotification 즉시 FCM 알람 전송 (패키지 레벨 함수)
func SendImmediateFCMNotification(userUUID, message, chatbotUUID string) error {
	// Firebase 클라이언트 생성 (임시로 빈 설정 사용)
	// TODO: 설정을 전달받아서 생성하도록 수정 필요
	firebaseClient, err := firebase.NewClient(nil)
	if err != nil {
		return fmt.Errorf("Firebase 클라이언트 생성 실패: %w", err)
	}
	defer firebaseClient.Close()

	// FCMSender 생성
	fcmSender := NewFCMSender(firebaseClient)

	// FCM 전송
	return fcmSender.SendImmediateFCM(userUUID, message)
}

// getUserFCMTokens 사용자의 FCM 토큰들 조회
func (fs *FCMSender) getUserFCMTokens(userUUID string) ([]string, error) {
	var fcmTokens []models.FCMToken
	if err := database.DB.Where("user_uuid = ?", userUUID).Find(&fcmTokens).Error; err != nil {
		return nil, fmt.Errorf("FCM 토큰 조회 실패: %w", err)
	}

	var tokens []string
	for _, fcmToken := range fcmTokens {
		if fcmToken.FCMToken != "" {
			tokens = append(tokens, fcmToken.FCMToken)
		}
	}

	return tokens, nil
}

// createSimpleFCMMessage 간단한 FCM 메시지 생성
func (fs *FCMSender) createSimpleFCMMessage(message string) *firebase.ChatNotification {
	return firebase.NewChatNotification(
		"Sermo",
		"",
		"",
		message,
		time.Now().Unix(),
	)
}

// sendSingleFCM 단일 FCM 메시지 전송
func (fs *FCMSender) sendSingleFCM(ctx context.Context, message *firebase.ChatNotification, token string) error {
	// FCM 메시지로 변환
	fcmMessage := message.ToFCMMessage(token)

	// Firebase Messaging 클라이언트 가져오기
	messagingClient := fs.firebaseClient.GetMessagingClient()

	// FCM 전송
	_, err := messagingClient.Send(ctx, fcmMessage)
	return err
}
