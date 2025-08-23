package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"sermo-be/internal/core/status"
	"sermo-be/internal/models"
	"sermo-be/pkg/database"
	"sermo-be/pkg/openai"
	"sermo-be/pkg/prompt"
	"sermo-be/pkg/redis"
)

// ScheduledAlarm 예약된 알람 정보
type ScheduledAlarm struct {
	UserUUID      string    `json:"user_uuid"`
	ChatbotUUID   string    `json:"chatbot_uuid"`
	ChatbotName   string    `json:"chatbot_name"`
	ChatbotAvatar string    `json:"chatbot_avatar"`
	Message       string    `json:"message"`
	ScheduledAt   time.Time `json:"scheduled_at"`
	Priority      int       `json:"priority"` // 1: 높음, 2: 보통, 3: 낮음
	Context       string    `json:"context"`  // 원본 상태 정보
	CreatedAt     time.Time `json:"created_at"`
}

// AlarmScheduler 알람 스케줄링을 담당하는 구조체
type AlarmScheduler struct {
	statusService *status.StatusService
	openaiClient  *openai.Client
}

// NewAlarmScheduler 새로운 AlarmScheduler 생성
func NewAlarmScheduler(openaiClient *openai.Client) *AlarmScheduler {
	return &AlarmScheduler{
		statusService: status.GetStatusService(),
		openaiClient:  openaiClient,
	}
}

// ScheduleAlarmFromStatus 상태 정보를 바탕으로 알람 예약
func (as *AlarmScheduler) ScheduleAlarmFromStatus(userUUID, chatbotUUID string) error {
	// 사용자의 활성 상태 정보 조회
	var userStatuses []models.UserStatus
	if err := database.DB.Where("user_uuid = ? AND chatbot_uuid = ? AND is_active = ? AND valid_until > ?",
		userUUID, chatbotUUID, true, time.Now().Add(1*time.Hour)).Find(&userStatuses).Error; err != nil {
		return fmt.Errorf("사용자 상태 정보 조회 실패: %w", err)
	}

	if len(userStatuses) == 0 {
		return nil // 예약할 알람이 없음
	}

	// 채팅봇 정보 조회
	var chatbot models.Chatbot
	if err := database.DB.Where("uuid = ?", chatbotUUID).First(&chatbot).Error; err != nil {
		return fmt.Errorf("채팅봇 정보 조회 실패: %w", err)
	}

	// AI가 알람 예약 여부와 메시지를 판단
	alarmDecisions := as.generateAlarmDecisions(userStatuses, &chatbot)
	if len(alarmDecisions) == 0 {
		return nil
	}

	// Redis에 알람 예약
	return as.scheduleAlarmsInRedis(alarmDecisions)
}

// generateAlarmDecisions AI가 알람 예약 여부와 메시지를 판단
func (as *AlarmScheduler) generateAlarmDecisions(userStatuses []models.UserStatus, chatbot *models.Chatbot) []ScheduledAlarm {
	var alarms []ScheduledAlarm

	// 상태 정보들을 AI에게 전달하여 알람 예약 여부 판단
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Second)
	defer cancel()

	// AI 판단을 위한 프롬프트 구성
	decisionPrompt := prompt.GetAlarmDecisionPrompt(userStatuses, chatbot)

	messages := []openai.ChatMessage{
		{
			Role:    "system",
			Content: prompt.GetAlarmSchedulingSystemPrompt(),
		},
		{
			Role:    "user",
			Content: decisionPrompt,
		},
	}

	// AI 응답 생성
	response, err := as.openaiClient.ChatCompletion(ctx, messages)
	if err != nil {
		log.Printf("AI 알람 판단 실패: %v", err)
		return alarms
	}

	// AI 응답 파싱하여 알람 결정
	alarms = as.parseAlarmDecisions(response.Message.Content, userStatuses, chatbot)
	return alarms
}

// parseAlarmDecisions AI 응답을 파싱하여 알람 결정
func (as *AlarmScheduler) parseAlarmDecisions(aiResponse string, userStatuses []models.UserStatus, chatbot *models.Chatbot) []ScheduledAlarm {
	var alarms []ScheduledAlarm

	// AI 응답에서 JSON 부분 추출
	jsonStart := strings.Index(aiResponse, "[")
	jsonEnd := strings.LastIndex(aiResponse, "]")
	if jsonStart == -1 || jsonEnd == -1 {
		log.Printf("AI 응답에서 JSON 배열을 찾을 수 없음: %s", aiResponse)
		return alarms
	}

	jsonStr := aiResponse[jsonStart : jsonEnd+1]

	var alarmData []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &alarmData); err != nil {
		log.Printf("AI 응답 JSON 파싱 실패: %v", err)
		return alarms
	}

	// 파싱된 데이터를 ScheduledAlarm으로 변환
	for _, data := range alarmData {
		alarm, err := as.convertToScheduledAlarm(data, chatbot)
		if err != nil {
			log.Printf("알람 데이터 변환 실패: %v", err)
			continue
		}
		alarms = append(alarms, alarm)
	}

	return alarms
}

// convertToScheduledAlarm 데이터를 ScheduledAlarm으로 변환
func (as *AlarmScheduler) convertToScheduledAlarm(data map[string]interface{}, chatbot *models.Chatbot) (ScheduledAlarm, error) {
	alarm := ScheduledAlarm{
		UserUUID:      data["user_uuid"].(string),
		ChatbotUUID:   data["chatbot_uuid"].(string),
		ChatbotName:   data["chatbot_name"].(string),
		ChatbotAvatar: data["chatbot_avatar"].(string),
		Message:       data["message"].(string),
		Priority:      int(data["priority"].(float64)),
		Context:       data["context"].(string),
		CreatedAt:     time.Now(),
	}

	// scheduled_at 파싱
	scheduledAtStr := data["scheduled_at"].(string)
	scheduledAt, err := time.Parse(time.RFC3339, scheduledAtStr)
	if err != nil {
		return alarm, fmt.Errorf("시간 파싱 실패: %w", err)
	}
	alarm.ScheduledAt = scheduledAt

	return alarm, nil
}

// scheduleAlarmsInRedis Redis에 알람 예약
func (as *AlarmScheduler) scheduleAlarmsInRedis(alarms []ScheduledAlarm) error {
	ctx := context.Background()

	for _, alarm := range alarms {
		// Redis 키 생성: alarm:{userUUID}:{timestamp}
		key := fmt.Sprintf("alarm:%s:%d", alarm.UserUUID, alarm.ScheduledAt.Unix())

		// JSON으로 직렬화
		alarmData, err := json.Marshal(alarm)
		if err != nil {
			log.Printf("알람 데이터 직렬화 실패: %v", err)
			continue
		}

		// Redis에 저장 (TTL은 예약 시간 + 1시간)
		ttl := time.Until(alarm.ScheduledAt) + time.Hour
		if err := redis.SetKey(ctx, key, string(alarmData), ttl); err != nil {
			log.Printf("Redis 알람 저장 실패: %v", err)
			continue
		}

		log.Printf("알람 예약 완료 - 사용자: %s, 시간: %s, 메시지: %s",
			alarm.UserUUID, alarm.ScheduledAt.Format("2006-01-02 15:04:05"), alarm.Message)
	}

	return nil
}
