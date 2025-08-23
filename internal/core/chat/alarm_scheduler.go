package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"sermo-be/internal/models"
	"sermo-be/pkg/database"
	"sermo-be/pkg/firebase"
)

// AlarmScheduler 알람 스케줄링을 담당하는 서비스
type AlarmScheduler struct {
	firebaseClient *firebase.Client
	stopChan       chan struct{}
}

// NewAlarmScheduler 새로운 알람 스케줄러 생성
func NewAlarmScheduler(firebaseClient *firebase.Client) *AlarmScheduler {
	return &AlarmScheduler{
		firebaseClient: firebaseClient,
		stopChan:       make(chan struct{}),
	}
}

// Start 알람 스케줄러 시작
func (as *AlarmScheduler) Start() {
	log.Println("🔔 알람 스케줄러 시작 (5초마다 실행)")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// 즉시 한 번 실행
	as.processScheduledAlarms()

	for {
		select {
		case <-ticker.C:
			as.processScheduledAlarms()
		case <-as.stopChan:
			log.Println("🛑 알람 스케줄러 종료")
			return
		}
	}
}

// Stop 알람 스케줄러 종료
func (as *AlarmScheduler) Stop() {
	close(as.stopChan)
}

// processScheduledAlarms 예약된 알람들을 처리
func (as *AlarmScheduler) processScheduledAlarms() {
	ctx := context.Background()

	// 데이터베이스에서 전송할 알람 조회
	alarms, err := as.getAlarmsToSend(ctx)
	if err != nil {
		log.Printf("❌ 전송할 알람 조회 실패: %v", err)
		return
	}

	if len(alarms) == 0 {
		return // 전송할 알람이 없음
	}

	log.Printf("📱 %d개의 알람을 전송합니다", len(alarms))

	// 각 알람을 FCM으로 전송
	for _, alarm := range alarms {
		if err := as.sendAlarmToFCM(ctx, alarm); err != nil {
			log.Printf("❌ FCM 전송 실패 - 사용자: %s, 에러: %v", alarm.UserUUID, err)
			continue
		}

		// 전송 완료 후 알람을 전송 완료 상태로 표시
		if err := as.markAlarmAsSended(ctx, alarm); err != nil {
			log.Printf("⚠️ 알람 상태 업데이트 실패: %v", err)
		}

		log.Printf("✅ FCM 알람 전송 완료 - 사용자: %s, 메시지: %s", alarm.UserUUID, alarm.Message)
	}
}

// getAlarmsToSend 전송할 알람들을 데이터베이스에서 조회
func (as *AlarmScheduler) getAlarmsToSend(ctx context.Context) ([]models.AlarmSchedule, error) {
	var alarms []models.AlarmSchedule

	now := time.Now()

	// 데이터베이스에서 전송할 알람 조회 (전송 시간이 되었고 아직 전송되지 않은 알람)
	err := database.DB.Where("send_time <= ? AND sended = ?", now, false).Find(&alarms).Error
	if err != nil {
		return alarms, fmt.Errorf("데이터베이스 알람 조회 실패: %w", err)
	}

	return alarms, nil
}

// sendAlarmToFCM 개별 알람을 FCM으로 전송
func (as *AlarmScheduler) sendAlarmToFCM(ctx context.Context, alarm models.AlarmSchedule) error {
	// 사용자의 FCM 토큰 조회
	fcmTokens, err := as.getUserFCMTokens(alarm.UserUUID)
	if err != nil {
		return fmt.Errorf("FCM 토큰 조회 실패: %w", err)
	}

	if len(fcmTokens) == 0 {
		return fmt.Errorf("사용자의 FCM 토큰이 없음")
	}

	// FCM 메시지 생성
	message := as.createFCMMessage(alarm)

	// 각 FCM 토큰으로 전송
	for _, token := range fcmTokens {
		if err := as.sendSingleFCM(ctx, message, token); err != nil {
			log.Printf("⚠️ 개별 FCM 전송 실패 - 토큰: %s, 에러: %v", token, err)
			continue
		}
	}

	return nil
}

// getUserFCMTokens 사용자의 FCM 토큰들 조회
func (as *AlarmScheduler) getUserFCMTokens(userUUID string) ([]string, error) {
	// 데이터베이스에서 FCM 토큰 조회
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
func (as *AlarmScheduler) createFCMMessage(alarm models.AlarmSchedule) *firebase.ChatNotification {
	// keywords를 문자열 배열로 변환 (JSON에서 파싱)
	var keywords []string
	if alarm.Keywords != nil {
		if err := json.Unmarshal(alarm.Keywords, &keywords); err != nil {
			log.Printf("⚠️ keywords JSON 파싱 실패: %v", err)
			keywords = []string{}
		}
	}

	return firebase.NewChatNotification(
		alarm.ChatbotName,
		alarm.ChatbotAvatar,
		alarm.ChatbotUUID,
		alarm.Message,
		time.Now().Unix(),
	)
}

// sendSingleFCM 단일 FCM 메시지 전송
func (as *AlarmScheduler) sendSingleFCM(ctx context.Context, message *firebase.ChatNotification, token string) error {
	// FCM 메시지로 변환
	fcmMessage := message.ToFCMMessage(token)

	// Firebase Messaging 클라이언트 가져오기
	messagingClient := as.firebaseClient.GetMessagingClient()

	// FCM 전송
	_, err := messagingClient.Send(ctx, fcmMessage)
	return err
}

// markAlarmAsSended 알람을 전송 완료 상태로 표시
func (as *AlarmScheduler) markAlarmAsSended(ctx context.Context, alarm models.AlarmSchedule) error {
	return database.DB.Model(&alarm).Update("sended", true).Error
}
