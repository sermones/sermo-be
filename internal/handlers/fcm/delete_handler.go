package fcm

import (
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"

	"github.com/gofiber/fiber/v2"
)

// DeleteFCMToken FCM 토큰 삭제 (인증 필요)
// @Summary FCM 토큰 삭제
// @Description 현재 인증된 사용자의 특정 FCM 토큰을 삭제합니다 (soft delete).
// @Tags FCM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "FCM 토큰 ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /fcm/token/{id} [delete]
func DeleteFCMToken(c *fiber.Ctx) error {
	// URL 파라미터에서 FCM 토큰 ID 가져오기
	tokenID := c.Params("id")
	if tokenID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "FCM token ID is required",
		})
	}

	// context에서 사용자 UUID 가져오기
	userUUID := middleware.GetUserUUID(c)

	// context에서 database 가져오기
	db := middleware.GetDB(c)

	// FCM 토큰 존재 여부 및 소유권 확인
	var fcmTokenModel models.FCMToken
	if err := db.Where("id = ? AND user_uuid = ?", tokenID, userUUID).First(&fcmTokenModel).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "FCM token not found or access denied",
		})
	}

	// FCM 토큰 삭제 (soft delete)
	if err := db.Delete(&fcmTokenModel).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete FCM token",
		})
	}

	return c.JSON(fiber.Map{
		"message": "FCM token deleted successfully",
	})
}
