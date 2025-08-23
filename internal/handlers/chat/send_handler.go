package chat

import (
	"encoding/json"
	"fmt"
	"time"

	"sermo-be/internal/core/chat"
	"sermo-be/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// SendMessageRequest 메시지 전송 요청 DTO
type SendMessageRequest struct {
	ChatbotUUID string `json:"chatbot_uuid" validate:"required"`
	Message     string `json:"message" validate:"required"`
}

// SendMessageResponse 메시지 전송 응답 DTO
type SendMessageResponse struct {
	SessionID   string `json:"session_id"`
	Message     string `json:"message"`
	MessageType string `json:"message_type"`
	Timestamp   string `json:"timestamp"`
}

// SSEMessage SSE 메시지 구조
type SSEMessage struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	SessionID string `json:"session_id"`
}

// SendMessage 메시지 전송
// @Summary 메시지 전송
// @Description 기존 세션에 메시지를 전송합니다
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SendMessageRequest true "메시지 요청"
// @Success 200 {object} SendMessageResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /chat/send [post]
func SendMessage(c *fiber.Ctx) error {
	// 사용자 UUID 가져오기
	userUUID := middleware.GetUserUUID(c)
	if userUUID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// 요청 파싱
	var req SendMessageRequest
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

	// 1. 먼저 사용자 메시지를 SSE로 전송
	currentTime := time.Now()
	userSSEMessage := SSEMessage{
		Type:      "user",
		Content:   req.Message,
		Timestamp: currentTime.Format(time.RFC3339),
		SessionID: targetSession.SessionID,
	}
	userSSEData, _ := json.Marshal(userSSEMessage)
	userMessageData := fmt.Sprintf("data: %s\n\n", string(userSSEData))

	// SSE 세션에 사용자 메시지 전송
	if err := sseManager.SendMessage(targetSession.SessionID, userMessageData); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to send user message via SSE"})
	}

	// 2. SSE 전송 성공 시에만 DB에 저장
	messageService := chat.GetMessageService()
	userChatMessage, err := messageService.CreateUserMessage(
		targetSession.SessionID,
		userUUID,
		req.ChatbotUUID,
		req.Message,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// 응답 반환
	response := SendMessageResponse{
		SessionID:   targetSession.SessionID,
		Message:     userChatMessage.Content,
		MessageType: "user",
		Timestamp:   userChatMessage.CreatedAt.Format(time.RFC3339),
	}

	return c.JSON(response)
}
