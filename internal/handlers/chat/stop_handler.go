package chat

import (
	"sermo-be/internal/core/chat"
	"sermo-be/internal/middleware"

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
	go func() {
		// OpenAI 클라이언트 가져오기
		openaiClient := middleware.GetOpenAIClient(c)
		if openaiClient != nil {
			// 통합 스케줄러 생성 및 세션 종료 시 알람 예약 처리
			scheduler := chat.NewIntegratedScheduler(openaiClient)
			scheduler.ProcessSessionEnd(userUUID, req.ChatbotUUID)
		}
	}()

	return c.JSON(fiber.Map{"message": "Chat session stopped successfully"})
}
