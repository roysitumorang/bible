package sanitizer

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	verseModel "github.com/roysitumorang/bible/modules/verse/model"
)

var (
	pattern = regexp.MustCompile(`^(.*)\s(\d+)(:(\d+)(-(\d+))?)?$`)
)

func FindVerses(ctx context.Context, c *fiber.Ctx) (response *verseModel.Filter, err error) {
	var filterOptions []verseModel.FilterOption
	if versionCode := c.Query("version"); versionCode != "" {
		filterOptions = append(filterOptions, verseModel.WithVersionCode(versionCode))
	}
	keywords := strings.Split(c.Query("q"), ";")
	if len(keywords) > 0 && keywords[0] != "" {
		for _, keyword := range keywords {
			parts := pattern.FindStringSubmatch(keyword)
			n := len(parts)
			if n != 7 {
				continue
			}
			bookName := parts[1]
			chapter, verseNoStart, verseNoEnd := 1, 1, 1
			if parts[2] != "" {
				chapter, _ = strconv.Atoi(parts[2])
				if parts[4] != "" {
					verseNoStart, _ = strconv.Atoi(parts[4])
					if parts[6] != "" {
						verseNoEnd, _ = strconv.Atoi(parts[6])
					}
				}
			}
			verseNoEnd = max(min(verseNoEnd, verseNoStart+20), verseNoStart)
			filterOptions = append(filterOptions, verseModel.WithBook(bookName, chapter, verseNoStart, verseNoEnd))
		}
	}
	response = verseModel.NewFilter(filterOptions...)
	return
}
