package routes

import (
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// 헬스체크 라우터 설정
	SetupHealthRoutes(app)

	// 인증 라우터 설정
	SetupAuthRoutes(app)

	// 사용자 라우터 설정
	SetupUserRoutes(app)
}
