package image

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"sermo-be/internal/middleware"
	"sermo-be/internal/models"
	"sermo-be/pkg/database"

	"github.com/gofiber/fiber/v2"
)

// UploadImageRequest 이미지 업로드 요청
type UploadImageRequest struct {
	UserID string `form:"user_id"`
}

// UploadImageResponse 이미지 업로드 응답
type UploadImageResponse struct {
	Message string       `json:"message"`
	Image   models.Image `json:"image"`
}

// UploadImage 이미지 업로드 핸들러
// @Summary 이미지 업로드
// @Description 사용자 이미지를 업로드하고 DB에 정보를 저장합니다
// @Tags image
// @Accept multipart/form-data
// @Produce json
// @Param user_id formData string true "사용자 ID"
// @Param file formData file true "이미지 파일"
// @Success 200 {object} UploadImageResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /image/upload [post]
func UploadImage(c *fiber.Ctx) error {
	// 사용자 ID 파싱
	userID := c.FormValue("user_id")
	if userID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "사용자 ID가 필요합니다",
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

	// R2 클라이언트 가져오기
	r2Client := middleware.GetR2Client(c)

	// 파일 키 생성 (사용자ID/타임스탬프_파일명)
	timestamp := time.Now().Unix()
	fileExt := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%d_%s", timestamp, strings.TrimSuffix(file.Filename, fileExt))
	key := fmt.Sprintf("images/%s/%s%s", userID, fileName, fileExt)

	// 파일을 R2에 업로드
	fileReader, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "파일을 열 수 없습니다",
		})
	}
	defer fileReader.Close()

	if err := r2Client.UploadFile(c.Context(), key, fileReader); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "파일 업로드에 실패했습니다",
		})
	}

	// R2 URL 생성 (public으로 설정)
	url := fmt.Sprintf("%s/%s", strings.TrimSuffix(r2Client.GetEndpoint(), "/"), key)

	// DB에 이미지 정보 저장
	image := models.NewImage(
		userID,
		file.Filename,
		file.Header.Get("Content-Type"),
		url,
		file.Size,
	)

	if err := database.DB.Create(&image).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "이미지 정보 저장에 실패했습니다",
		})
	}

	return c.JSON(UploadImageResponse{
		Message: "이미지 업로드가 완료되었습니다",
		Image:   *image,
	})
}

// isValidImageType 이미지 타입 검증
func isValidImageType(mimeType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, validType := range validTypes {
		if mimeType == validType {
			return true
		}
	}
	return false
}
