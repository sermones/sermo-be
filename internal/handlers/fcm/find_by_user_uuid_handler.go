package fcm

import (
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"

	"github.com/gofiber/fiber/v2"
)

// FindFCMTokensByUserUUID 사용자별 FCM 토큰 조회 (인증 필요)
// @Summary 사용자별 FCM 토큰 조회
// @Description 현재 인증된 사용자의 모든 FCM 토큰을 조회합니다.
// @Tags FCM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.FCMToken
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /fcm/token/user [get]
func FindFCMTokensByUserUUID(c *fiber.Ctx) error {
	// context에서 사용자 UUID 가져오기
	userUUID := middleware.GetUserUUID(c)

	// context에서 database 가져오기
	db := middleware.GetDB(c)

	// 사용자별 FCM 토큰 목록 조회
	var tokens []models.FCMToken
	if err := db.Where("user_uuid = ?", userUUID).Find(&tokens).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch FCM tokens",
		})
	}

	return c.JSON(tokens)
}
