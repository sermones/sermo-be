package middleware

import (
	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
)

// DatabaseMiddleware database connection을 context에 주입하는 미들웨어
func DatabaseMiddleware(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// context에 database 저장
		c.Locals("db", db)
		return c.Next()
	}
}

// GetDB context에서 database 가져오기
func GetDB(c *fiber.Ctx) *gorm.DB {
	return c.Locals("db").(*gorm.DB)
}
