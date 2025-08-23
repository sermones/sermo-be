package chat

import (
	"context"
	"log"
	"time"

	"sermo-be/internal/core/status"
	"sermo-be/internal/middleware"
	"sermo-be/pkg/openai"
	"sermo-be/pkg/prompt"
)

// StatusGenerator 상태 정보 수집 및 저장을 담당하는 구조체
type StatusGenerator struct {
	statusService *status.StatusService
}

// NewStatusGenerator 새로운 StatusGenerator 생성
func NewStatusGenerator() *StatusGenerator {
	return &StatusGenerator{
		statusService: status.GetStatusService(),
	}
}

// ExtractAndSaveStatus 사용자 메시지에서 상태 정보 추출 및 저장
func (sg *StatusGenerator) ExtractAndSaveStatus(session *middleware.SSESession, userMessage string, openaiClient *openai.Client) {
	// 상태 정보 추출
	statusResult := sg.extractUserStatus(userMessage, openaiClient)
	if statusResult == nil {
		return
	}

	// 저장이 필요한 경우에만 저장
	if statusResult.NeedsSave {
		go sg.saveUserStatus(session, statusResult.Event, statusResult.ValidUntil, statusResult.Context)
	}
}

// extractUserStatus 사용자 메시지에서 상태 정보 추출
func (sg *StatusGenerator) extractUserStatus(userMessage string, openaiClient *openai.Client) *status.StatusExtractionResult {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 상태 정보 추출을 위한 프롬프트 구성
	statusPrompt := prompt.GetStatusExtractionPrompt()
	statusMessages := []openai.ChatMessage{
		{
			Role:    "system",
			Content: statusPrompt,
		},
		{
			Role:    "user",
			Content: userMessage,
		},
	}

	// 상태 정보 추출 요청
	statusResponse, err := openaiClient.ChatCompletion(ctx, statusMessages)
	if err != nil {
		return nil
	}

	// 상태 정보 파싱
	statusResult, err := sg.statusService.ParseStatusExtractionResult(statusResponse.Message.Content)
	if err != nil {
		return nil
	}

	return statusResult
}

// saveUserStatus 사용자 상태 정보 저장
func (sg *StatusGenerator) saveUserStatus(session *middleware.SSESession, event string, validUntil time.Time, context string) {
	err := sg.statusService.SaveUserStatus(
		session.UserUUID,
		session.ChatbotUUID,
		event,
		validUntil,
		context,
	)
	if err != nil {
		log.Printf("상태 정보 저장 실패 - 세션: %s, 에러: %v", session.SessionID, err)
	}
}
