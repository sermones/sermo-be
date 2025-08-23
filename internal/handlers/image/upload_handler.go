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
	// UserID는 JWT 토큰에서 자동으로 추출됩니다
}

// UploadImageResponse 이미지 업로드 응답
type UploadImageResponse struct {
	Message string       `json:"message"`
	Image   models.Image `json:"image"`
}

// UploadImage 이미지 업로드 핸들러
// @Summary 이미지 업로드
// @Description 사용자 이미지를 업로드하고 DB에 정보를 저장합니다
// @Tags Image
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "이미지 파일"
// @Success 200 {object} UploadImageResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /image/upload [post]
func UploadImage(c *fiber.Ctx) error {
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

	// R2 클라이언트 가져오기
	r2Client := middleware.GetR2Client(c)
	if r2Client == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "R2 클라이언트 초기화 실패",
		})
	}

	// 파일 키 생성 (사용자UUID/타임스탬프_파일명)
	timestamp := time.Now().Unix()
	fileExt := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%d_%s", timestamp, strings.TrimSuffix(file.Filename, fileExt))
	key := fmt.Sprintf("images/%s/%s%s", userUUID, fileName, fileExt)

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
			"error":   "파일 업로드에 실패했습니다",
			"details": err.Error(),
		})
	}

	// DB에 이미지 정보 저장 (파일 키만 저장)
	image := models.NewImage(
		userUUID,
		file.Filename,
		file.Header.Get("Content-Type"),
		key, // 파일 키만 저장 (images/userid/filename)
		file.Size,
	)

	if err := database.DB.Create(&image).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "이미지 정보 저장에 실패했습니다",
			"details": err.Error(),
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
