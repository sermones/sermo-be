package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// SetupHealthRoutes 헬스체크 라우터 설정
func SetupHealthRoutes(app *fiber.App) {
	// 헬스체크 엔드포인트
	// @Summary 헬스체크
	// @Description 서버 상태를 확인합니다.
	// @Tags System
	// @Accept json
	// @Produce json
	// @Success 200 {object} map[string]interface{}
	// @Router /health [get]
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"timestamp": time.Now(),
			"service":   "Sermo Backend",
			"version":   "1.0.0",
		})
	})
}
