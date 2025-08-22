package auth

import (
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"
	"sermo-be/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// DTO 정의
type LoginRequest struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
}

// Login 로그인 핸들러
// @Summary 로그인
// @Description 사용자 인증 후 JWT 토큰을 반환합니다.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "로그인 정보"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/login [post]
func Login(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// 유효성 검사
	if req.ID == "" || req.Password == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "ID and password are required",
		})
	}

	// context에서 database 가져오기
	db := middleware.GetDB(c)

	// 데이터베이스에서 사용자 조회
	var user models.User
	if err := db.Where("id = ?", req.ID).First(&user).Error; err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// 비밀번호 검증 (TODO: 실제로는 해시화된 비밀번호 비교)
	if user.Password != req.Password {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// JWT 토큰 생성
	token, err := jwt.GenerateToken(user.UUID.String())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	response := LoginResponse{
		Token: token,
		User: struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		}{
			ID:       user.UUID.String(),
			Username: user.ID,
		},
	}

	return c.JSON(response)
}
