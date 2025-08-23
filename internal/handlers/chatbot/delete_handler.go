package chatbot

import (
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// DeleteChatbotResponse 채팅봇 삭제 응답 DTO
type DeleteChatbotResponse struct {
	Message string `json:"message"`
}

// DeleteChatbot 채팅봇 삭제 (인증 필요)
// @Summary 채팅봇 삭제
// @Description 특정 채팅봇 ID로 채팅봇을 삭제합니다.
// @Tags Chatbot
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "채팅봇 UUID"
// @Success 200 {object} DeleteChatbotResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /chatbot/{id} [delete]
func DeleteChatbot(c *fiber.Ctx) error {
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

	// 채팅봇 존재 여부 및 소유권 확인
	var chatbot models.Chatbot
	if err := db.Where("uuid = ? AND user_uuid = ?", chatbotUUID, userUUID).First(&chatbot).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Chatbot not found or access denied",
		})
	}

	// 채팅봇 삭제
	if err := db.Delete(&chatbot).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete chatbot",
		})
	}

	response := DeleteChatbotResponse{
		Message: "Chatbot deleted successfully",
	}

	return c.JSON(response)
}
