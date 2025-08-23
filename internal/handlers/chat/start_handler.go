package chat

import (
	"bufio"
	"fmt"
	"log"
	"time"

	"sermo-be/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// StartChat 채팅 시작 및 SSE 연결
// @Summary 채팅 시작
// @Description 채팅을 시작하고 SSE 연결을 설정합니다
// @Tags Chat
// @Accept json
// @Produce text/event-stream
// @Security BearerAuth
// @Param chatbot_uuid query string true "채팅봇 UUID"
// @Success 200 {string} string "SSE 스트림"
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /chat/start [get]
func StartChat(c *fiber.Ctx) error {
	// 사용자 UUID 가져오기
	userUUID := middleware.GetUserUUID(c)
	if userUUID == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// 쿼리 파라미터에서 chatbot_uuid 가져오기
	chatbotUUID := c.Query("chatbot_uuid")
	if chatbotUUID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "chatbot_uuid is required"})
	}

	// SSE 매니저 가져오기
	sseManager := middleware.GetSSEManager()

	// 새로운 세션 생성
	session, err := sseManager.CreateSession(userUUID, chatbotUUID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// SSE 헤더 설정
	middleware.SSEHeaders(c)

	// SSE 스트림 시작
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		log.Printf("SSE 스트림 시작 - 세션: %s", session.SessionID)

		// 연결 유지를 위한 heartbeat
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		// 세션이 활성화되어 있는 동안 메시지 수신 대기
		log.Printf("SSE 루프 시작 - 세션: %s, IsActive: %v", session.SessionID, session.IsActive)
		for session.IsActive {
			select {
			case message := <-session.Channel:
				// 메시지 전송 (봇 응답 등)
				log.Printf("메시지 수신 - 세션: %s, 메시지: %s", session.SessionID, message)
				data := fmt.Sprintf("data: %s\n\n", message)
				_, err := w.Write([]byte(data))
				if err != nil {
					log.Printf("메시지 전송 실패 - 세션: %s, 에러: %v", session.SessionID, err)
					sseManager.DeleteSession(session.SessionID)
					return
				}
				w.Flush()

			case <-ticker.C:
				// heartbeat 전송
				log.Printf("heartbeat 전송 - 세션: %s", session.SessionID)
				heartbeat := fmt.Sprintf("data: %s\n\n", "heartbeat")
				_, err := w.Write([]byte(heartbeat))
				if err != nil {
					log.Printf("heartbeat 전송 실패 - 세션: %s, 에러: %v", session.SessionID, err)
					sseManager.DeleteSession(session.SessionID)
					return
				}
				w.Flush()

			case <-session.Done:
				// SSE Manager에서 전송한 종료 신호
				log.Printf("SSE Manager에서 세션 종료 신호 수신 - 세션: %s", session.SessionID)
				return
			}
			// 루프 상태 로깅
			log.Printf("루프 반복 - 세션: %s, IsActive: %v", session.SessionID, session.IsActive)
		}

		log.Printf("SSE 스트림 종료 - 세션: %s", session.SessionID)
	})

	return nil
}
