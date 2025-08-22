package chatbot

import (
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"

	"github.com/gofiber/fiber/v2"
)

// CreateChatbotRequest 채팅봇 생성 요청 DTO
type CreateChatbotRequest struct {
	Name     string   `json:"name"`     // 채팅봇 이름 (3-100자)
	ImageID  string   `json:"image_id"` // 사진 ID
	Hashtags []string `json:"hashtags"` // 해시태그 배열
	Gender   string   `json:"gender"`   // 성별 (male, female, unspecified)
	Details  string   `json:"details"`  // 상세 설명
}

// CreateChatbotResponse 채팅봇 생성 응답 DTO
type CreateChatbotResponse struct {
	Message   string `json:"message"`
	ChatbotID string `json:"chatbot_id"`
}

// CreateChatbot 채팅봇 생성 (인증 필요)
// @Summary 채팅봇 생성
// @Description 새로운 채팅봇을 생성합니다.
// @Tags Chatbot
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateChatbotRequest true "채팅봇 생성 정보"
// @Success 201 {object} CreateChatbotResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /chatbot [post]
func CreateChatbot(c *fiber.Ctx) error {
	// 요청 파싱
	var req CreateChatbotRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// 입력 검증
	if req.Name == "" || len(req.Name) < 3 || len(req.Name) > 100 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Name must be between 3 and 100 characters",
		})
	}

	if req.ImageID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Image ID is required",
		})
	}

	// 성별 검증
	gender := models.GenderUnspecified
	switch req.Gender {
	case "male":
		gender = models.GenderMale
	case "female":
		gender = models.GenderFemale
	case "unspecified":
		gender = models.GenderUnspecified
	default:
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid gender. Must be 'male', 'female', or 'unspecified'",
		})
	}

	// context에서 사용자 UUID 가져오기
	userUUID := middleware.GetUserUUID(c)

	// context에서 database 가져오기
	db := middleware.GetDB(c)

	// 채팅봇 생성
	chatbot := models.NewChatbot(
		req.Name,
		req.ImageID,
		req.Hashtags,
		gender,
		req.Details,
		userUUID,
	)

	// 데이터베이스에 저장
	if err := db.Create(chatbot).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create chatbot",
		})
	}

	response := CreateChatbotResponse{
		Message:   "Chatbot created successfully",
		ChatbotID: chatbot.UUID.String(),
	}

	return c.Status(http.StatusCreated).JSON(response)
}
