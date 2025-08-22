package chatbot

import (
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ChatbotResponse 채팅봇 응답 DTO
type ChatbotResponse struct {
	UUID      string   `json:"uuid"`
	Name      string   `json:"name"`
	ImageID   string   `json:"image_id"`
	Hashtags  []string `json:"hashtags"`
	Gender    string   `json:"gender"`
	Details   string   `json:"details"`
	CreatedAt string   `json:"created_at"`
}

// FindChatbotByID 특정 채팅봇 ID로 조회 (인증 필요)
// @Summary 채팅봇 ID로 조회
// @Description 특정 채팅봇 ID로 채팅봇 정보를 조회합니다.
// @Tags Chatbot
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "채팅봇 UUID"
// @Success 200 {object} ChatbotResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /chatbot/{id} [get]
func FindChatbotByID(c *fiber.Ctx) error {
	// URL 파라미터에서 채팅봇 ID 가져오기
	chatbotID := c.Params("id")
	if chatbotID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Chatbot ID is required",
		})
	}

	// UUID 형식 검증
	chatbotUUID, err := uuid.Parse(chatbotID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid chatbot ID format",
		})
	}

	// context에서 사용자 UUID 가져오기
	userUUID := middleware.GetUserUUID(c)

	// context에서 database 가져오기
	db := middleware.GetDB(c)

	// 채팅봇 조회 (사용자 소유권 확인)
	var chatbot models.Chatbot
	if err := db.Where("uuid = ? AND user_uuid = ?", chatbotUUID, userUUID).First(&chatbot).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Chatbot not found or access denied",
		})
	}

	// 응답 데이터 변환
	response := ChatbotResponse{
		UUID:      chatbot.UUID.String(),
		Name:      chatbot.Name,
		ImageID:   chatbot.ImageID,
		Hashtags:  chatbot.Hashtags,
		Gender:    string(chatbot.Gender),
		Details:   chatbot.Details,
		CreatedAt: chatbot.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return c.JSON(response)
}
