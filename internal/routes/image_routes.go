package routes

import (
	"sermo-be/internal/handlers/image"
	"sermo-be/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupImageRoutes 이미지 관련 라우트 설정
func SetupImageRoutes(app *fiber.App) {
	// 이미지 라우터 그룹 (인증 필요 + Gemini 미들웨어)
	imageGroup := app.Group("/image", middleware.AuthMiddleware(), middleware.GeminiMiddleware())

	// 이미지 업로드
	imageGroup.Post("/upload", image.UploadImage)

	// AI 이미지 생성
	imageGroup.Post("/generate", image.GenerateImage)

	// 이미지 조회
	imageGroup.Get("/:image_id", image.GetImage)

	// 이미지 다운로드
	imageGroup.Get("/:image_id/download", image.DownloadImage)

	// 이미지 삭제
	imageGroup.Delete("/delete", image.DeleteImage)
}
