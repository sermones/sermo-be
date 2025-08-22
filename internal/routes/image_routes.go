package routes

import (
	"sermo-be/internal/handlers/image"
	"sermo-be/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupImageRoutes 이미지 관련 라우트 설정
func SetupImageRoutes(app *fiber.App) {
	// 이미지 라우터 그룹 (인증 필요)
	imageGroup := app.Group("/image", middleware.AuthMiddleware())

	// 이미지 업로드
	imageGroup.Post("/upload", image.UploadImage)

	// 이미지 삭제
	imageGroup.Delete("/delete", image.DeleteImage)
}
