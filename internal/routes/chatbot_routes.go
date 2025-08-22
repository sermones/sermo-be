package routes

import (
	"sermo-be/internal/handlers/chatbot"
	"sermo-be/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupChatbotRoutes 채팅봇 관련 라우터 설정
func SetupChatbotRoutes(app *fiber.App) {
	// 채팅봇 라우터 그룹 (인증 필요)
	chatbotGroup := app.Group("/chatbot", middleware.AuthMiddleware())

	// 채팅봇 생성
	chatbotGroup.Post("/", chatbot.CreateChatbot)

	// 사용자별 채팅봇 목록 조회
	chatbotGroup.Get("/", chatbot.FindChatbotsByUserUUID)

	// 특정 채팅봇 ID로 조회
	chatbotGroup.Get("/:id", chatbot.FindChatbotByID)
}
