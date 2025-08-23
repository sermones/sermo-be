package prompt

import (
	"fmt"
	"strings"

	"sermo-be/internal/models"
)

func BuildSummaryPrompt(userStatuses []models.UserStatus, chatbot *models.Chatbot, chatHistory []models.ChatMessage) string {
	prompt := fmt.Sprintf(`Chatbot: %s (%s)

User Status (Priority):
`, chatbot.Name, chatbot.Gender)

	// UserStatus를 우선으로 배치
	for i, status := range userStatuses {
		prompt += fmt.Sprintf(`
%d. Event: %s
   Context: %s
   Valid Until: %s`, i+1, status.Event, status.Context, status.ValidUntil.Format("2006-01-02 15:04:05"))
	}

	if len(chatHistory) > 0 {
		// 채팅 히스토리를 평문으로 변환
		var chatTexts []string
		for _, msg := range chatHistory {
			chatTexts = append(chatTexts, msg.Content)
		}
		chatText := strings.Join(chatTexts, "\n")

		prompt += fmt.Sprintf(`

Recent Chat Summary:
%s`, chatText)
	} else {
		prompt += `

Recent Chat Summary:
No chat history`
	}

	prompt += `

Based on the user status and recent conversation, extract 2 key keywords that best represent the user's current situation and needs.

Keywords should focus on:
1. Most recent and frequent status events
2. User's current emotional or practical needs

Format your response as:
Keywords: [keyword1], [keyword2]

Example:
Keywords: birthday, celebration`

	return prompt
}

// BuildPersonalizedAlarmPrompt 개인화된 알람 메시지 생성을 위한 프롬프트
func BuildPersonalizedAlarmPrompt(keywords []string, userStatuses []models.UserStatus, chatbot *models.Chatbot, user *models.User) string {
	if len(userStatuses) == 0 {
		return "No user status available."
	}

	latestStatus := userStatuses[0]

	prompt := fmt.Sprintf(`Chatbot Character: %s (%s)
Character Details: %s
Character Summary: %s

User Name: %s
Status Event: %s
Status Context: %s

Extracted Keywords: %s

Create a personalized alarm message that:
1. Uses the user's name naturally
2. Reflects the chatbot's personality and speaking style from the summary
3. Incorporates the keywords and status information
4. Sounds friendly and engaging
5. Uses the character's unique speech patterns and vocabulary

Also suggest the best time to send this alarm message based on the event type and user's situation.

Format your response as:
[Your personalized alarm message]

Send Time: YYYY-MM-DD HH:MM:SS (e.g., "2025-01-25 00:00:00" for birthday at midnight, "2025-01-25 13:30:00" for 30 minutes before exam)`,
		chatbot.Name, chatbot.Gender, chatbot.Details, chatbot.GetSummary(),
		user.Nickname, latestStatus.Event, latestStatus.Context,
		strings.Join(keywords, ", "))

	return prompt
}
