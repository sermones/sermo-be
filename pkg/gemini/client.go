package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ImageClient Gemini 이미지 생성 API 클라이언트
type ImageClient struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// Config Gemini 클라이언트 설정
type Config struct {
	APIKey string
}

// ImageGenerationRequest 이미지 생성 요청
type ImageGenerationRequest struct {
	Prompt string `json:"prompt"`
}

// Gemini API 응답 구조 (Google 공식 문서 기반)
type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

type Content struct {
	Parts []Part `json:"parts"`
}

type Part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inlineData,omitempty"`
}

type InlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

// ImageGenerationResponse 이미지 생성 응답 (내부 사용)
type ImageGenerationResponse struct {
	Images []ImageData `json:"images"`
}

// ImageData 이미지 데이터
type ImageData struct {
	Data string `json:"data"` // base64 인코딩된 이미지 데이터
}

// NewImageClient 새로운 Gemini 이미지 생성 클라이언트 생성
func NewImageClient(cfg *Config) (*ImageClient, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	return &ImageClient{
		apiKey: cfg.APIKey,
		httpClient: &http.Client{
			Timeout: 180 * time.Second, // 이미지 생성은 시간이 걸릴 수 있음
		},
		baseURL: "https://generativelanguage.googleapis.com/v1beta/models",
	}, nil
}

// GenerateImage 텍스트 프롬프트로 이미지 생성 (Google 공식 문서 구조)
func (c *ImageClient) GenerateImage(ctx context.Context, prompt string) (*ImageGenerationResponse, error) {
	if prompt == "" {
		return nil, fmt.Errorf("prompt is required")
	}

	// Google 공식 문서에 맞는 요청 구조
	requestData := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": prompt,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"responseModalities": []string{"TEXT", "IMAGE"},
			"temperature":        0.7,
			"topK":               40,
			"topP":               0.95,
			"maxOutputTokens":    2048,
		},
	}

	// JSON 인코딩
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// HTTP 요청 생성 (올바른 모델명 사용)
	url := fmt.Sprintf("%s/gemini-2.0-flash-preview-image-generation:generateContent?key=%s", c.baseURL, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// HTTP 요청 실행
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// 응답 상태 확인
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 응답 본문 읽기
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 디버깅: 응답 내용 출력
	fmt.Printf("Gemini API Response: %s\n", string(body))

	// Gemini API 응답 파싱
	var geminiResponse GeminiResponse
	if err := json.Unmarshal(body, &geminiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode Gemini response: %w", err)
	}

	// 이미지 데이터 추출 (Google 공식 문서 구조에 맞게)
	var images []ImageData
	if len(geminiResponse.Candidates) > 0 {
		for _, part := range geminiResponse.Candidates[0].Content.Parts {
			if part.InlineData != nil && part.InlineData.Data != "" {
				images = append(images, ImageData{
					Data: part.InlineData.Data,
				})
			}
		}
	}

	// 디버깅: 추출된 이미지 정보
	fmt.Printf("Extracted %d images from response\n", len(images))

	// 내부 응답 구조로 변환
	response := ImageGenerationResponse{
		Images: images,
	}

	return &response, nil
}

// GenerateImageWithStyle 특정 스타일로 이미지 생성
func (c *ImageClient) GenerateImageWithStyle(ctx context.Context, prompt, style string) (*ImageGenerationResponse, error) {
	// 스타일을 프롬프트에 추가하여 더 구체적인 이미지 생성
	enhancedPrompt := fmt.Sprintf("%s, style: %s. Please generate an image for this prompt.", prompt, style)
	return c.GenerateImage(ctx, enhancedPrompt)
}

// GenerateImageWithSize 특정 크기로 이미지 생성
func (c *ImageClient) GenerateImageWithSize(ctx context.Context, prompt, size string) (*ImageGenerationResponse, error) {
	// 크기 정보를 프롬프트에 추가
	enhancedPrompt := fmt.Sprintf("%s, size: %s. Please generate an image for this prompt.", prompt, size)
	return c.GenerateImage(ctx, enhancedPrompt)
}

// Close 클라이언트 리소스 정리
func (c *ImageClient) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}
