package chat

import (
	"log"
	"time"

	"sermo-be/pkg/database"
	"sermo-be/pkg/openai"
)

// IntegratedScheduler 알람 예약과 FCM 전송을 통합 관리하는 스케줄러
type IntegratedScheduler struct {
	alarmScheduler *AlarmScheduler
	fcmSender      *FCMSender
	openaiClient   *openai.Client
}

// NewIntegratedScheduler 새로운 IntegratedScheduler 생성
func NewIntegratedScheduler(openaiClient *openai.Client) *IntegratedScheduler {
	return &IntegratedScheduler{
		alarmScheduler: NewAlarmScheduler(openaiClient),
		fcmSender:      NewFCMSender(),
		openaiClient:   openaiClient,
	}
}

// StartSchedulers 모든 스케줄러 시작
func (is *IntegratedScheduler) StartSchedulers() {
	// 1시간마다 알람 예약 스케줄러
	go is.startAlarmScheduler()

	// 1분마다 FCM 전송 스케줄러
	go is.startFCMDeliveryScheduler()

	log.Println("✅ 통합 스케줄러 시작 완료")
}

// startAlarmScheduler 1시간마다 알람 예약을 확인하는 스케줄러
func (is *IntegratedScheduler) startAlarmScheduler() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	log.Println("🕐 알람 예약 스케줄러 시작 (1시간마다)")

	// 즉시 한 번 실행
	is.processAlarmScheduling()

	for range ticker.C {
		is.processAlarmScheduling()
	}
}

// startFCMDeliveryScheduler 1분마다 FCM 전송을 확인하는 스케줄러
func (is *IntegratedScheduler) startFCMDeliveryScheduler() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("📱 FCM 전송 스케줄러 시작 (1분마다)")

	// 즉시 한 번 실행
	is.processFCMDelivery()

	for range ticker.C {
		is.processFCMDelivery()
	}
}

// processAlarmScheduling 알람 예약 처리
func (is *IntegratedScheduler) processAlarmScheduling() {
	log.Println("🔔 알람 예약 처리 시작")

	// 모든 활성 사용자 조회
	var users []struct {
		UserUUID string `gorm:"column:user_uuid"`
	}

	if err := database.DB.Table("user_statuses").
		Select("DISTINCT user_uuid").
		Where("is_active = ? AND valid_until > ?", true, time.Now().Add(1*time.Hour)).
		Find(&users).Error; err != nil {
		log.Printf("사용자 조회 실패: %v", err)
		return
	}

	// 각 사용자별로 알람 예약 처리
	for _, user := range users {
		if err := is.processUserAlarms(user.UserUUID); err != nil {
			log.Printf("사용자 알람 처리 실패 - 사용자: %s, 에러: %v", user.UserUUID, err)
			continue
		}
	}

	log.Printf("✅ 알람 예약 처리 완료 - 처리된 사용자 수: %d", len(users))
}

// processUserAlarms 특정 사용자의 알람 예약 처리
func (is *IntegratedScheduler) processUserAlarms(userUUID string) error {
	// 사용자의 활성 상태 정보가 있는 채팅봇들 조회
	var chatbots []struct {
		ChatbotUUID string `gorm:"column:chatbot_uuid"`
	}

	if err := database.DB.Table("user_statuses").
		Select("DISTINCT chatbot_uuid").
		Where("user_uuid = ? AND is_active = ? AND valid_until > ?", userUUID, true, time.Now().Add(1*time.Hour)).
		Find(&chatbots).Error; err != nil {
		return err
	}

	// 각 채팅봇별로 알람 예약
	for _, chatbot := range chatbots {
		if err := is.alarmScheduler.ScheduleAlarmFromStatus(userUUID, chatbot.ChatbotUUID); err != nil {
			log.Printf("채팅봇 알람 예약 실패 - 사용자: %s, 채팅봇: %s, 에러: %v",
				userUUID, chatbot.ChatbotUUID, err)
			continue
		}
	}

	return nil
}

// processFCMDelivery FCM 전송 처리
func (is *IntegratedScheduler) processFCMDelivery() {
	if err := is.fcmSender.SendScheduledAlarms(); err != nil {
		log.Printf("FCM 전송 처리 실패: %v", err)
	}
}

// ProcessSessionEnd 세션 종료 시 즉시 알람 예약 처리
func (is *IntegratedScheduler) ProcessSessionEnd(userUUID, chatbotUUID string) {
	log.Printf("🔄 세션 종료 시 알람 예약 처리 - 사용자: %s, 채팅봇: %s", userUUID, chatbotUUID)

	if err := is.alarmScheduler.ScheduleAlarmFromStatus(userUUID, chatbotUUID); err != nil {
		log.Printf("세션 종료 시 알람 예약 실패: %v", err)
	}
}

// GetAlarmScheduler 알람 스케줄러 반환
func (is *IntegratedScheduler) GetAlarmScheduler() *AlarmScheduler {
	return is.alarmScheduler
}

// GetFCMSender FCM 전송기 반환
func (is *IntegratedScheduler) GetFCMSender() *FCMSender {
	return is.fcmSender
}
