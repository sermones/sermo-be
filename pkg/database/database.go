package database

import (
	"fmt"
	"log"
	"time"

	"sermo-be/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Config 데이터베이스 연결 설정
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Connect 데이터베이스 연결
func Connect(config *Config) error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.Host,
		config.User,
		config.Password,
		config.DBName,
		config.Port,
		config.SSLMode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 개발환경에서는 SQL 로그 출력
		NowFunc: func() time.Time {
			return time.Now().UTC() // UTC로 저장하고 GORM이 자동으로 변환
		},
	})

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// 연결 풀 설정
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %v", err)
	}

	// 연결 풀 설정
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("✅ Database connected successfully")
	return nil
}

// Close 데이터베이스 연결 종료
func Close() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// RunMigrations 데이터베이스 마이그레이션 실행
func RunMigrations() error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	// 마이그레이션할 모델들
	models := []interface{}{
		&models.User{},
		&models.Image{},
		&models.Chatbot{},
		&models.ChatMessage{},
		&models.SentenceBookmark{},
	}

	err := DB.AutoMigrate(models...)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	log.Println("✅ Database migrations completed successfully")
	return nil
}
