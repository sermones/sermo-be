// @title Sermo Backend API
// @version 1.0
// @description Sermo Backend API 서버입니다.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:3000
// @BasePath /
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT 토큰을 사용한 Bearer 인증

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "sermo-be/docs"
	"sermo-be/internal/config"
	"sermo-be/internal/middleware"
	"sermo-be/internal/routes"
	"sermo-be/pkg/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// 설정 로드
	cfg := config.Load()

	// 데이터베이스 연결
	dbConfig := &database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}

	if err := database.Connect(dbConfig); err != nil {
		log.Fatalf("데이터베이스 연결 실패: %v", err)
	}

	// 데이터베이스 마이그레이션 실행
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("데이터베이스 마이그레이션 실패: %v", err)
	}

	// Fiber 앱 생성
	app := fiber.New(fiber.Config{
		AppName: "Sermo Backend",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// 미들웨어 설정
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "*",
	}))

	// DI 미들웨어 설정
	app.Use(middleware.ConfigMiddleware(cfg))
	app.Use(middleware.DatabaseMiddleware(database.DB))
	app.Use(middleware.R2Middleware(cfg))
	app.Use(middleware.OpenAIMiddleware(cfg))

	// 라우터 설정
	routes.SetupRoutes(app)

	// 서버 시작
	serverAddr := cfg.Server.Host + ":" + cfg.Server.Port
	go func() {
		if err := app.Listen(serverAddr); err != nil {
			log.Fatalf("서버 시작 실패: %v", err)

		}
	}()

	log.Printf("🚀 Sermo Backend 서버가 %s에서 시작되었습니다", serverAddr)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 서버를 종료합니다...")

	// SSE 세션 정리
	log.Println("🔄 SSE 세션 정리 중...")
	sseManager := middleware.GetSSEManager()
	sseManager.Shutdown()

	if err := app.Shutdown(); err != nil {
		log.Fatalf("서버 종료 실패: %v", err)
	}
	log.Println("✅ 서버가 안전하게 종료되었습니다")
}
