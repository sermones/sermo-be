package fcm

import (
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CreateFCMTokenRequest struct {
	FCMToken   string `json:"fcm_token" binding:"required"`
	DeviceInfo string `json:"device_info"`
}

type CreateFCMTokenResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    *models.FCMToken `json:"data,omitempty"`
}

// CreateFCMToken FCM 토큰 생성/업데이트 (인증 필요)
// @Summary FCM 토큰 생성/업데이트
// @Description FCM 토큰을 생성하거나 기존 토큰을 업데이트합니다.
// @Tags FCM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateFCMTokenRequest true "FCM 토큰 정보"
// @Success 201 {object} CreateFCMTokenResponse
// @Success 200 {object} CreateFCMTokenResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /fcm/token [post]

func CreateFCMToken(c *fiber.Ctx) error {
	// 요청 파싱
	var req CreateFCMTokenRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// context에서 사용자 UUID 가져오기
	userUUIDStr := middleware.GetUserUUID(c)
	userUUID, err := uuid.Parse(userUUIDStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user UUID",
		})
	}

	// context에서 database 가져오기
	db := middleware.GetDB(c)

	// FCM 토큰 중복 확인
	var existingToken models.FCMToken
	if err := db.Where("fcm_token = ?", req.FCMToken).First(&existingToken).Error; err == nil {
		// 기존 토큰이 있으면 업데이트
		updates := map[string]interface{}{
			"device_info": req.DeviceInfo,
			"user_uuid":   userUUID,
		}

		if err := db.Model(&existingToken).Updates(updates).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update FCM token",
			})
		}

		response := CreateFCMTokenResponse{
			Success: true,
			Message: "기존 토큰이 업데이트되었습니다",
			Data:    &existingToken,
		}

		return c.Status(http.StatusOK).JSON(response)
	}

	// 새 토큰 생성 (인증된 사용자로)
	newToken := models.NewFCMTokenWithUser(userUUID, req.FCMToken, req.DeviceInfo)

	if err := db.Create(newToken).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create FCM token",
		})
	}

	response := CreateFCMTokenResponse{
		Success: true,
		Message: "FCM 토큰이 성공적으로 생성되었습니다",
		Data:    newToken,
	}

	return c.Status(http.StatusCreated).JSON(response)
}
