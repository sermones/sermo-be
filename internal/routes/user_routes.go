package routes

import (
	"sermo-be/internal/handlers/user"
	"sermo-be/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupUserRoutes 사용자 관련 라우터 설정
func SetupUserRoutes(app *fiber.App) {
	// 사용자 라우터 그룹 (인증 필요)
	userGroup := app.Group("/user", middleware.AuthMiddleware())

	// 프로필 조회
	userGroup.Get("/profile", user.GetProfile)
}
