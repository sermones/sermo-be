package chat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"sermo-be/internal/core/chat"
	"sermo-be/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// checkConnectionStatus 클라이언트 연결 상태를 확인합니다
func checkConnectionStatus(w *bufio.Writer) error {
	// 연결 상태 확인을 위한 테스트 메시지 전송
	testMessage := "data: {\"type\":\"connection_check\",\"timestamp\":\"" + time.Now().Format(time.RFC3339) + "\"}\n\n"

	// 버퍼에 메시지 쓰기 시도
	if _, err := w.Write([]byte(testMessage)); err != nil {
		return fmt.Errorf("연결 상태 체크 실패: %w", err)
	}

	// 버퍼 플러시 시도
	if err := w.Flush(); err != nil {
		return fmt.Errorf("버퍼 플러시 실패: %w", err)
	}

	return nil
}

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

	// 봇 고루틴 시작
	botGoroutine := chat.GetBotGoroutine()

	// OpenAI 클라이언트 가져오기
	openaiClient := middleware.GetOpenAIClient(c)
	if openaiClient == nil {
		return c.Status(500).JSON(fiber.Map{"error": "OpenAI service unavailable"})
	}

	botChannel := botGoroutine.StartBotGoroutine(session, openaiClient)

	// SSE 스트림 시작
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		log.Printf("SSE 스트림 시작 - 세션: %s", session.SessionID)

		// 연결 유지를 위한 heartbeat (30초마다)
		heartbeatTicker := time.NewTicker(30 * time.Second)
		defer heartbeatTicker.Stop()

		// 연결 상태 체크 (5초마다)
		connectionCheckTicker := time.NewTicker(5 * time.Second)
		defer connectionCheckTicker.Stop()

		// 클라이언트 메시지와 봇 메시지를 처리
		for session.IsActive {
			select {
			case message := <-session.Channel:
				// 클라이언트에서 온 메시지 처리
				log.Printf("클라이언트 메시지 수신 - 세션: %s, 메시지: %s", session.SessionID, message)

				// SSE 메시지에서 "data: " 접두사 제거
				jsonData := message
				if len(message) > 6 && message[:6] == "data: " {
					jsonData = message[6:]
				}

				// 메시지 타입 확인
				var baseMessage struct {
					Type string `json:"type"`
				}
				if err := json.Unmarshal([]byte(jsonData), &baseMessage); err != nil {
					log.Printf("메시지 타입 파싱 실패: %v", err)
					continue
				}

				// 사용자 메시지와 onkeyboard 이벤트를 봇 채널로 전달
				if baseMessage.Type == "user" || baseMessage.Type == "onkeyboard" {
					log.Printf("%s 이벤트를 봇 채널로 전달 - 세션: %s", baseMessage.Type, session.SessionID)
					select {
					case botChannel <- message:
						// 전달 성공
					default:
						log.Printf("봇 채널이 가득 참 - 세션: %s", session.SessionID)
					}
				}

				// 타이핑 이벤트는 클라이언트에만 전송 (봇 채널로 전달하지 않음)
				if baseMessage.Type == "bot_typing" {
					log.Printf("봇 타이핑 이벤트 클라이언트 전송 - 세션: %s", session.SessionID)
				}

				// 클라이언트에 메시지 전송 (echo)
				_, err := w.Write([]byte(message))
				if err != nil {
					log.Printf("클라이언트 메시지 전송 실패 - 세션: %s, 에러: %v", session.SessionID, err)
					sseManager.DeleteSession(session.SessionID)
					return
				}
				w.Flush()

			case <-heartbeatTicker.C:
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

			case <-connectionCheckTicker.C:
				// 5초마다 연결 상태 체크
				if err := checkConnectionStatus(w); err != nil {
					log.Printf("연결 상태 체크 실패 - 세션: %s, 에러: %v", session.SessionID, err)
					sseManager.DeleteSession(session.SessionID)
					return
				}

			case <-session.Done:
				// SSE Manager에서 전송한 종료 신호
				log.Printf("SSE Manager에서 세션 종료 신호 수신 - 세션: %s", session.SessionID)
				return
			}
		}

		log.Printf("SSE 스트림 종료 - 세션: %s", session.SessionID)
	})

	return nil
}
