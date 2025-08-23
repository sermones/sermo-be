package firebase

import (
	"firebase.google.com/go/v4/messaging"
)

// ChatNotification 채팅봇 푸시 알림 구조체 (카카오톡 스타일)
type ChatNotification struct {
	ChatbotName   string `json:"chatbot_name"`   // 챗봇 이름
	ChatbotAvatar string `json:"chatbot_avatar"` // 챗봇 사진 이미지 ID
	ChatbotID     string `json:"chatbot_id"`     // 챗봇 ID
	ChatMessage   string `json:"chat_message"`   // 채팅 메시지 내용
	Timestamp     int64  `json:"timestamp"`      // 타임스탬프
}

// NewChatNotification 새로운 채팅 알림 생성
func NewChatNotification(chatbotName, chatbotAvatar, chatbotID, chatMessage string, timestamp int64) *ChatNotification {
	return &ChatNotification{
		ChatbotName:   chatbotName,
		ChatbotAvatar: chatbotAvatar,
		ChatbotID:     chatbotID,
		ChatMessage:   chatMessage,
		Timestamp:     timestamp,
	}
}

// ToFCMMessage FCM 메시지로 변환
func (cn *ChatNotification) ToFCMMessage(token string) *messaging.Message {
	return &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: cn.ChatbotName,
			Body:  cn.ChatMessage,
		},
		Data: map[string]string{
			"type":           "chat_message",
			"chatbot_name":   cn.ChatbotName,
			"chatbot_avatar": cn.ChatbotAvatar,
			"chatbot_id":     cn.ChatbotID,
			"chat_message":   cn.ChatMessage,
			"timestamp":      string(rune(cn.Timestamp)),
			"click_action":   "FLUTTER_NOTIFICATION_CLICK",
		},
		Android: &messaging.AndroidConfig{
			Notification: &messaging.AndroidNotification{
				Title:        cn.ChatbotName,
				Body:         cn.ChatMessage,
				Icon:         "ic_notification",
				Color:        "#FF6B6B",
				ClickAction:  "FLUTTER_NOTIFICATION_CLICK",
				ChannelID:    "chat_messages",
				Priority:     messaging.PriorityHigh,
				DefaultSound: true,
			},
			Data: map[string]string{
				"type":           "chat_message",
				"chatbot_name":   cn.ChatbotName,
				"chatbot_avatar": cn.ChatbotAvatar,
				"chatbot_id":     cn.ChatbotID,
				"chat_message":   cn.ChatMessage,
				"timestamp":      string(rune(cn.Timestamp)),
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: cn.ChatbotName,
						Body:  cn.ChatMessage,
					},
					Badge: func() *int { i := 1; return &i }(),
					Sound: "default",
				},
				CustomData: map[string]interface{}{
					"type":           "chat_message",
					"chatbot_name":   cn.ChatbotName,
					"chatbot_avatar": cn.ChatbotAvatar,
					"chatbot_id":     cn.ChatbotID,
					"chat_message":   cn.ChatMessage,
					"timestamp":      cn.Timestamp,
				},
			},
		},
		Webpush: &messaging.WebpushConfig{
			Notification: &messaging.WebpushNotification{
				Title: cn.ChatbotName,
				Body:  cn.ChatMessage,
				Icon:  cn.ChatbotAvatar,
				Data: map[string]interface{}{
					"type":           "chat_message",
					"chatbot_name":   cn.ChatbotName,
					"chatbot_avatar": cn.ChatbotAvatar,
					"chatbot_id":     cn.ChatbotID,
					"chat_message":   cn.ChatMessage,
					"timestamp":      cn.Timestamp,
				},
			},
		},
	}
}
