package image

import (
	"fmt"
	"time"

	"sermo-be/internal/middleware"
	"sermo-be/internal/models"
	"sermo-be/pkg/database"

	"github.com/gofiber/fiber/v2"
)

// TestUploadImage 테스트용 이미지 업로드 핸들러 (R2 없이)
// @Summary 테스트용 이미지 업로드
// @Description R2 없이 테스트하는 이미지 업로드
// @Tags image
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "이미지 파일"
// @Success 200 {object} UploadImageResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /image/test-upload [post]
func TestUploadImage(c *fiber.Ctx) error {
	// JWT 토큰에서 사용자 UUID 추출
	userUUID := middleware.GetUserUUID(c)
	if userUUID == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "인증이 필요합니다",
		})
	}

	// 파일 파싱
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "파일이 필요합니다",
		})
	}

	// 파일 타입 검증
	if !isValidImageType(file.Header.Get("Content-Type")) {
		return c.Status(400).JSON(fiber.Map{
			"error": "지원하지 않는 이미지 형식입니다",
		})
	}

	// 테스트용 URL 생성 (실제 업로드 없음)
	timestamp := time.Now().Unix()
	testURL := fmt.Sprintf("https://test.example.com/images/%s/%d_%s", userUUID, timestamp, file.Filename)

	// DB에 이미지 정보 저장 (테스트용)
	image := models.NewImage(
		userUUID,
		file.Filename,
		file.Header.Get("Content-Type"),
		testURL,
		file.Size,
	)

	if err := database.DB.Create(&image).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "이미지 정보 저장에 실패했습니다",
			"details": err.Error(),
		})
	}

	return c.JSON(UploadImageResponse{
		Message: "테스트 이미지 업로드가 완료되었습니다 (R2 없음)",
		Image:   *image,
	})
}
