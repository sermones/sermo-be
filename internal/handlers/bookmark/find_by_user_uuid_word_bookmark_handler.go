package bookmark

import (
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"

	"github.com/gofiber/fiber/v2"
)

// WordBookmarkResponse 단어 북마크 응답
type WordBookmarkResponse struct {
	UUID      string `json:"uuid"`
	Word      string `json:"word"`
	CreatedAt string `json:"created_at"`
}

// FindByUserUUIDWordBookmark 사용자의 모든 단어 북마크 조회 (인증 필요)
// @Summary 단어 북마크 전체 조회
// @Description 현재 인증된 사용자의 모든 단어 북마크를 조회합니다. 최신순으로 정렬되어 반환됩니다.
// @Tags Bookmark
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} WordBookmarkResponse "단어 북마크 목록"
// @Failure 401 {object} map[string]interface{} "인증 실패"
// @Router /bookmark/word [get]
func FindByUserUUIDWordBookmark(c *fiber.Ctx) error {
	// context에서 사용자 UUID 가져오기
	userUUID := middleware.GetUserUUID(c)

	// context에서 database 가져오기
	db := middleware.GetDB(c)

	// 사용자의 모든 단어 북마크 조회 (최신순)
	var bookmarks []models.WordBookmark
	if err := db.Where("user_uuid = ?", userUUID).
		Order("created_at DESC").
		Find(&bookmarks).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch word bookmarks",
		})
	}

	// 응답 데이터 변환
	var response []WordBookmarkResponse
	for _, bookmark := range bookmarks {
		response = append(response, WordBookmarkResponse{
			UUID:      bookmark.UUID.String(),
			Word:      bookmark.Word,
			CreatedAt: bookmark.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return c.JSON(response)
}
