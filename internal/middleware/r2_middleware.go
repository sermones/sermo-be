package middleware

import (
	"sermo-be/internal/config"
	"sermo-be/pkg/r2"

	"github.com/gofiber/fiber/v2"
)

// R2Middleware R2 클라이언트를 context에 주입하는 미들웨어
func R2Middleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// R2 클라이언트 생성
		r2Client, err := r2.NewClient(&r2.Config{
			AccessKeyID:     cfg.R2.AccessKeyID,
			SecretAccessKey: cfg.R2.SecretAccessKey,
			Endpoint:        cfg.R2.Endpoint,
			Bucket:          cfg.R2.Bucket,
		})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to initialize R2 client",
			})
		}

		// context에 R2 클라이언트 저장
		c.Locals("r2_client", r2Client)
		return c.Next()
	}
}

// GetR2Client context에서 R2 클라이언트 가져오기
func GetR2Client(c *fiber.Ctx) *r2.Client {
	return c.Locals("r2_client").(*r2.Client)
}
