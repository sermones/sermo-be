package middleware

import (
	"log"

	"sermo-be/internal/config"
	"sermo-be/pkg/openai"

	"github.com/gofiber/fiber/v2"
)

// OpenAI 미들웨어
func OpenAIMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// OpenAI 클라이언트 생성
		openaiClient, err := openai.NewClient(&openai.Config{
			APIKey:              cfg.OpenAI.APIKey,
			Model:               cfg.OpenAI.Model,
			MaxCompletionTokens: cfg.OpenAI.MaxCompletionTokens,
		})
		if err != nil {
			log.Printf("OpenAI 클라이언트 생성 실패: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "OpenAI service unavailable"})
		}

		// 컨텍스트에 OpenAI 클라이언트 저장
		c.Locals("openai_client", openaiClient)

		return c.Next()
	}
}

// GetOpenAIClient 컨텍스트에서 OpenAI 클라이언트 가져오기
func GetOpenAIClient(c *fiber.Ctx) *openai.Client {
	if client, ok := c.Locals("openai_client").(*openai.Client); ok {
		return client
	}
	return nil
}
