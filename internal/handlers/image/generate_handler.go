package image

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"sermo-be/pkg/prompt"
)

// GenerateImageRequest 이미지 생성 요청 DTO
type GenerateImageRequest struct {
	Prompt string `json:"prompt"` // 이미지 생성 프롬프트 (필수)
	Style  string `json:"style"`  // 이미지 스타일 (선택사항, 비워두면 config의 애니메이션 프로필 사진 스타일 사용)
}

// GenerateImageResponse 이미지 생성 응답 DTO
type GenerateImageResponse struct {
	Message  string   `json:"message"`
	ImageIDs []string `json:"image_ids"` // 생성된 이미지 ID 리스트
}

// GenerateImage Gemini API를 사용하여 이미지 생성 (인증 필요)
// @Summary AI 이미지 생성
// @Description Gemini API를 사용하여 텍스트 프롬프트로 이미지를 생성합니다. 회원가입한 사용자만 사용 가능합니다. 이미지 사이즈는 config에서 자동으로 1024x1024로 설정됩니다.
// @Tags Image
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body GenerateImageRequest true "이미지 생성 정보 (prompt는 필수, style은 선택사항)"
// @Success 200 {object} GenerateImageResponse
// @Success 200 {object} GenerateImageResponse{image_ids=[]string}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /image/generate [post]
func GenerateImage(c *fiber.Ctx) error {
	// 요청 파싱
	var req GenerateImageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// 입력 검증
	if req.Prompt == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Prompt is required",
		})
	}

	// context에서 Gemini 클라이언트 가져오기
	geminiClient := middleware.GetGeminiClient(c)
	if geminiClient == nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gemini client not available",
		})
	}

	// 향상된 프롬프트 생성 (Google 공식 문서 권장사항 적용)
	enhancedPrompt := buildEnhancedPrompt(req.Prompt, req.Style, c)

	fmt.Printf("Enhanced prompt: %s\n", enhancedPrompt)

	// 이미지 생성
	response, err := geminiClient.GenerateImage(c.Context(), enhancedPrompt)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate image: " + err.Error(),
		})
	}

	// 디버깅: Gemini 응답 확인
	fmt.Printf("Gemini response: %d images received\n", len(response.Images))
	for i, img := range response.Images {
		fmt.Printf("Image %d: Data length=%d, First 100 chars=%s...\n",
			i, len(img.Data), img.Data[:min(100, len(img.Data))])
	}

	// 이미지가 생성되지 않은 경우 처리
	if len(response.Images) == 0 {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "No images were generated. Please try with a different prompt or check if image generation is available in your region.",
		})
	}

	// context에서 database 가져오기
	db := middleware.GetDB(c)

	// context에서 사용자 UUID 가져오기
	userUUID := middleware.GetUserUUID(c)

	// context에서 R2 클라이언트 가져오기
	r2Client := middleware.GetR2Client(c)
	if r2Client == nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "R2 client not available",
		})
	}

	// 생성된 이미지들을 R2에 업로드하고 DB에 메타데이터 저장
	var imageIDs []string
	for _, img := range response.Images {
		// 랜덤 파일명 생성
		fileName := generateRandomFileName()

		// R2 파일 키 생성 (images/user_uuid/filename 형식)
		fileKey := fmt.Sprintf("images/%s/%s", userUUID, fileName)

		// base64 데이터를 바이트로 변환
		imageBytes, err := base64.StdEncoding.DecodeString(img.Data)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to decode base64 image data",
			})
		}

		// 디버깅: 이미지 크기와 데이터 확인
		fmt.Printf("Image %d: Size=%d bytes, FileKey=%s\n", len(imageIDs), len(imageBytes), fileKey)

		// R2에 이미지 업로드
		reader := bytes.NewReader(imageBytes)
		if err := r2Client.UploadFile(c.Context(), fileKey, reader); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to upload image to R2: " + err.Error(),
			})
		}

		fmt.Printf("Successfully uploaded image %d to R2\n", len(imageIDs))

		// 새로운 이미지 모델 생성 (R2 파일 키 저장)
		image := models.NewImage(
			userUUID,               // 사용자 UUID
			fileName,               // 프롬프트 기반 파일명
			"image/png",            // MIME 타입
			fileKey,                // R2 파일 키
			int64(len(imageBytes)), // 파일 크기
		)

		// 데이터베이스에 저장
		if err := db.Create(image).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save image metadata to database",
			})
		}

		imageIDs = append(imageIDs, image.ID.String())
	}

	result := GenerateImageResponse{
		Message:  "Images generated and saved successfully",
		ImageIDs: imageIDs,
	}

	return c.JSON(result)
}

// buildEnhancedPrompt 향상된 프롬프트 생성 (Google 공식 문서 권장사항 적용)
func buildEnhancedPrompt(basePrompt, style string, c *fiber.Ctx) string {
	// config에서 기본 이미지 사이즈와 스타일 가져오기
	config := middleware.GetConfig(c)

	// 기본 프롬프트에 이미지 생성 요청을 명시적으로 추가
	enhancedPrompt := fmt.Sprintf("Please generate an image of \n Detail(you must follow this instruction): %s", basePrompt)

	// 스타일 정보 추가 (사용자 입력이 있으면 사용, 없으면 config 기본값 사용)
	if style != "" {
		enhancedPrompt += fmt.Sprintf(", in %s style", style)
	} else if config != nil && config.Gemini.ImageStyle != "" {
		enhancedPrompt += fmt.Sprintf(", in %s style", config.Gemini.ImageStyle)
	}

	// config에서 가져온 이미지 사이즈 추가
	if config != nil && config.Gemini.ImageSize != "" {
		enhancedPrompt += fmt.Sprintf(", %s size", config.Gemini.ImageSize)
	}

	enhancedPrompt += fmt.Sprintf(", additional_instruction: %s", prompt.GetProfileImageGeneratePrompt())

	// 이미지 생성을 명시적으로 요청 (Google 공식 문서 권장사항)
	enhancedPrompt += ". Please provide an image for this prompt."

	return enhancedPrompt
}

// generateRandomFileName 랜덤한 파일명 생성
func generateRandomFileName() string {
	// UUID 기반 랜덤 파일명 생성
	randomID := uuid.New().String()[:8] // UUID의 앞 8자만 사용
	return fmt.Sprintf("gemini_%s.png", randomID)
}

// min 두 정수 중 작은 값 반환
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
