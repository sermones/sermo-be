package chatbot

import (
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"

	"github.com/gofiber/fiber/v2"
)

// ChatbotListResponse 채팅봇 목록 응답 DTO
type ChatbotListResponse struct {
	UUID      string   `json:"uuid"`
	Name      string   `json:"name"`
	ImageID   string   `json:"image_id"`
	Hashtags  []string `json:"hashtags"`
	Gender    string   `json:"gender"`
	Details   string   `json:"details"`
	CreatedAt string   `json:"created_at"`
}

// FindChatbotsByUserUUID 사용자별 채팅봇 목록 조회 (인증 필요)
// @Summary 사용자별 채팅봇 목록 조회
// @Description 현재 인증된 사용자의 모든 채팅봇 목록을 조회합니다.
// @Tags Chatbot
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} ChatbotListResponse
// @Failure 401 {object} map[string]interface{}
// @Router /chatbot [get]
func FindChatbotsByUserUUID(c *fiber.Ctx) error {
	// context에서 사용자 UUID 가져오기
	userUUID := middleware.GetUserUUID(c)

	// context에서 database 가져오기
	db := middleware.GetDB(c)

	// 사용자별 채팅봇 목록 조회
	var chatbots []models.Chatbot
	if err := db.Where("user_uuid = ?", userUUID).Find(&chatbots).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch chatbots",
		})
	}

	// 응답 데이터 변환
	var responses []ChatbotListResponse
	for _, chatbot := range chatbots {
		response := ChatbotListResponse{
			UUID:      chatbot.UUID.String(),
			Name:      chatbot.Name,
			ImageID:   chatbot.ImageID,
			Hashtags:  chatbot.Hashtags,
			Gender:    string(chatbot.Gender),
			Details:   chatbot.Details,
			CreatedAt: chatbot.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		responses = append(responses, response)
	}

	return c.JSON(responses)
}
