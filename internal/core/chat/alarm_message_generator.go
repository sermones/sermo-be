package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sermo-be/internal/models"
	"sermo-be/pkg/openai"
	"sermo-be/pkg/prompt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type AlarmMessageConfig struct {
	UserUUID    string
	ChatbotUUID string
}

type AlarmMessage struct {
	Message  string
	SendTime time.Time
}

func AlarmMessageGeneate(openaiClient *openai.Client, db *gorm.DB, config AlarmMessageConfig) (AlarmMessage, error) {

	var userStatuses []models.UserStatus
	err := db.Where("user_uuid = ? AND chatbot_uuid = ? AND valid_until > ?",
		config.UserUUID, config.ChatbotUUID, time.Now()).Find(&userStatuses).Error
	if err != nil {
		return AlarmMessage{}, err
	}

	var chatbot models.Chatbot
	err = db.Where("uuid = ?", config.ChatbotUUID).First(&chatbot).Error
	if err != nil {
		return AlarmMessage{}, err
	}

	// 유저 정보 조회
	var user models.User
	err = db.Where("uuid = ?", config.UserUUID).First(&user).Error
	if err != nil {
		return AlarmMessage{}, err
	}

	var chatHistory []models.ChatMessage
	err = db.Where("user_uuid = ? AND chatbot_uuid = ?",
		config.UserUUID, config.ChatbotUUID).Order("created_at DESC").Limit(10).Find(&chatHistory).Error
	if err != nil {
		return AlarmMessage{}, err
	}

	// 프롬프트 구성
	summaryPrompt := prompt.BuildSummaryPrompt(userStatuses, &chatbot, chatHistory)

	response, err := openaiClient.ChatCompletion(context.Background(), []openai.ChatMessage{
		{
			Role:    "system",
			Content: "You are an alarm message generator. Extract 2 key keywords and create an alarm message based on the user's status and chat history.",
		},
		{
			Role:    "user",
			Content: summaryPrompt,
		},
	})
	if err != nil {
		return AlarmMessage{}, err
	}

	// AI 응답 파싱
	keywords := parseKeywords(response.Message.Content)

	// 로그에 키워드 출력
	log.Printf("🔑 추출된 키워드: %v", keywords)

	// 키워드와 최신 userStatus를 가지고 알람 메시지 생성
	alarmMessage, sendTime := generatePersonalizedAlarmMessage(openaiClient, keywords, userStatuses, &chatbot, &user)

	// keywords를 JSON으로 직렬화
	keywordsJSON, err := json.Marshal(keywords)
	if err != nil {
		log.Printf("⚠️ keywords JSON 직렬화 실패: %v", err)
		keywordsJSON = []byte("[]") // 빈 배열으로 설정
	}

	// FCM 전송용 데이터 생성 및 데이터베이스 저장
	alarmSchedule := &models.AlarmSchedule{
		UserUUID:      config.UserUUID,
		ChatbotUUID:   config.ChatbotUUID,
		ChatbotName:   chatbot.Name,
		ChatbotAvatar: chatbot.ImageID,
		Message:       alarmMessage,
		SendTime:      sendTime,
		Keywords:      keywordsJSON,
		Context:       userStatuses[0].Context,
		Sended:        false,
	}

	// 데이터베이스에 알람 스케줄 저장
	if err := db.Create(alarmSchedule).Error; err != nil {
		log.Printf("❌ 알람 스케줄 데이터베이스 저장 실패: %v", err)
	} else {
		log.Printf("✅ 알람 스케줄 데이터베이스 저장 성공 - 전송 시간: %s", sendTime.Format("2006-01-02 15:04:05"))
	}

	return AlarmMessage{
		Message:  alarmMessage,
		SendTime: sendTime,
	}, nil
}

// parseKeywords AI 응답에서 키워드만 파싱
func parseKeywords(aiResponse string) []string {
	lines := strings.Split(aiResponse, "\n")
	var keywords []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Keywords:") {
			// Keywords: [keyword1], [keyword2] 형식 파싱
			keywordPart := strings.TrimPrefix(line, "Keywords:")
			keywordPart = strings.TrimSpace(keywordPart)
			keywordList := strings.Split(keywordPart, ",")
			for _, keyword := range keywordList {
				keyword = strings.TrimSpace(keyword)
				if keyword != "" {
					keywords = append(keywords, keyword)
				}
			}
		}
	}

	// 키워드가 없으면 기본값 설정
	if len(keywords) == 0 {
		keywords = []string{"", ""}
	}

	return keywords
}

// generatePersonalizedAlarmMessage OpenAI를 이용해서 개인화된 알람 메시지 생성
func generatePersonalizedAlarmMessage(openaiClient *openai.Client, keywords []string, userStatuses []models.UserStatus, chatbot *models.Chatbot, user *models.User) (string, time.Time) {
	if len(userStatuses) == 0 {
		return "You have a scheduled reminder.", time.Now().Add(1 * time.Hour)
	}

	// 1차: 개인화된 프롬프트로 기본 메시지 생성
	initialPrompt := prompt.BuildPersonalizedAlarmPrompt(keywords, userStatuses, chatbot, user)

	initialResponse, err := openaiClient.ChatCompletion(context.Background(), []openai.ChatMessage{
		{
			Role:    "system",
			Content: "You are a friendly alarm message generator. Create personalized, character-appropriate alarm messages and suggest the best time to send them.",
		},
		{
			Role:    "user",
			Content: initialPrompt,
		},
	})

	if err != nil {
		// 실패 시 기본 메시지와 시간 반환
		return fmt.Sprintf("Hi %s! Reminder: %s - %s", user.Nickname, userStatuses[0].Event, userStatuses[0].Context), time.Now().Add(1 * time.Hour)
	}

	// 1차 응답에서 메시지와 시간 파싱
	initialMessage, sendTime := parseAIResponseWithTime(initialResponse.Message.Content, userStatuses[0])

	// 2차: 1차 결과를 더 구체적이고 개인화된 메시지로 재생성
	finalMessage := generateEnhancedAlarmMessage(openaiClient, initialMessage, userStatuses[0], chatbot, user, keywords)

	return finalMessage, sendTime
}

