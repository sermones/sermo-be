package chat

import (
	"log"
	"sermo-be/internal/core/chat"
	"sermo-be/internal/middleware"
	"sermo-be/pkg/database"

	"github.com/gofiber/fiber/v2"
)

// StopChatRequest 채팅 중단 요청 DTO
type StopChatRequest struct {
	ChatbotUUID string `json:"chatbot_uuid" validate:"required"`
}

// StopChat 채팅 세션 중단
// @Summary 채팅 세션 중단
// @Description 활성 채팅 세션을 중단합니다
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body StopChatRequest true "세션 중단 요청"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /chat/stop [post]
func StopChat(c *fiber.Ctx) error {
	// 사용자 UUID 가져오기
	userUUID := middleware.GetUserUUID(c)
	if userUUID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// 요청 파싱
	var req StopChatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// SSE 매니저 가져오기
	sseManager := middleware.GetSSEManager()

	// 사용자의 활성 세션 찾기
	userSessions := sseManager.GetUserSessions(userUUID)
	var targetSession *middleware.SSESession

	for _, session := range userSessions {
		if session.ChatbotUUID == req.ChatbotUUID {
			targetSession = session
			break
		}
	}

	if targetSession == nil {
		return c.Status(400).JSON(fiber.Map{"error": "No active session found"})
	}

	// 세션 중단
	if err := sseManager.StopSession(targetSession.SessionID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// 세션 종료 시 알람 예약 처리
	// 고루틴 시작 전에 필요한 데이터를 미리 가져오기
	openaiClient := middleware.GetOpenAIClient(c)
	if openaiClient != nil {
		go func() {
			log.Printf("🔄 알람 메시지 생성 시작 - 사용자: %s, 챗봇: %s", userUUID, req.ChatbotUUID)

			// 데이터베이스 연결 확인
			if database.DB == nil {
				log.Printf("❌ 데이터베이스 연결이 없음 - 알람 생성 중단")
				return
			}

			// 알람 메시지 생성 및 데이터베이스 저장
			config := chat.AlarmMessageConfig{
				UserUUID:    userUUID,
				ChatbotUUID: req.ChatbotUUID,
			}

			log.Printf("📝 알람 메시지 생성 중...")
			alarmMessage, err := chat.AlarmMessageGeneate(openaiClient, database.DB, config)
			if err != nil {
				log.Printf("❌ 알람 메시지 생성 실패: %v", err)
				return
			}

			log.Printf("✅ 알람 메시지 생성 성공 - 전송 시간: %s", alarmMessage.SendTime.Format("2006-01-02 15:04:05"))

			// FCM 즉시 전송
			if err := chat.SendImmediateFCMNotification(userUUID, alarmMessage.Message, req.ChatbotUUID); err != nil {
				log.Printf("❌ FCM 전송 실패: %v", err)
			} else {
				log.Printf("✅ FCM 알람 전송 완료")
			}
		}()
	} else {
		log.Printf("❌ OpenAI 클라이언트를 가져올 수 없음 - 알람 생성 중단")
	}

	return c.JSON(fiber.Map{"message": "Chat session stopped successfully"})
}
