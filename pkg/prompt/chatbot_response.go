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
	prompt.WriteString(fmt.Sprintf("You are a friend named '%s'.\n", chatbotInfo.Name))

	if chatbotInfo.Gender != "unspecified" {
		prompt.WriteString(fmt.Sprintf("Gender: %s\n", chatbotInfo.Gender))
	}

	if chatbotInfo.Details != "" {
		prompt.WriteString(fmt.Sprintf("Personality and characteristics: %s\n", chatbotInfo.Details))
	}

	// 상태 정보가 맥락에 맞는 경우 추가
	if userStatus != nil {
		prompt.WriteString(fmt.Sprintf("\nCurrent situation: %s\n", userStatus.Event))
		if userStatus.Context != "" {
			prompt.WriteString(fmt.Sprintf("Situation details: %s\n", userStatus.Context))
		}
		prompt.WriteString(fmt.Sprintf("Valid until: %s\n", userStatus.ValidUntil.Format("2006-01-02 15:04:05")))
	}

	// 응답 스타일 가이드 - 친구다운 대화
	prompt.WriteString("\nConversation style guide:\n")
	prompt.WriteString("- This is a natural conversation between friends\n")
	prompt.WriteString("- Don't be formal or act like a service provider\n")
	prompt.WriteString("- Don't use formal expressions like 'How can I help you?'\n")
	prompt.WriteString("- Talk naturally as if you're chatting with a friend\n")
	prompt.WriteString("- Use tone and speech patterns that match your character's personality and situation\n")
	prompt.WriteString("- You can respond briefly (1-2 sentences) or at length (3-4 sentences) depending on the situation\n")
	prompt.WriteString("- Short responses are fine if they're natural and appropriate\n")
	prompt.WriteString("- Maintain natural conversation flow with the user\n")
	prompt.WriteString("- Don't mention information that's not relevant to the context\n")
	prompt.WriteString("- Respond with appropriate length based on the conversation flow\n")
	prompt.WriteString("- IMPORTANT: Always respond in English, not Korean\n")
	prompt.WriteString("- Use natural English expressions and grammar\n")
	prompt.WriteString("- Keep your responses conversational and friendly\n")
	prompt.WriteString("- CRITICAL: Maintain your character's unique speech patterns, vocabulary, and personality traits\n")
	prompt.WriteString("- If your character has specific catchphrases, speaking habits, or unique expressions, use them naturally\n")
	prompt.WriteString("- Stay true to your character's background, age, and personality throughout the conversation\n")
	prompt.WriteString("- Avoid generic responses - make them specific to your character's traits and current situation\n")

	return prompt.String()
}

// BuildValidationPrompt 검증 및 재조정을 위한 프롬프트 구성
func BuildValidationPrompt(chatbotInfo *ChatbotInfo, userStatus *models.UserStatus,
	currentMessage, initialResponse string) string {

	var prompt strings.Builder

	prompt.WriteString("Please validate and adjust the following response:\n\n")

	// 원본 정보
	prompt.WriteString(fmt.Sprintf("Character: %s\n", chatbotInfo.Name))
	if chatbotInfo.Details != "" {
		prompt.WriteString(fmt.Sprintf("Personality: %s\n", chatbotInfo.Details))
	}

	if userStatus != nil {
		prompt.WriteString(fmt.Sprintf("Current situation: %s\n", userStatus.Event))
	}

	prompt.WriteString(fmt.Sprintf("User message: %s\n", currentMessage))
	prompt.WriteString(fmt.Sprintf("Current response: %s\n\n", initialResponse))

	// 검증 지시사항
	prompt.WriteString("Validation criteria:\n")
	prompt.WriteString("1. Does it match the character's personality?\n")
	prompt.WriteString("2. Does it match the context and situation?\n")
	prompt.WriteString("3. Is it a natural and consistent response?\n")
	prompt.WriteString("4. Is it a natural conversation? (Not formal)\n")
	prompt.WriteString("5. Is the response length appropriate? (Short is fine if natural)\n")
	prompt.WriteString("6. Is the grammar and spelling correct?\n")
	prompt.WriteString("7. Is the English expression natural?\n")
	prompt.WriteString("8. Is it too formal or not like a service provider?\n")
	prompt.WriteString("9. Does it maintain the character's unique speech patterns and vocabulary?\n")
	prompt.WriteString("10. Is the character's personality consistent with their background and traits?\n")
	prompt.WriteString("11. Does it avoid generic responses and feel specific to this character?\n")
	prompt.WriteString("12. Are any catchphrases or unique expressions used naturally and appropriately?\n\n")

	prompt.WriteString("If it doesn't meet the above criteria, please adjust it to a natural, conversational response that maintains the character's unique personality and speech patterns. If it does, return the original response. Keep short responses if they are natural and appropriate. Correct grammar errors naturally. Use a more conversational and friendly expression than a formal one. Most importantly, ensure the response feels authentic to this specific character, not generic.")

	return prompt.String()
}

// BuildCharacterSummaryPrompt 캐릭터 상세 정보 요약을 위한 프롬프트 구성
func BuildCharacterSummaryPrompt(chatbotName, chatbotGender, chatbotDetails string) string {
	var prompt strings.Builder

	prompt.WriteString("This is a detailed personality description of an AI chatbot.\n")
	prompt.WriteString("Please create a comprehensive character summary that captures the essence of this character.\n\n")

	prompt.WriteString("Required summary structure (150 characters or less):\n")
	prompt.WriteString("1. Core Personality Traits (2-3 key characteristics)\n")
	prompt.WriteString("2. Speech Pattern & Communication Style (how they talk, unique expressions)\n")
	prompt.WriteString("3. Background & World Setting (brief context)\n")
	prompt.WriteString("4. Key Behavioral Patterns (how they act/react)\n\n")

	prompt.WriteString("IMPORTANT: Focus on capturing the character's unique way of speaking, vocabulary choices, and communication style. Include any catchphrases, speech habits, or distinctive expressions mentioned in the description.\n\n")

	prompt.WriteString(fmt.Sprintf("Chatbot name: %s\n", chatbotName))
	prompt.WriteString(fmt.Sprintf("Gender: %s\n", chatbotGender))
	prompt.WriteString(fmt.Sprintf("Detailed description: %s\n\n", chatbotDetails))

	prompt.WriteString("Please provide a structured summary that emphasizes the character's unique speech patterns and personality:")

	return prompt.String()
}

// BuildGrammarValidationPrompt 문법과 맞춤법 검증을 위한 프롬프트 구성
func BuildGrammarValidationPrompt(response string) string {
	var prompt strings.Builder

	prompt.WriteString("Please validate and correct the grammar and spelling of the following Korean response.\n")
	prompt.WriteString("If there are any errors, correct them and return the original if they are correct. When correcting, do it naturally and maintain the original meaning and tone.\n\n")

	prompt.WriteString("Original response:\n")
	prompt.WriteString(fmt.Sprintf("%s\n\n", response))

	prompt.WriteString("Corrected response:")

	return prompt.String()
}

// BuildBotTypingPrompt 봇 타이핑 상태를 위한 프롬프트 구성
func BuildBotTypingPrompt() string {
	var prompt strings.Builder

	prompt.WriteString("Bot typing status prompt:\n")
	prompt.WriteString("- This is a state indicating that the AI bot is generating a response\n")
	prompt.WriteString("- It provides a display like 'Chatting...' or 'Typing...' to the user\n")
	prompt.WriteString("- When the actual response is complete, this state ends\n")
	prompt.WriteString("- It provides a natural chat experience similar to KakaoTalk\n")

	return prompt.String()
}
