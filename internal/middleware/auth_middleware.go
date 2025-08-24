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
		var token string

		// 1. 쿼리 파라미터에서 토큰 확인 (EventSource용)
		if queryToken := c.Query("token"); queryToken != "" {
			token = queryToken
		} else {
			// 2. Authorization 헤더에서 토큰 추출 (일반 HTTP 요청용)
			authHeader := c.Get("Authorization")
			if authHeader == "" {
				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
					"error": "Authorization header or token query parameter is required",
				})
			}

			// Bearer 토큰 형식 확인
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid authorization format. Use 'Bearer <token>'",
				})
			}

			// 토큰 추출
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

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

		// context에 사용자 UUID 저장
		c.Locals("user_uuid", claims.UUID)
		c.Locals("user_claims", claims)

		return c.Next()
	}
}

// GetUserUUID context에서 사용자 UUID 가져오기
func GetUserUUID(c *fiber.Ctx) string {
	return c.Locals("user_uuid").(string)
}

// GetUserClaims context에서 사용자 클레임 가져오기
func GetUserClaims(c *fiber.Ctx) *jwt.Claims {
	return c.Locals("user_claims").(*jwt.Claims)
}
