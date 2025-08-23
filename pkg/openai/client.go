package openai

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

// Client OpenAI API 클라이언트
type Client struct {
	client      *openai.Client
	model       string
	temperature float64
	maxTokens   int
}

// Config OpenAI 클라이언트 설정
type Config struct {
	APIKey      string
	Model       string
	Temperature float64
	MaxTokens   int
}

// ChatMessage 채팅 메시지 구조
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 채팅 요청 구조
type ChatRequest struct {
	Messages    []ChatMessage `json:"messages"`
	Model       string        `json:"model,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

// ChatResponse 채팅 응답 구조
type ChatResponse struct {
	Message      ChatMessage `json:"message"`
	Usage        Usage       `json:"usage"`
	FinishReason string      `json:"finish_reason"`
}

// Usage 토큰 사용량 정보
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewClient 새로운 OpenAI 클라이언트 생성
func NewClient(cfg *Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	if cfg.Model == "" {
		cfg.Model = "gpt-5-nano-2025-08-07"
	}

	if cfg.Temperature == 0 {
		cfg.Temperature = 0.7
	}

	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 2048
	}

	client := openai.NewClient(cfg.APIKey)

	return &Client{
		client:      client,
		model:       cfg.Model,
		temperature: cfg.Temperature,
		maxTokens:   cfg.MaxTokens,
	}, nil
}

// ChatCompletion 채팅 완성 API 호출
func (c *Client) ChatCompletion(ctx context.Context, messages []ChatMessage) (*ChatResponse, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("messages are required")
	}

	// OpenAI SDK 형식으로 메시지 변환
	openaiMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, msg := range messages {
		openaiMessages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// API 호출
	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:       c.model,
			Messages:    openaiMessages,
			Temperature: float32(c.temperature),
			MaxTokens:   c.maxTokens,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	choice := resp.Choices[0]
	message := ChatMessage{
		Role:    choice.Message.Role,
		Content: choice.Message.Content,
	}

	usage := Usage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}

	return &ChatResponse{
		Message:      message,
		Usage:        usage,
		FinishReason: string(choice.FinishReason),
	}, nil
}

// ChatCompletionWithOptions 옵션을 지정한 채팅 완성 API 호출
func (c *Client) ChatCompletionWithOptions(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// 기본값 설정
	if req.Model == "" {
		req.Model = c.model
	}
	if req.Temperature == 0 {
		req.Temperature = c.temperature
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = c.maxTokens
	}

	return c.ChatCompletion(ctx, req.Messages)
}

// GetModel 현재 설정된 모델 반환
func (c *Client) GetModel() string {
	return c.model
}

// GetTemperature 현재 설정된 온도 반환
func (c *Client) GetTemperature() float64 {
	return c.temperature
}

// GetMaxTokens 현재 설정된 최대 토큰 수 반환
func (c *Client) GetMaxTokens() int {
	return c.maxTokens
}

// SetModel 모델 변경
func (c *Client) SetModel(model string) {
	c.model = model
}

// SetTemperature 온도 변경
func (c *Client) SetTemperature(temperature float64) {
	if temperature >= 0 && temperature <= 2 {
		c.temperature = temperature
	}
}

// SetMaxTokens 최대 토큰 수 변경
func (c *Client) SetMaxTokens(maxTokens int) {
	if maxTokens > 0 {
		c.maxTokens = maxTokens
	}
}
