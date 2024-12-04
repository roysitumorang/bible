package sanitizer

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	verseModel "github.com/roysitumorang/bible/modules/verse/model"
)

func FindVerses(ctx context.Context, c *fiber.Ctx) (response *verseModel.Filter, err error) {
	var filterOptions []verseModel.FilterOption
	if versionCode := c.Query("version"); versionCode != "" {
		filterOptions = append(filterOptions, verseModel.WithVersionCode(versionCode))
	}
	keywords := strings.Split(c.Query("q"), ";")
	for _, keyword := range keywords {
		parts := strings.Split(keyword, " ")
		n := len(parts)
		if n != 2 && parts[0] == "" {
			continue
		}
		bookName, chapters := parts[0], parts[1]
		var chapterStart, chapterEnd int
		if strings.Contains(chapters, "-") {
			subParts := strings.Split(chapters, "-")
			if len(subParts) != 2 || subParts[0] == "" {
				continue
			}
			if chapterStart, err = strconv.Atoi(subParts[0]); err != nil {
				return
			}
			if chapterEnd, err = strconv.Atoi(subParts[1]); err != nil {
				return
			}
			chapterEnd = min(chapterEnd, chapterStart+20)
		} else if chapterStart, err = strconv.Atoi(chapters); err != nil {
			return
		}
		filterOptions = append(filterOptions, verseModel.WithBook(bookName, chapterStart, chapterEnd))
	}
	response = verseModel.NewFilter(filterOptions...)
	return
}
