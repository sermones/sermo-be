package middleware

import (
	"context"
	"sermo-be/pkg/gemini"

	"github.com/gofiber/fiber/v2"
)

// GeminiContextKey Gemini 클라이언트를 위한 context 키
type GeminiContextKey struct{}

// GeminiMiddleware Gemini 클라이언트를 context에 주입하는 미들웨어
func GeminiMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// context에서 config 가져오기
		cfg := GetConfig(c)
		if cfg == nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Config not available",
			})
		}

		// Gemini 이미지 생성 클라이언트 생성
		geminiClient, err := gemini.NewImageClient(&gemini.Config{
			APIKey: cfg.Gemini.APIKey,
		})
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to initialize Gemini client",
			})
		}

		// context에 Gemini 클라이언트 주입
		ctx := context.WithValue(c.Context(), GeminiContextKey{}, geminiClient)
		c.SetUserContext(ctx)

		// 다음 핸들러 호출
		if err := c.Next(); err != nil {
			return err
		}

		// 리소스 정리
		geminiClient.Close()
		return nil
	}
}

// GetGeminiClient context에서 Gemini 이미지 생성 클라이언트 가져오기
func GetGeminiClient(c *fiber.Ctx) *gemini.ImageClient {
	if client, ok := c.UserContext().Value(GeminiContextKey{}).(*gemini.ImageClient); ok {
		return client
	}
	return nil
}
