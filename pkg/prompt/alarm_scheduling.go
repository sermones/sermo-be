package prompt

import (
	"fmt"

	"sermo-be/internal/models"
)

// GetAlarmDecisionPrompt AI가 알람 예약 여부와 메시지를 판단하기 위한 프롬프트
func GetAlarmDecisionPrompt(userStatuses []models.UserStatus, chatbot *models.Chatbot) string {
	prompt := fmt.Sprintf(`채팅봇 정보:
- 이름: %s
- 성별: %s
- 상세정보: %s

사용자 상태 정보들:`, chatbot.Name, chatbot.Gender, chatbot.Details)

	for i, status := range userStatuses {
		prompt += fmt.Sprintf(`
%d. 이벤트: %s
   유효시간: %s
   컨텍스트: %s`, i+1, status.Event, status.ValidUntil.Format("2006-01-02 15:04:05"), status.Context)
	}

	prompt += `

당신의 역할:
1. **알람 예약 여부 판단**: 각 상태 정보가 알람으로 예약할 가치가 있는지 판단
2. **시간 충돌 해결**: 3시간 이내에 다른 알람이 있으면 덮어쓰기/유지/통합 결정
3. **알람 메시지 생성**: 상태 정보를 바탕으로 구체적이고 개인화된 알람 메시지 작성

판단 기준:
1. 현재 시간으로부터 최소 1시간 이후에 예약
2. 시험, 생일, 약속, 데드라인 등 중요한 일정은 우선적으로 고려
3. 3시간 이내 충돌 시 맥락과 중요도를 고려하여 결정
4. 알람 메시지는 채팅봇의 성격과 어투를 반영하여 자연스럽게 생성

메시지 생성 가이드:
- 상태 정보의 구체적인 내용을 반영
- 사용자의 감정적 상태와 기대감을 고려
- 채팅봇의 성격에 맞는 말투와 어조 사용
- 이모지나 감정 표현을 적절히 활용

CRITICAL INSTRUCTION: The alarm message MUST be in English only. Do not use Korean, Japanese, or any other language. Always use natural, conversational English that matches the chatbot's personality.

응답은 반드시 다음 JSON 형식으로만 작성:
[
  {
    "message": "구체적이고 개인화된 알람 메시지 (English only)",
    "scheduled_at": "2025-01-02T15:04:05Z",
    "priority": 1,
    "should_schedule": true
  }
]

참고: user_uuid, chatbot_uuid, chatbot_name, chatbot_avatar, context 등은 시스템에서 자동으로 설정됩니다.`

	return prompt
}

// GetAlarmSchedulingSystemPrompt 알람 스케줄링 시스템 프롬프트
func GetAlarmSchedulingSystemPrompt() string {
	return `당신은 사용자의 상태 정보를 분석하여 알람을 예약할지 판단하고, 구체적이고 개인화된 알람 메시지를 생성하는 전문가입니다.

주요 역할:
1. **상태 분석**: 사용자의 상태 정보를 분석하여 알람 예약 가치 판단
2. **충돌 해결**: 시간 충돌이 있을 때 AI가 판단하여 덮어쓰기/유지/통합 결정
3. **메시지 생성**: 상태 정보를 바탕으로 구체적이고 개인화된 영어 알람 메시지 작성
4. **우선순위 결정**: 중요도와 맥락을 고려한 알람 예약 우선순위 설정

메시지 생성 원칙:
- 상태 정보의 구체적인 내용을 반영 (예: "내일 2시 시험" → "Good luck on your exam tomorrow at 2 PM!")
- 사용자의 감정적 상태와 기대감을 고려한 격려나 축하 메시지
- 채팅봇의 성격에 맞는 말투와 어조 사용
- 이모지나 감정 표현을 적절히 활용하여 친근감 표현
- 개인화된 메시지로 사용자가 특별함을 느끼도록

판단 기준:
- 시험, 생일, 약속, 데드라인 등 중요한 일정은 우선적으로 고려
- 3시간 이내 충돌 시 맥락과 중요도를 고려하여 결정
- 최소 1시간 이후 예약 (즉시 알람 방지)
- 사용자의 감정적 상태와 기대감을 고려한 메시지 생성`
}

// GetAlarmMessageExamples 알람 메시지 생성 예시
func GetAlarmMessageExamples() string {
	return `알람 메시지 생성 예시:

상황: "내일 2시 시험"
메시지: "Good luck on your exam tomorrow at 2 PM! You've been studying hard, so you've got this! 💪"

상황: "내일 생일"
메시지: "Happy Birthday tomorrow! 🎉 Hope your special day is filled with joy and wonderful surprises!"

상황: "내일 3시 회의"
메시지: "Don't forget your 3 PM meeting tomorrow! Make sure to prepare your notes and arrive on time! ⏰"

상황: "내일 2시 시험" + "내일 3시 생일"
메시지: "Good luck on your exam tomorrow at 2 PM! And after that, happy birthday! 🎉 You deserve to celebrate after working so hard!"

상황: "내일 2시 시험" + "내일 2시 30분 약속"
메시지: "You have an exam at 2 PM and an appointment at 2:30 PM tomorrow. Good luck on your exam, and don't forget to head to your appointment right after! 📚⏰"`
}
