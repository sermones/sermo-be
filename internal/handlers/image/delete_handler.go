package image

import (
	"fmt"
	"strings"

	"sermo-be/internal/middleware"
	"sermo-be/internal/models"
	"sermo-be/pkg/database"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// DeleteImageRequest 이미지 삭제 요청
type DeleteImageRequest struct {
	UserID  string `json:"user_id"`
	ImageID string `json:"image_id"`
}

// DeleteImageResponse 이미지 삭제 응답
type DeleteImageResponse struct {
	Message string `json:"message"`
}

// DeleteImage 이미지 삭제 핸들러
// @Summary 이미지 삭제
// @Description 사용자 이미지를 삭제합니다
// @Tags image
// @Accept json
// @Produce json
// @Param request body DeleteImageRequest true "삭제 요청"
// @Success 200 {object} DeleteImageResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /image/delete [delete]
func DeleteImage(c *fiber.Ctx) error {
	// 요청 파싱
	var req DeleteImageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "잘못된 요청 형식입니다",
		})
	}

	// 필수 필드 검증
	if req.UserID == "" || req.ImageID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "사용자 ID와 이미지 ID가 필요합니다",
		})
	}

	// 이미지 ID 파싱
	imageUUID, err := uuid.Parse(req.ImageID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "잘못된 이미지 ID 형식입니다",
		})
	}

	// DB에서 이미지 정보 조회
	var image models.Image
	if err := database.DB.Where("id = ? AND user_id = ?", imageUUID, req.UserID).First(&image).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "이미지를 찾을 수 없습니다",
		})
	}

	// R2 클라이언트 가져오기
	r2Client := middleware.GetR2Client(c)

	// R2에서 파일 삭제 (키는 URL에서 추출)
	// URL: https://035ea6d0a6159a3e4c6f41ded546d78d.r2.cloudflarestorage.com/images/userid/filename.jpg
	// 키: images/userid/filename.jpg
	urlParts := strings.Split(image.URL, "/")
	if len(urlParts) >= 4 {
		key := strings.Join(urlParts[3:], "/") // images/userid/filename.jpg 부분 추출

		if err := r2Client.DeleteFile(c.Context(), key); err != nil {
			// R2 삭제 실패해도 DB는 삭제 (일관성 유지)
			fmt.Printf("R2 파일 삭제 실패: %v\n", err)
		}
	}

	// DB에서 이미지 정보 삭제
	if err := database.DB.Delete(&image).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "이미지 정보 삭제에 실패했습니다",
		})
	}

	return c.JSON(DeleteImageResponse{
		Message: "이미지가 성공적으로 삭제되었습니다",
	})
}
