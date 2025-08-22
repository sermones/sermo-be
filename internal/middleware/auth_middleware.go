package middleware

import (
	"net/http"
	"sermo-be/pkg/jwt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware JWT 토큰을 검증하는 미들웨어
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Authorization 헤더에서 토큰 추출
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header is required",
			})
		}

		// Bearer 토큰 형식 확인
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization format. Use 'Bearer <token>'",
			})
		}

		// 토큰 추출
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token is required",
			})
		}

		// JWT 토큰 검증
		claims, err := jwt.ValidateToken(token)
		if err != nil {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// context에 사용자 ID 저장
		c.Locals("user_id", claims.UserID)
		c.Locals("user_claims", claims)

		return c.Next()
	}
}

// GetUserID context에서 사용자 ID 가져오기
func GetUserID(c *fiber.Ctx) string {
	return c.Locals("user_id").(string)
}

// GetUserClaims context에서 사용자 클레임 가져오기
func GetUserClaims(c *fiber.Ctx) *jwt.Claims {
	return c.Locals("user_claims").(*jwt.Claims)
}
