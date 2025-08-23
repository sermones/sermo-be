package status

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"sermo-be/internal/models"
	"sermo-be/pkg/database"
)

// StatusExtractionResult 상태 추출 결과
type StatusExtractionResult struct {
	NeedsSave  bool      `json:"needs_save"`
	Event      string    `json:"event,omitempty"`
	ValidUntil time.Time `json:"valid_until,omitempty"`
	Context    string    `json:"context,omitempty"`
}

// StatusService 사용자 상태 정보 관리 서비스
type StatusService struct{}

// SaveUserStatus 사용자 상태 정보 저장
func (s *StatusService) SaveUserStatus(userUUID, chatbotUUID, event string, validUntil time.Time, context string) error {
	userStatus := models.NewUserStatus(userUUID, chatbotUUID, event, validUntil, context)

	if err := database.DB.Create(userStatus).Error; err != nil {
		return fmt.Errorf("사용자 상태 정보 저장 실패: %w", err)
	}

	log.Printf("사용자 상태 정보 저장 완료 - 사용자: %s, 이벤트: %s, 유효시간: %s",
		userUUID, event, validUntil.Format("2006-01-02 15:04:05"))

	return nil
}

// ParseStatusExtractionResult 상태 추출 결과 파싱
func (s *StatusService) ParseStatusExtractionResult(response string) (*StatusExtractionResult, error) {
	var result StatusExtractionResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return nil, fmt.Errorf("상태 추출 결과 파싱 실패: %w", err)
	}
	return &result, nil
}

// GetStatusService 전역 StatusService 반환
func GetStatusService() *StatusService {
	return globalStatusService
}

// 전역 StatusService 인스턴스
var globalStatusService = &StatusService{}
