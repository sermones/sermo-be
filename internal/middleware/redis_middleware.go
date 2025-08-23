package middleware

import (
	"context"
	"sermo-be/pkg/redis"

	"github.com/gofiber/fiber/v2"
)

const RedisContextKey = "redis"

// RedisMiddleware Redis 클라이언트를 컨텍스트에 주입하는 미들웨어
func RedisMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.WithValue(c.Context(), RedisContextKey, redis.RedisClient)
		c.Locals("redis", redis.RedisClient)
		c.SetUserContext(ctx)
		return c.Next()
	}
}

// GetRedis 컨텍스트에서 Redis 클라이언트를 가져오는 함수
func GetRedis(c *fiber.Ctx) interface{} {
	return c.Locals("redis")
}
