package user

import (
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"

	"github.com/gofiber/fiber/v2"
)

// ProfileResponse 사용자 프로필 응답
type ProfileResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// GetProfile 사용자 프로필 조회 (인증 필요)
func GetProfile(c *fiber.Ctx) error {
	// context에서 사용자 ID 가져오기
	userID := middleware.GetUserID(c)

	// context에서 database 가져오기
	db := middleware.GetDB(c)

	// 데이터베이스에서 사용자 조회
	var user models.User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	response := ProfileResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return c.JSON(response)
}
