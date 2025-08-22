package image

import (
	"fmt"
	"io"
	"time"

	"sermo-be/internal/middleware"
	"sermo-be/internal/models"
	"sermo-be/pkg/database"

	"github.com/gofiber/fiber/v2"
)

// GetImageRequest 이미지 조회 요청
type GetImageRequest struct {
	ImageID string `json:"image_id" validate:"required"`
}

// GetImageResponse 이미지 조회 응답
type GetImageResponse struct {
	Message string       `json:"message"`
	Image   models.Image `json:"image"`
	URL     string       `json:"url,omitempty"` // 프리사인드 URL (선택사항)
}

// GetImage 이미지 조회 핸들러
// @Summary 이미지 조회
// @Description 이미지 정보를 조회하고 프리사인드 URL을 생성합니다
// @Tags image
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param image_id path string true "이미지 ID"
// @Success 200 {object} GetImageResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /image/{image_id} [get]
func GetImage(c *fiber.Ctx) error {
	// JWT 토큰에서 사용자 UUID 추출
	userUUID := middleware.GetUserUUID(c)
	if userUUID == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "인증이 필요합니다",
		})
	}

	// 이미지 ID 파싱
	imageID := c.Params("image_id")
	if imageID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "이미지 ID가 필요합니다",
		})
	}

	// R2 클라이언트 가져오기
	r2Client := middleware.GetR2Client(c)
	if r2Client == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "R2 클라이언트 초기화 실패",
		})
	}

	// DB에서 이미지 정보 조회
	var image models.Image
	if err := database.DB.Where("id = ? AND user_id = ?", imageID, userUUID).First(&image).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "이미지를 찾을 수 없습니다",
		})
	}

	// 프리사인드 URL 생성 (24시간 유효)
	presignedURL, err := r2Client.GeneratePresignedURL(c.Context(), image.FileKey, 24*time.Hour)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "프리사인드 URL 생성에 실패했습니다",
			"details": err.Error(),
		})
	}

	return c.JSON(GetImageResponse{
		Message: "이미지 조회가 완료되었습니다",
		Image:   image,
		URL:     presignedURL,
	})
}

// DownloadImage 이미지 다운로드 핸들러
// @Summary 이미지 다운로드
// @Description 이미지를 직접 다운로드합니다
// @Tags image
// @Accept json
// @Produce octet-stream
// @Security BearerAuth
// @Param image_id path string true "이미지 ID"
// @Success 200 {file} file
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /image/{image_id}/download [get]
func DownloadImage(c *fiber.Ctx) error {
	// JWT 토큰에서 사용자 UUID 추출
	userUUID := middleware.GetUserUUID(c)
	if userUUID == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "인증이 필요합니다",
		})
	}

	// 이미지 ID 파싱
	imageID := c.Params("image_id")
	if imageID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "이미지 ID가 필요합니다",
		})
	}

	// R2 클라이언트 가져오기
	r2Client := middleware.GetR2Client(c)
	if r2Client == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "R2 클라이언트 초기화 실패",
		})
	}

	// DB에서 이미지 정보 조회
	var image models.Image
	if err := database.DB.Where("id = ? AND user_id = ?", imageID, userUUID).First(&image).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "이미지를 찾을 수 없습니다",
		})
	}

	// R2에서 파일 다운로드
	fileReader, err := r2Client.DownloadFile(c.Context(), image.FileKey)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "파일 다운로드에 실패했습니다",
			"details": err.Error(),
		})
	}
	defer fileReader.Close()

	// 파일 내용을 응답으로 전송
	fileContent, err := io.ReadAll(fileReader)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "파일 읽기에 실패했습니다",
			"details": err.Error(),
		})
	}

	// 응답 헤더 설정
	c.Set("Content-Type", image.MimeType)
	c.Set("Content-Disposition", "attachment; filename="+image.FileName)
	c.Set("Content-Length", fmt.Sprintf("%d", image.FileSize))

	return c.Send(fileContent)
}
