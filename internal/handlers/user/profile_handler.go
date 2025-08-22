package user

import (
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"

	"github.com/gofiber/fiber/v2"
)

// ProfileResponse 사용자 프로필 응답
type ProfileResponse struct {
	UUID      string `json:"uuid"`
	ID        string `json:"id"`
	Nickname  string `json:"nickname"`
	CreatedAt string `json:"created_at"`
}

// GetProfile 사용자 프로필 조회 (인증 필요)
// @Summary 사용자 프로필 조회
// @Description 현재 인증된 사용자의 프로필 정보를 조회합니다.
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ProfileResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /user/profile [get]
func GetProfile(c *fiber.Ctx) error {
	// context에서 사용자 ID 가져오기
	userUUID := middleware.GetUserUUID(c)

	// context에서 database 가져오기
	db := middleware.GetDB(c)

	// 데이터베이스에서 사용자 조회
	var user models.User
	if err := db.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	response := ProfileResponse{
		UUID:      user.UUID.String(),
		ID:        user.ID,
		Nickname:  user.Nickname,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return c.JSON(response)
}
