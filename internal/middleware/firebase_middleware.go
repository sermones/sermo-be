package middleware

import (
	"log"
	"sermo-be/internal/config"
	"sermo-be/pkg/firebase"

	"github.com/gofiber/fiber/v2"
)

const (
	FirebaseClientKey = "firebase_client"
)

// FirebaseMiddleware Firebase 클라이언트를 컨텍스트에 주입하는 미들웨어
func FirebaseMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Firebase 클라이언트 생성
		firebaseClient, err := firebase.NewClient(cfg)
		if err != nil {
			log.Printf("Failed to create Firebase client: %v", err)
			// Firebase 클라이언트 생성 실패 시에도 요청은 계속 진행
			// (Firebase 기능이 필수가 아닌 경우)
			return c.Next()
		}

		// 컨텍스트에 Firebase 클라이언트 저장
		c.Locals(FirebaseClientKey, firebaseClient)

		// 요청 완료 후 클라이언트 정리
		defer func() {
			if err := firebaseClient.Close(); err != nil {
				log.Printf("Failed to close Firebase client: %v", err)
			}
		}()

		return c.Next()
	}
}

// GetFirebaseClient 컨텍스트에서 Firebase 클라이언트 가져오기
func GetFirebaseClient(c *fiber.Ctx) *firebase.Client {
	if client, ok := c.Locals(FirebaseClientKey).(*firebase.Client); ok {
		return client
	}
	return nil
}
