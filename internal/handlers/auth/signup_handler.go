package auth

import (
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"

	"github.com/gofiber/fiber/v2"
)

// DTO 정의
type SignUpRequest struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

type SignUpResponse struct {
	Message string `json:"message"`
	UUID    string `json:"uuid"`
}

// SignUp 회원가입 핸들러
// @Summary 회원가입
// @Description 새로운 사용자 계정을 생성합니다.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body SignUpRequest true "회원가입 정보"
// @Success 201 {object} SignUpResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /auth/signup [post]
func SignUp(c *fiber.Ctx) error {
	var req SignUpRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// 유효성 검사
	if len(req.ID) < 3 || len(req.ID) > 20 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "ID must be between 3 and 20 characters",
		})
	}

	if len(req.Nickname) < 1 || len(req.Nickname) > 100 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Nickname must be between 1 and 100 characters",
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
	if err := db.Where("id = ?", req.ID).First(&existingUser).Error; err == nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error": "ID already exists",
		})
	}

	// 새 사용자 생성
	user := models.NewUser(req.ID, req.Nickname, req.Password)

	// 데이터베이스에 사용자 저장
	if err := db.Create(user).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	response := SignUpResponse{
		Message: "User created successfully",
		UUID:    user.UUID.String(),
	}

	return c.Status(http.StatusCreated).JSON(response)
}
