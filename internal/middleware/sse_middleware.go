package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// SSESession SSE 세션 정보
type SSESession struct {
	SessionID   string
	UserUUID    string
	ChatbotUUID string
	Channel     chan string
	CreatedAt   time.Time
	IsActive    bool
}

// SSEManager SSE 세션 관리자
type SSEManager struct {
	sessions    map[string]*SSESession
	mutex       sync.RWMutex
	maxSessions int
}

// NewSSEManager 새로운 SSE 매니저 생성
func NewSSEManager(maxSessions int) *SSEManager {
	return &SSEManager{
		sessions:    make(map[string]*SSESession),
		maxSessions: maxSessions,
	}
}

// CreateSession 새로운 SSE 세션 생성
func (sm *SSEManager) CreateSession(userUUID, chatbotUUID string) (*SSESession, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 최대 세션 수 확인
	if len(sm.sessions) >= sm.maxSessions {
		return nil, fiber.NewError(fiber.StatusServiceUnavailable, "Maximum number of sessions reached")
	}

	// 기존 활성 세션 확인 (한 유저당 하나의 채팅봇과 하나의 세션만 허용)
	for _, session := range sm.sessions {
		if session.UserUUID == userUUID && session.ChatbotUUID == chatbotUUID && session.IsActive {
			return nil, fiber.NewError(fiber.StatusConflict, "Active session already exists for this user and chatbot")
		}
	}

	sessionID := uuid.New().String()
	session := &SSESession{
		SessionID:   sessionID,
		UserUUID:    userUUID,
		ChatbotUUID: chatbotUUID,
		Channel:     make(chan string, 100), // 버퍼 크기 100
		CreatedAt:   time.Now(),
		IsActive:    true,
	}

	sm.sessions[sessionID] = session
	return session, nil
}

// GetSession 세션 조회
func (sm *SSEManager) GetSession(sessionID string) (*SSESession, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]
	return session, exists
}

// StopSession 세션 중단
func (sm *SSEManager) StopSession(sessionID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Session not found")
	}

	if !session.IsActive {
		return fiber.NewError(fiber.StatusBadRequest, "Session already stopped")
	}

	// 세션 비활성화
	session.IsActive = false
	close(session.Channel)

	// 세션 제거
	delete(sm.sessions, sessionID)

	return nil
}

// SendMessage 세션에 메시지 전송
func (sm *SSEManager) SendMessage(sessionID, message string) error {
	sm.mutex.RLock()
	session, exists := sm.sessions[sessionID]
	sm.mutex.RUnlock()

	if !exists {
		return fiber.NewError(fiber.StatusNotFound, "Session not found")
	}

	if !session.IsActive {
		return fiber.NewError(fiber.StatusBadRequest, "Session is not active")
	}

	select {
	case session.Channel <- message:
		return nil
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to send message to session")
	}
}

// GetActiveSessionsCount 활성 세션 수 조회
func (sm *SSEManager) GetActiveSessionsCount() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return len(sm.sessions)
}

// GetUserSessions 사용자의 활성 세션 조회
func (sm *SSEManager) GetUserSessions(userUUID string) []*SSESession {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var userSessions []*SSESession
	for _, session := range sm.sessions {
		if session.UserUUID == userUUID && session.IsActive {
			userSessions = append(userSessions, session)
		}
	}
	return userSessions
}

// CleanupInactiveSessions 비활성 세션 정리
func (sm *SSEManager) CleanupInactiveSessions() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	for sessionID, session := range sm.sessions {
		// 30분 이상 비활성인 세션 정리
		if now.Sub(session.CreatedAt) > 30*time.Minute {
			session.IsActive = false
			close(session.Channel)
			delete(sm.sessions, sessionID)
		}
	}
}

// 전역 SSE 매니저 인스턴스
var globalSSEManager = NewSSEManager(20) // 최대 20개 세션

// GetSSEManager 전역 SSE 매니저 반환
func GetSSEManager() *SSEManager {
	return globalSSEManager
}

// SSEHeaders SSE 응답 헤더 설정
func SSEHeaders(c *fiber.Ctx) {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Access-Control-Allow-Headers", "Cache-Control")
}
