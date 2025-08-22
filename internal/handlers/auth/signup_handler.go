package auth

import (
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// DTO 정의
type SignUpRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SignUpResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}

// SignUp 회원가입 핸들러
func SignUp(c *fiber.Ctx) error {
	var req SignUpRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// 유효성 검사
	if len(req.Username) < 3 || len(req.Username) > 20 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Username must be between 3 and 20 characters",
		})
	}

	if len(req.Password) < 6 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at least 6 characters",
		})
	}

	// context에서 database 가져오기
	db := middleware.GetDB(c)

	// 데이터베이스에서 중복 사용자 체크
	var existingUser models.User
	if err := db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error": "Username already exists",
		})
	}

	// 새 사용자 생성
	user := &models.User{
		ID:       uuid.New(),
		Username: req.Username,
		Password: req.Password, // TODO: 실제로는 해시화 필요
	}

	// 데이터베이스에 사용자 저장
	if err := db.Create(user).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	response := SignUpResponse{
		Message: "User created successfully",
		UserID:  user.ID.String(),
	}

	return c.Status(http.StatusCreated).JSON(response)
}
