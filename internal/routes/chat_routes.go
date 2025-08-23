package routes

import (
	"sermo-be/internal/handlers/chat"
	"sermo-be/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupChatRoutes 채팅 관련 라우트 설정
func SetupChatRoutes(app *fiber.App) {
	// 채팅 라우터 그룹 (인증 필요)
	chatGroup := app.Group("/chat", middleware.AuthMiddleware())

	// 채팅 시작 (SSE 연결)
	chatGroup.Get("/start", chat.StartChat)

	// 메시지 전송
	chatGroup.Post("/send", chat.SendMessage)

	// 채팅 세션 중단
	chatGroup.Post("/stop", chat.StopChat)

	// 채팅 히스토리 조회
	chatGroup.Post("/history", chat.GetChatHistory)
}
