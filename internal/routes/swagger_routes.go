package routes

import (
	"github.com/gofiber/fiber/v2"
	swagger "github.com/swaggo/fiber-swagger"
)

func SetupSwaggerRoutes(app *fiber.App) {
	// Swagger UI 설정
	app.Get("/swagger/*", swagger.FiberWrapHandler())
}