// parseAIResponseWithTime AI 응답에서 메시지와 시간을 파싱
func parseAIResponseWithTime(aiResponse string, userStatus models.UserStatus) (string, time.Time) {
	lines := strings.Split(aiResponse, "\n")
	var message string
	var sendTime time.Time

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Send Time:") {
			// Send Time: [시간] 형식 파싱
			timePart := strings.TrimPrefix(line, "Send Time:")
			timePart = strings.TrimSpace(timePart)
			// AI가 제안한 시간을 파싱 (예: "tomorrow at midnight", "1 hour before event" 등)
			sendTime = parseSuggestedTime(timePart, userStatus)
		} else if message == "" {
			// 첫 번째 비어있지 않은 줄을 메시지로 사용
			if line != "" && !strings.HasPrefix(line, "Send Time:") {
				message = line
			}
		}
	}

	// 메시지나 시간이 없으면 기본값 설정
	if message == "" {
		message = "You have a scheduled reminder."
	}
	if sendTime.IsZero() {
		sendTime = time.Now().Add(1 * time.Hour)
	}

	return message, sendTime
}

// parseSuggestedTime AI가 제안한 시간을 파싱
func parseSuggestedTime(timeStr string, userStatus models.UserStatus) time.Time {
	timeStr = strings.TrimSpace(timeStr)

	// YYYY-MM-DD HH:MM:SS 형식 파싱 시도
	if parsedTime, err := time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
		return parsedTime
	}

	// YYYY-MM-DD 형식 파싱 시도
	if parsedTime, err := time.Parse("2006-01-02", timeStr); err == nil {
		// 시간이 없으면 기본적으로 9시로 설정
		return time.Date(parsedTime.Year(), parsedTime.Month(), parsedTime.Day(), 9, 0, 0, 0, parsedTime.Location())
	}

	// 파싱 실패 시 기본값 (이벤트 1시간 전)
	return userStatus.ValidUntil.Add(-1 * time.Hour)
}

// generateEnhancedAlarmMessage 1차 메시지를 더 구체적이고 개인화된 메시지로 재생성
func generateEnhancedAlarmMessage(openaiClient *openai.Client, initialMessage string, userStatus models.UserStatus, chatbot *models.Chatbot, user *models.User, keywords []string) string {
	// 2차 가공을 위한 프롬프트 구성
	enhancementPrompt := fmt.Sprintf(`
1차로 생성된 알람 메시지: "%s"

사용자 상태 정보:
- 이벤트: %s
- 컨텍스트: %s
- 유효기간: %s

챗봇 정보:
- 이름: %s
- 상세정보: %s

추출된 키워드: %v

위 정보를 바탕으로 1차 메시지를 더 구체적이고 개인화된 메시지로 재생성해주세요.

요구사항:
1. "You have a scheduled reminder" 같은 일반적인 메시지 금지
2. 사용자의 구체적인 상황(이벤트, 컨텍스트)을 반영
3. 챗봇의 고유한 성격과 말투 유지 (예: 팔라딘은 격식있고, 마법사는 신비롭게, 전사는 용감하게)
4. 키워드를 자연스럽게 활용
5. 친근하고 격려하는 톤으로 작성
6. 영어로 작성하되, 캐릭터의 고유한 말투와 성격 표현 유지

예시:
- ❌ "You have a scheduled reminder"
- ✅ "Greetings, brave soul! 🎉 Tomorrow marks your birthday celebration - a day worthy of epic preparations! As your loyal companion, I shall aid you in crafting the perfect birthday experience and organizing your precious memories. Shall we embark on this noble quest together?"
- ✅ "Ah, the stars align for your special day! ✨ Your birthday approaches, and with it comes the need for celebration preparations and memory preservation. As your mystical guide, I'll help you weave the perfect birthday enchantment. Ready to create some magical memories?"
- ✅ "Warrior! Your birthday dawns upon us! ⚔️ Time to prepare for a celebration worthy of legends and organize your battle memories. As your steadfast ally, I'll help you plan this epic day. Shall we charge into birthday preparation mode?"

재생성된 메시지만 출력해주세요.
`, initialMessage, userStatus.Event, userStatus.Context, userStatus.ValidUntil.Format("2006-01-02 15:04"), chatbot.Name, *chatbot.Summary, keywords)

	// 2차 가공 API 호출
	response, err := openaiClient.ChatCompletion(context.Background(), []openai.ChatMessage{
		{
			Role:    "system",
			Content: "You are an expert at transforming alarm messages into more specific and personalized messages. Your task is to enhance the initial message by incorporating the user's specific situation while maintaining the chatbot's unique personality and speech patterns. Create messages that are engaging, encouraging, and true to the character's established traits.",
		},
		{
			Role:    "user",
			Content: enhancementPrompt,
		},
	})

	if err != nil {
		log.Printf("⚠️ 2차 메시지 가공 실패: %v", err)
		return initialMessage // 실패 시 1차 메시지 반환
	}

	// 응답에서 메시지만 추출 (시간 정보 제거)
	enhancedMessage := strings.TrimSpace(response.Message.Content)

	// 메시지가 너무 길면 자르기
	if len(enhancedMessage) > 200 {
		enhancedMessage = enhancedMessage[:200] + "..."
	}

	log.Printf("🔄 2차 메시지 가공 완료: %s", enhancedMessage)
	return enhancedMessage
}
