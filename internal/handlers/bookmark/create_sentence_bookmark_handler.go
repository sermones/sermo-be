package bookmark

import (
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// CreateSentenceBookmarkRequest 문장 북마크 생성 요청
type CreateSentenceBookmarkRequest struct {
	Sentence string `json:"sentence" validate:"required,min=1,max=1000"`
}

// CreateSentenceBookmarkResponse 문장 북마크 생성 응답
type CreateSentenceBookmarkResponse struct {
	Message string `json:"message"`
	UUID    string `json:"uuid"`
}

// CreateSentenceBookmark 문장 북마크 생성 (인증 필요)
// @Summary 문장 북마크 생성
// @Description 새로운 문장 북마크를 생성합니다. 문장은 1-1000자까지 입력 가능합니다.
// @Tags Bookmark
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateSentenceBookmarkRequest true "북마크 생성 요청 (sentence 필수)"
// @Success 201 {object} CreateSentenceBookmarkResponse "북마크 생성 성공"
// @Failure 400 {object} map[string]interface{} "잘못된 요청 (문장 길이 제한 등)"
// @Failure 401 {object} map[string]interface{} "인증 실패"
// @Router /bookmark/sentence [post]
func CreateSentenceBookmark(c *fiber.Ctx) error {
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
	var req CreateSentenceBookmarkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// 문장 길이 검증
	if len(req.Sentence) == 0 || len(req.Sentence) > 1000 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Sentence must be between 1 and 1000 characters",
		})
	}

	// 새로운 북마크 생성
	bookmark := models.NewSentenceBookmark(userUUID, req.Sentence)

	// 데이터베이스에 저장
	if err := db.Create(bookmark).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create bookmark",
		})
	}

	response := CreateSentenceBookmarkResponse{
		Message: "Bookmark created successfully",
		UUID:    bookmark.UUID.String(),
	}

	return c.Status(http.StatusCreated).JSON(response)
}
