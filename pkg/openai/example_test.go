package openai

import (
	"context"
	"fmt"
	"log"
)

// ExampleUsage OpenAI 클라이언트 사용 예시
func ExampleUsage() {
	// 설정
	cfg := &Config{
		APIKey:      "your-api-key-here",
		Model:       "gpt-5-nano-2025-08-07",
		Temperature: 0.7,
		MaxTokens:   2048,
	}

	// 클라이언트 생성
	client, err := NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// 컨텍스트 생성
	ctx := context.Background()

	// 채팅 메시지 준비
	messages := []ChatMessage{
		{
			Role:    "system",
			Content: "당신은 도움이 되는 AI 어시스턴트입니다.",
		},
		{
			Role:    "user",
			Content: "안녕하세요! 오늘 날씨가 어때요?",
		},
	}

	// 채팅 완성 요청
	response, err := client.ChatCompletion(ctx, messages)
	if err != nil {
		log.Fatalf("Failed to get chat completion: %v", err)
	}

	// 응답 출력
	fmt.Printf("AI 응답: %s\n", response.Message.Content)
	fmt.Printf("사용된 토큰: %d\n", response.Usage.TotalTokens)
	fmt.Printf("완료 이유: %s\n", response.FinishReason)
}
