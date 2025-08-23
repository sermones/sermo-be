package routes

import (
	"sermo-be/internal/handlers/bookmark"
	"sermo-be/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupBookmarkRoutes 북마크 라우터 설정
func SetupBookmarkRoutes(app *fiber.App) {
	// 북마크 라우터 그룹 (인증 필요)
	bookmarkGroup := app.Group("/bookmark", middleware.AuthMiddleware())

	// 문장 북마크 라우트
	bookmarkGroup.Post("/sentence", bookmark.CreateSentenceBookmark)
	bookmarkGroup.Get("/sentence", bookmark.FindByUserUUIDSentenceBookmark)
	bookmarkGroup.Get("/sentence/date", bookmark.FindByDateSentenceBookmark)
}
