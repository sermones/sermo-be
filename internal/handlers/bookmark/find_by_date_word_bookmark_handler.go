package bookmark

import (
	"net/http"
	"sermo-be/internal/middleware"
	"sermo-be/internal/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

// FindByDateWordBookmark 특정 날짜의 단어 북마크 조회 (인증 필요)
// @Summary 날짜별 단어 북마크 조회
// @Description 현재 인증된 사용자의 특정 날짜 단어 북마크를 조회합니다.
// @Tags Bookmark
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param date query string true "조회할 날짜 (YYYY-MM-DD 형식)"
// @Success 200 {array} WordBookmarkResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /bookmark/word/date [get]
func FindByDateWordBookmark(c *fiber.Ctx) error {
	// context에서 사용자 UUID 가져오기
	userUUID := middleware.GetUserUUID(c)

	// context에서 database 가져오기
	db := middleware.GetDB(c)

	// 날짜 파라미터 가져오기
	dateStr := c.Query("date")
	if dateStr == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Date parameter is required (YYYY-MM-DD format)",
		})
	}

	// 날짜 파싱 (YYYY-MM-DD 형식)
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date format. Use YYYY-MM-DD",
		})
	}

	// 다음 날짜 계산 (해당 날짜의 23:59:59까지 포함)
	nextDate := date.Add(24 * time.Hour)

	// 해당 날짜의 북마크 조회
	var bookmarks []models.WordBookmark
	if err := db.Where("user_uuid = ? AND created_at >= ? AND created_at < ?",
		userUUID, date, nextDate).
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
