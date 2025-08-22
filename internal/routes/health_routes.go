package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// SetupHealthRoutes 헬스체크 라우터 설정
func SetupHealthRoutes(app *fiber.App) {
	// 헬스체크 엔드포인트
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"timestamp": time.Now(),
			"service":   "Sermo Backend",
			"version":   "1.0.0",
		})
	})
}
