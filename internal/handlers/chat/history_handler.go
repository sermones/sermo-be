package chat

import (
	"time"

	"sermo-be/internal/middleware"
	"sermo-be/internal/models"
	"sermo-be/pkg/database"

	"github.com/gofiber/fiber/v2"
)

// ChatHistoryRequest 채팅 히스토리 요청 DTO
type ChatHistoryRequest struct {
	ChatbotUUID string `json:"chatbot_uuid" validate:"required"`
	Limit       int    `json:"limit" validate:"min=1,max=100"`
	Offset      int    `json:"offset" validate:"min=0"`
}

// ChatHistoryResponse 채팅 히스토리 응답 DTO
type ChatHistoryResponse struct {
	Messages []ChatMessageResponse `json:"messages"`
	Total    int64                 `json:"total"`
}

// ChatMessageResponse 채팅 메시지 응답 DTO
type ChatMessageResponse struct {
	UUID        string `json:"uuid"`
	MessageType string `json:"message_type"`
	Content     string `json:"content"`
	CreatedAt   string `json:"created_at"`
}

// GetChatHistory 채팅 히스토리 조회
// @Summary 채팅 히스토리 조회
// @Description 특정 채팅봇과의 대화 히스토리를 조회합니다
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChatHistoryRequest true "히스토리 요청"
// @Success 200 {object} ChatHistoryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /chat/history [post]
func GetChatHistory(c *fiber.Ctx) error {
	// 사용자 UUID 가져오기
	userUUID := middleware.GetUserUUID(c)
	if userUUID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// 요청 파싱
	var req ChatHistoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// 기본값 설정
	if req.Limit == 0 {
		req.Limit = 50
	}

	// 총 메시지 수 조회
	var total int64
	if err := database.DB.Model(&models.ChatMessage{}).
		Where("user_uuid = ? AND chatbot_uuid = ?", userUUID, req.ChatbotUUID).
		Count(&total).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count messages"})
	}

	// 메시지 조회
	var messages []models.ChatMessage
	if err := database.DB.Where("user_uuid = ? AND chatbot_uuid = ?", userUUID, req.ChatbotUUID).
		Order("created_at DESC").
		Limit(req.Limit).
		Offset(req.Offset).
		Find(&messages).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch messages"})
	}

	// 응답 변환
	var messageResponses []ChatMessageResponse
	for _, msg := range messages {
		messageResponses = append(messageResponses, ChatMessageResponse{
			UUID:        msg.UUID.String(),
			MessageType: string(msg.MessageType),
			Content:     msg.Content,
			CreatedAt:   msg.CreatedAt.Format(time.RFC3339),
		})
	}

	response := ChatHistoryResponse{
		Messages: messageResponses,
		Total:    total,
	}

	return c.JSON(response)
}
