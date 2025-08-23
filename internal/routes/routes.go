package routes

import (
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// 헬스체크 라우터 설정
	SetupHealthRoutes(app)

	// Swagger UI 라우터 설정
	SetupSwaggerRoutes(app)

	// 인증 라우터 설정
	SetupAuthRoutes(app)

	// 사용자 라우터 설정
	SetupUserRoutes(app)

	// 이미지 라우터 설정
	SetupImageRoutes(app)

	// 채팅봇 라우터 설정
	SetupChatbotRoutes(app)

	// 채팅 라우터 설정
	SetupChatRoutes(app)

	// 북마크 라우터 설정
	SetupBookmarkRoutes(app)
}
