package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"sermo-be/internal/models"
	"sermo-be/pkg/database"
	"sermo-be/pkg/firebase"
	"sermo-be/pkg/redis"
)

// FCMSender FCM 전송을 담당하는 구조체
type FCMSender struct{}

// NewFCMSender 새로운 FCMSender 생성
func NewFCMSender() *FCMSender {
	return &FCMSender{}
}

// SendScheduledAlarms 예약된 알람을 FCM으로 전송
func (fs *FCMSender) SendScheduledAlarms() error {
	ctx := context.Background()

	// Redis에서 전송할 알람 조회
	alarms, err := fs.getAlarmsToSend(ctx)
	if err != nil {
		return fmt.Errorf("전송할 알람 조회 실패: %w", err)
	}

	if len(alarms) == 0 {
		return nil // 전송할 알람이 없음
	}

	// 각 알람을 FCM으로 전송
	for _, alarm := range alarms {
		if err := fs.sendAlarmToFCM(ctx, alarm); err != nil {
			log.Printf("FCM 전송 실패 - 사용자: %s, 에러: %v", alarm.UserUUID, err)
			continue
		}

		// 전송 완료 후 Redis에서 제거
		key := fmt.Sprintf("alarm:%s:%d", alarm.UserUUID, alarm.ScheduledAt.Unix())
		if err := redis.DeleteKey(ctx, key); err != nil {
			log.Printf("Redis 알람 삭제 실패: %v", err)
		}

		log.Printf("FCM 알람 전송 완료 - 사용자: %s, 메시지: %s", alarm.UserUUID, alarm.Message)
	}

	return nil
}

// getAlarmsToSend 전송할 알람들을 Redis에서 조회
func (fs *FCMSender) getAlarmsToSend(ctx context.Context) ([]ScheduledAlarm, error) {
	var alarms []ScheduledAlarm

	// Redis에서 모든 알람 키 조회
	keys, err := redis.Keys(ctx, "alarm:*")
	if err != nil {
		return alarms, fmt.Errorf("Redis 키 조회 실패: %w", err)
	}

	now := time.Now()

	for _, key := range keys {
		// 키에서 timestamp 추출
		parts := strings.Split(key, ":")
		if len(parts) != 3 {
			continue
		}

		// 알람 데이터 조회
		alarmData, err := redis.GetKey(ctx, key)
		if err != nil {
			log.Printf("알람 데이터 조회 실패 - 키: %s, 에러: %v", key, err)
			continue
		}

		var alarm ScheduledAlarm
		if err := json.Unmarshal([]byte(alarmData), &alarm); err != nil {
			log.Printf("알람 데이터 파싱 실패 - 키: %s, 에러: %v", key, err)
			continue
		}

		// 전송 시간이 된 알람만 필터링
		if alarm.ScheduledAt.Before(now) || alarm.ScheduledAt.Equal(now) {
			alarms = append(alarms, alarm)
		}
	}

	return alarms, nil
}

// sendAlarmToFCM 개별 알람을 FCM으로 전송
func (fs *FCMSender) sendAlarmToFCM(ctx context.Context, alarm ScheduledAlarm) error {
	// 사용자의 FCM 토큰 조회
	fcmTokens, err := fs.getUserFCMTokens(alarm.UserUUID)
	if err != nil {
		return fmt.Errorf("FCM 토큰 조회 실패: %w", err)
	}

	if len(fcmTokens) == 0 {
		return fmt.Errorf("사용자의 FCM 토큰이 없음")
	}

	// FCM 메시지 생성
	message := fs.createFCMMessage(alarm)

	// 각 FCM 토큰으로 전송
	for _, token := range fcmTokens {
		if err := fs.sendSingleFCM(message, token); err != nil {
			log.Printf("개별 FCM 전송 실패 - 토큰: %s, 에러: %v", token, err)
			continue
		}
	}

	return nil
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

// createFCMMessage FCM 메시지 생성
func (fs *FCMSender) createFCMMessage(alarm ScheduledAlarm) *firebase.ChatNotification {
	return firebase.NewChatNotification(
		alarm.ChatbotName,
		alarm.ChatbotAvatar,
		alarm.ChatbotUUID,
		alarm.Message,
		time.Now().Unix(),
	)
}

// sendSingleFCM 단일 FCM 메시지 전송
func (fs *FCMSender) sendSingleFCM(message *firebase.ChatNotification, token string) error {
	// FCM 메시지로 변환
	fcmMessage := message.ToFCMMessage(token)

	// Firebase 클라이언트 생성 (임시로 빈 설정 사용)
	// TODO: 설정을 전달받아서 생성하도록 수정 필요
	firebaseClient, err := firebase.NewClient(nil)
	if err != nil {
		return fmt.Errorf("Firebase 클라이언트 생성 실패: %w", err)
	}
	defer firebaseClient.Close()

	// Firebase Messaging 클라이언트 가져오기
	messagingClient := firebaseClient.GetMessagingClient()

	// FCM 전송
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = messagingClient.Send(ctx, fcmMessage)
	return err
}
