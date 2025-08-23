package firebase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sermo-be/internal/config"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// Client Firebase 클라이언트 구조체
type Client struct {
	app       *firebase.App
	messaging *messaging.Client
	projectID string
}

// NewClient 새로운 Firebase 클라이언트 생성
func NewClient(cfg *config.Config) (*Client, error) {
	// 서비스 계정 키 JSON 생성
	serviceAccount := map[string]interface{}{
		"type":                        "service_account",
		"project_id":                  cfg.Firebase.ProjectID,
		"private_key_id":              cfg.Firebase.PrivateKeyID,
		"private_key":                 cfg.Firebase.PrivateKey,
		"client_email":                cfg.Firebase.ClientEmail,
		"client_id":                   cfg.Firebase.ClientID,
		"auth_uri":                    cfg.Firebase.AuthURI,
		"token_uri":                   cfg.Firebase.TokenURI,
		"auth_provider_x509_cert_url": cfg.Firebase.AuthProviderCertURL,
		"client_x509_cert_url":        cfg.Firebase.ClientCertURL,
		"universe_domain":             cfg.Firebase.UniverseDomain,
	}

	// JSON으로 직렬화
	serviceAccountJSON, err := json.Marshal(serviceAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal service account: %w", err)
	}

	// Firebase 앱 초기화
	opt := option.WithCredentialsJSON(serviceAccountJSON)
	app, err := firebase.NewApp(context.Background(), &firebase.Config{
		ProjectID: cfg.Firebase.ProjectID,
	}, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase app: %w", err)
	}

	// Messaging 클라이언트 생성
	messagingClient, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create messaging client: %w", err)
	}

	log.Printf("Firebase client initialized successfully for project: %s", cfg.Firebase.ProjectID)

	return &Client{
		app:       app,
		messaging: messagingClient,
		projectID: cfg.Firebase.ProjectID,
	}, nil
}

// GetMessagingClient 메시징 클라이언트 반환
func (c *Client) GetMessagingClient() *messaging.Client {
	return c.messaging
}

// GetProjectID 프로젝트 ID 반환
func (c *Client) GetProjectID() string {
	return c.projectID
}

// Close 클라이언트 리소스 정리
func (c *Client) Close() error {
	// Firebase App은 자동으로 리소스를 관리하므로 별도 정리 불필요
	// 필요한 경우 여기에 추가 정리 로직을 구현할 수 있음
	return nil
}
