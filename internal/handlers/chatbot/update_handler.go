package chatbot

import (
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UpdateChatbotRequest 채팅봇 수정 요청 DTO
type UpdateChatbotRequest struct {
	Name     string   `json:"name"`     // 채팅봇 이름 (3-100자)
	ImageID  string   `json:"image_id"` // 사진 ID
	Hashtags []string `json:"hashtags"` // 해시태그 배열
	Gender   string   `json:"gender"`   // 성별 (male, female, unspecified)
	Details  string   `json:"details"`  // 상세 설명
}

// UpdateChatbotResponse 채팅봇 수정 응답 DTO
type UpdateChatbotResponse struct {
	Message string `json:"message"`
}

// UpdateChatbot 채팅봇 수정 (인증 필요)
// @Summary 채팅봇 수정
// @Description 특정 채팅봇 ID로 채팅봇 내용을 수정합니다. UUID와 UserID는 변경할 수 없습니다.
// @Tags Chatbot
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "채팅봇 UUID"
// @Param request body UpdateChatbotRequest true "수정할 채팅봇 정보"
// @Success 200 {object} UpdateChatbotResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /chatbot/{id} [put]
func UpdateChatbot(c *fiber.Ctx) error {
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

	// 요청 파싱
	var req UpdateChatbotRequest
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

	// 채팅봇 존재 여부 및 소유권 확인
	var existingChatbot models.Chatbot
	if err := db.Where("uuid = ? AND user_uuid = ?", chatbotUUID, userUUID).First(&existingChatbot).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Chatbot not found or access denied",
		})
	}

	// 채팅봇 내용 업데이트 (UUID와 UserID는 유지)
	updates := map[string]interface{}{
		"name":       req.Name,
		"image_id":   req.ImageID,
		"hashtags":   req.Hashtags,
		"gender":     gender,
		"details":    req.Details,
		"updated_at": time.Now(),
	}

	// 데이터베이스 업데이트
	if err := db.Model(&existingChatbot).Updates(updates).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update chatbot",
		})
	}

	response := UpdateChatbotResponse{
		Message: "Chatbot updated successfully",
	}

	return c.JSON(response)
}
