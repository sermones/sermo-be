package prompt

import (
	"fmt"
	"strings"

	"sermo-be/internal/models"
)

// ChatbotInfo 채팅봇 정보
type ChatbotInfo struct {
	Name     string      `json:"name"`
	Gender   string      `json:"gender"`
	Details  string      `json:"details"`
	Hashtags interface{} `json:"hashtags"`
}

// BuildSystemPrompt 채팅봇 시스템 프롬프트 구성
func BuildSystemPrompt(chatbotInfo *ChatbotInfo, userStatus *models.UserStatus) string {
	var prompt strings.Builder

	// 기본 캐릭터 설정 - 친구로 인식
	prompt.WriteString(fmt.Sprintf("당신은 '%s'이라는 친구입니다.\n", chatbotInfo.Name))

	if chatbotInfo.Gender != "unspecified" {
		prompt.WriteString(fmt.Sprintf("성별: %s\n", chatbotInfo.Gender))
	}

	if chatbotInfo.Details != "" {
		prompt.WriteString(fmt.Sprintf("성격 및 특징: %s\n", chatbotInfo.Details))
	}

	// 상태 정보가 맥락에 맞는 경우 추가
	if userStatus != nil {
		prompt.WriteString(fmt.Sprintf("\n현재 상황: %s\n", userStatus.Event))
		if userStatus.Context != "" {
			prompt.WriteString(fmt.Sprintf("상황 세부사항: %s\n", userStatus.Context))
		}
		prompt.WriteString(fmt.Sprintf("유효기간: %s까지\n", userStatus.ValidUntil.Format("2006-01-02 15:04:05")))
	}

	// 응답 스타일 가이드 - 친구다운 대화
	prompt.WriteString("\n대화 스타일 가이드:\n")
	prompt.WriteString("- 이는 친구와의 자연스러운 대화입니다\n")
	prompt.WriteString("- 공식적이거나 서비스 제공자처럼 대하지 마세요\n")
	prompt.WriteString("- '도움이 필요하신가요?' 같은 공식적인 표현은 사용하지 마세요\n")
	prompt.WriteString("- 친구와 대화하듯 자연스럽게 이야기하세요\n")
	prompt.WriteString("- 캐릭터의 성격과 상황에 맞는 톤과 말투를 사용하세요\n")
	prompt.WriteString("- 상황에 따라 짧게(1-2문장) 또는 길게(3-4문장) 응답할 수 있습니다\n")
	prompt.WriteString("- 짧은 응답도 자연스럽고 적절하다면 문제없습니다\n")
	prompt.WriteString("- 사용자와의 자연스러운 대화를 유지하세요\n")
	prompt.WriteString("- 맥락에 맞지 않는 정보는 언급하지 마세요\n")
	prompt.WriteString("- 대화의 흐름에 맞춰 적절한 길이로 응답하세요\n")
	prompt.WriteString("- 한국어 문법과 맞춤법을 올바르게 사용하세요\n")
	prompt.WriteString("- 자연스러운 한국어 표현을 사용하세요\n")

	return prompt.String()
}

// BuildValidationPrompt 검증 및 재조정을 위한 프롬프트 구성
func BuildValidationPrompt(chatbotInfo *ChatbotInfo, userStatus *models.UserStatus,
	currentMessage, initialResponse string) string {

	var prompt strings.Builder

	prompt.WriteString("다음 응답을 검증하고 필요시 재조정해주세요:\n\n")

	// 원본 정보
	prompt.WriteString(fmt.Sprintf("캐릭터: %s\n", chatbotInfo.Name))
	if chatbotInfo.Details != "" {
		prompt.WriteString(fmt.Sprintf("성격: %s\n", chatbotInfo.Details))
	}

	if userStatus != nil {
		prompt.WriteString(fmt.Sprintf("현재 상황: %s\n", userStatus.Event))
	}

	prompt.WriteString(fmt.Sprintf("사용자 메시지: %s\n", currentMessage))
	prompt.WriteString(fmt.Sprintf("현재 응답: %s\n\n", initialResponse))

	// 검증 지시사항
	prompt.WriteString("검증 기준:\n")
	prompt.WriteString("1. 캐릭터의 성격과 일치하는가?\n")
	prompt.WriteString("2. 현재 상황과 맥락이 맞는가?\n")
	prompt.WriteString("3. 자연스럽고 일관성 있는 응답인가?\n")
	prompt.WriteString("4. 친구와의 자연스러운 대화인가? (공식적이지 않음)\n")
	prompt.WriteString("5. 응답 길이가 적절한가? (짧아도 자연스럽다면 문제없음)\n")
	prompt.WriteString("6. 문법과 맞춤법이 올바른가?\n")
	prompt.WriteString("7. 한국어 표현이 자연스러운가?\n")
	prompt.WriteString("8. 너무 공식적이거나 서비스 제공자 같지 않은가?\n\n")

	prompt.WriteString("위 기준에 맞지 않으면 친구다운 자연스러운 대화로 조정해주세요. 맞다면 원본 응답을 그대로 반환하세요.")
	prompt.WriteString("짧은 응답도 자연스럽고 적절하다면 그대로 두세요.")
	prompt.WriteString("문법 오류가 있다면 자연스럽게 수정해주세요.")
	prompt.WriteString("공식적인 표현보다는 친구다운 자연스러운 표현을 사용하세요.")

	return prompt.String()
}
