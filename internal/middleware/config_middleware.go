package middleware

import (
	"sermo-be/internal/config"

	"github.com/gofiber/fiber/v2"
)

// ConfigMiddleware config를 context에 주입하는 미들웨어
func ConfigMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// context에 config 저장
		c.Locals("config", cfg)
		return c.Next()
	}
}

// GetConfig context에서 config 가져오기
func GetConfig(c *fiber.Ctx) *config.Config {
	return c.Locals("config").(*config.Config)
}
