package chat

import (
	"encoding/json"
	"log"
	"time"

	"sermo-be/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// OnKeyboardRequest 키보드 입력 이벤트 요청 구조
type OnKeyboardRequest struct {
	ChatbotUUID string `json:"chatbot_uuid"`
}

// OnKeyboard 키보드 입력 이벤트 처리
// @Summary 키보드 입력 이벤트
// @Description 키보드 입력 중임을 알리는 이벤트를 처리합니다
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body OnKeyboardRequest true "키보드 입력 이벤트 요청"
// @Success 200 {object} map[string]interface{} "이벤트 전송 성공"
// @Failure 400 {object} map[string]interface{} "잘못된 요청"
// @Failure 401 {object} map[string]interface{} "인증 실패"
// @Failure 500 {object} map[string]interface{} "서버 오류"
// @Router /chat/onkeyboard [post]
func OnKeyboard(c *fiber.Ctx) error {
	// 사용자 UUID 가져오기
	userUUID := middleware.GetUserUUID(c)
	if userUUID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// 요청 바디 파싱
	var request OnKeyboardRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// chatbot_uuid 검증
	if request.ChatbotUUID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "chatbot_uuid is required"})
	}

	// SSE 매니저 가져오기
	sseManager := middleware.GetSSEManager()

	// 해당 사용자와 채팅봇의 활성 세션 찾기
	session := sseManager.FindSessionByUserAndChatbot(userUUID, request.ChatbotUUID)
	if session == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Active session not found"})
	}

	// onkeyboard 이벤트 생성
	onKeyboardEvent := map[string]interface{}{
		"type":       "onkeyboard",
		"session_id": session.SessionID,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	// JSON으로 직렬화
	eventData, err := json.Marshal(onKeyboardEvent)
	if err != nil {
		log.Printf("onkeyboard 이벤트 직렬화 실패: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create event"})
	}

	// SSE 형식으로 변환
	sseMessage := "data: " + string(eventData) + "\n\n"

	// 세션 채널로 이벤트 전송
	select {
	case session.Channel <- sseMessage:
		log.Printf("onkeyboard 이벤트 전송 성공 - 세션: %s", session.SessionID)
		return c.JSON(fiber.Map{"success": true, "message": "Event sent successfully"})
	default:
		log.Printf("세션 채널이 가득 참 - 세션: %s", session.SessionID)
		return c.Status(500).JSON(fiber.Map{"error": "Session channel is full"})
	}
}
