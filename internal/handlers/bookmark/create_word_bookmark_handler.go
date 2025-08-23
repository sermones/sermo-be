package bookmark

import (
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"
	"sermo-be/pkg/openai"
	"sermo-be/pkg/prompt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// CreateWordBookmarkRequest 단어 북마크 생성 요청
type CreateWordBookmarkRequest struct {
	Word string `json:"word" validate:"required,min=1,max=100"`
}

// CreateWordBookmarkResponse 단어 북마크 생성 응답
type CreateWordBookmarkResponse struct {
	Message string `json:"message"`
	UUID    string `json:"uuid"`
}

// CreateWordBookmark 단어 북마크 생성 (인증 필요)
// @Summary 단어 북마크 생성
// @Description 새로운 단어 북마크를 생성합니다. OpenAI를 사용하여 자동으로 한글 뜻을 추출합니다. 단어는 1-100자까지 입력 가능합니다.
// @Tags Bookmark
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateWordBookmarkRequest true "북마크 생성 요청 (word 필수)"
// @Success 201 {object} CreateWordBookmarkResponse "북마크 생성 성공"
// @Failure 400 {object} map[string]interface{} "잘못된 요청 (단어 길이 제한 등)"
// @Failure 401 {object} map[string]interface{} "인증 실패"
// @Failure 500 {object} map[string]interface{} "OpenAI API 오류 또는 북마크 생성 실패"
// @Router /bookmark/word [post]
func CreateWordBookmark(c *fiber.Ctx) error {
	// context에서 사용자 UUID 가져오기
	userUUIDStr := middleware.GetUserUUID(c)

	// string을 uuid.UUID로 변환
	userUUID, err := uuid.Parse(userUUIDStr)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid user UUID",
		})
	}

	// context에서 database 가져오기
	db := middleware.GetDB(c)

	// 요청 파싱
	var req CreateWordBookmarkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// 단어 길이 검증
	if len(req.Word) == 0 || len(req.Word) > 100 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Word must be between 1 and 100 characters",
		})
	}

	// OpenAI를 사용하여 한글 뜻 추출
	openaiClient := middleware.GetOpenAIClient(c)
	if openaiClient == nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "OpenAI service unavailable",
		})
	}

	meaningPrompt := prompt.GetWordBookmarkMeaningPrompt()

	// OpenAI 채팅 완성 API 호출
	messages := []openai.ChatMessage{
		{
			Role:    "user",
			Content: meaningPrompt + "\n\n단어: " + req.Word,
		},
	}

	chatResp, err := openaiClient.ChatCompletion(c.Context(), messages)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate meaning using OpenAI",
		})
	}

	meaning := chatResp.Message.Content

	// 한글 뜻 길이 검증 (500자 제한)
	if len(meaning) > 500 {
		meaning = meaning[:500]
	}

	// 새로운 북마크 생성
	bookmark := models.NewWordBookmark(userUUID, req.Word, meaning)

	// 데이터베이스에 저장
	if err := db.Create(bookmark).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create bookmark",
		})
	}

	response := CreateWordBookmarkResponse{
		Message: "Word bookmark created successfully",
		UUID:    bookmark.UUID.String(),
	}

	return c.Status(http.StatusCreated).JSON(response)
}
