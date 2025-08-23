package bookmark

import (
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"

	"github.com/gofiber/fiber/v2"
)

// SentenceBookmarkResponse 문장 북마크 응답
type SentenceBookmarkResponse struct {
	UUID      string `json:"uuid"`
	Sentence  string `json:"sentence"`
	CreatedAt string `json:"created_at"`
}

// FindByUserUUIDSentenceBookmark 사용자의 모든 문장 북마크 조회 (인증 필요)
// @Summary 문장 북마크 전체 조회
// @Description 현재 인증된 사용자의 모든 문장 북마크를 조회합니다. 최신순으로 정렬되어 반환됩니다.
// @Tags Bookmark
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} SentenceBookmarkResponse "문장 북마크 목록"
// @Failure 401 {object} map[string]interface{} "인증 실패"
// @Router /bookmark/sentence [get]
func FindByUserUUIDSentenceBookmark(c *fiber.Ctx) error {
	// context에서 사용자 UUID 가져오기
	userUUID := middleware.GetUserUUID(c)

	// context에서 database 가져오기
	db := middleware.GetDB(c)

	// 사용자의 모든 문장 북마크 조회 (최신순)
	var bookmarks []models.SentenceBookmark
	if err := db.Where("user_uuid = ?", userUUID).
		Order("created_at DESC").
		Find(&bookmarks).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch bookmarks",
		})
	}

	// 응답 데이터 변환
	var response []SentenceBookmarkResponse
	for _, bookmark := range bookmarks {
		response = append(response, SentenceBookmarkResponse{
			UUID:      bookmark.UUID.String(),
			Sentence:  bookmark.Sentence,
			CreatedAt: bookmark.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return c.JSON(response)
}
