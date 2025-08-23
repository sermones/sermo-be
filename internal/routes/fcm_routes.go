package routes

import (
	"sermo-be/internal/handlers/fcm"
	"sermo-be/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupFCMRoutes FCM 관련 라우트 설정
func SetupFCMRoutes(app *fiber.App) {
	// FCM 라우터 그룹 (인증 필요)
	fcmGroup := app.Group("/fcm", middleware.AuthMiddleware())

	// FCM 토큰 관리
	tokenGroup := fcmGroup.Group("/token")
	tokenGroup.Post("/", fcm.CreateFCMToken)             // 토큰 생성/업데이트
	tokenGroup.Get("/user", fcm.FindFCMTokensByUserUUID) // 현재 사용자 토큰 조회
	tokenGroup.Delete("/:id", fcm.DeleteFCMToken)        // 토큰 삭제 (ID 기준)
}
