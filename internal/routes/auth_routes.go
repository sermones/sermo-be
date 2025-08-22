package routes

import (
	"sermo-be/internal/handlers/auth"

	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(app *fiber.App) {
	// 인증 라우터 그룹
	authGroup := app.Group("/auth")

	// 회원가입
	authGroup.Post("/signup", auth.SignUp)

	// 로그인
	authGroup.Post("/login", auth.Login)
}
