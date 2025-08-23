package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// Connect Redis 연결
func Connect(host, port, password string, db int) error {
	addr := fmt.Sprintf("%s:%s", host, port)

	RedisClient = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// 연결 테스트
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	log.Println("✅ Redis connected successfully")
	return nil
}

// Close Redis 연결 종료
func Close() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

// SetKey 키-값 설정 (TTL 포함)
func SetKey(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return RedisClient.Set(ctx, key, value, ttl).Err()
}

// GetKey 키 값 조회
func GetKey(ctx context.Context, key string) (string, error) {
	return RedisClient.Get(ctx, key).Result()
}

// DeleteKey 키 삭제
func DeleteKey(ctx context.Context, key string) error {
	return RedisClient.Del(ctx, key).Err()
}

// SetExpire 키 TTL 설정
func SetExpire(ctx context.Context, key string, ttl time.Duration) error {
	return RedisClient.Expire(ctx, key, ttl).Err()
}
