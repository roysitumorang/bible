package presenter

import (
	"github.com/gofiber/fiber/v2"
	"github.com/roysitumorang/bible/helper"
	bookModel "github.com/roysitumorang/bible/modules/book/model"
	bookUseCase "github.com/roysitumorang/bible/modules/book/usecase"
	verseModel "github.com/roysitumorang/bible/modules/verse/model"
	"github.com/roysitumorang/bible/modules/verse/sanitizer"
	verseUseCase "github.com/roysitumorang/bible/modules/verse/usecase"
	"go.uber.org/zap"
)

type (
	verseHTTPHandler struct {
		bookUseCase  bookUseCase.BookUseCase
		verseUseCase verseUseCase.VerseUseCase
	}
)

func New(
	bookUseCase bookUseCase.BookUseCase,
	verseUseCase verseUseCase.VerseUseCase,
) *verseHTTPHandler {
	return &verseHTTPHandler{
		bookUseCase:  bookUseCase,
		verseUseCase: verseUseCase,
	}
}

func (q *verseHTTPHandler) Mount(r fiber.Router) {
	r.Get("", q.FindVerses)
}

// swagger:operation GET /verses Verse FindVerses
// Get verses
// ---
// produces:
//   - "application/json"
// parameters:
//   - name: version
//     in: query
//     required: true
//     type: string
//   - name: q
//     in: query
//     required: true
//     type: string
// responses:
//   200:
//     description: "successful operation"
//     schema:
//       $ref: "#/definitions/ResponsePassages"
func (q *verseHTTPHandler) FindVerses(c *fiber.Ctx) error {
	ctx := c.UserContext()
	ctxt := "VersePresenter-FindVerses"
	versesFilter, err := sanitizer.FindVerses(ctx, c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindVerses")
		return helper.NewResponse(c, fiber.StatusBadRequest, err.Error()).WriteResponse(c, nil)
	}
	mapBookChaptersCount := map[string]int{}
	if n := len(versesFilter.Books); n > 0 {
		bookNames := make([]string, n)
		for i, book := range versesFilter.Books {
			bookNames[i] = book.Name
		}
		books, err := q.bookUseCase.FindBooks(ctx, bookModel.NewFilter(bookModel.WithNames(bookNames...)))
		if err != nil {
			helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindBooks")
			return helper.NewResponse(c, fiber.StatusBadRequest, err.Error()).WriteResponse(c, nil)
		}
		for _, book := range books {
			mapBookChaptersCount[book.Name] = book.ChaptersCount
		}
	}
	verses, err := q.verseUseCase.SearchVerses(ctx, versesFilter)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSearchVerses")
		return helper.NewResponse(c, fiber.StatusBadRequest, err.Error()).WriteResponse(c, nil)
	}
	mapBookChapterVerses := map[string]map[int][]verseModel.Verse{}
	for _, verse := range verses {
		if _, ok := mapBookChapterVerses[verse.BookName]; !ok {
			mapBookChapterVerses[verse.BookName] = map[int][]verseModel.Verse{}
		}
		mapBookChapterVerses[verse.BookName][verse.Chapter] = append(mapBookChapterVerses[verse.BookName][verse.Chapter], verse.Doc())
	}
	passages := make([]verseModel.Passage, len(versesFilter.Books))
	for i, book := range versesFilter.Books {
		chaptersCount, ok := mapBookChaptersCount[book.Name]
		if !ok {
			continue
		}
		verseNoEnd := book.VerseNoStart
		if book.VerseNoEnd > 0 {
			verseNoEnd = min(book.VerseNoEnd, chaptersCount)
		}
		var k int
		chapters := make([]verseModel.Chapter, verseNoEnd-book.VerseNoStart+1)
		for j := book.VerseNoStart; j <= verseNoEnd; j++ {
			verses, ok := mapBookChapterVerses[book.Name][j]
			if !ok {
				verses = []verseModel.Verse{}
			}
			chapters[k] = verseModel.Chapter{
				Number: j,
				Verses: verses,
			}
			k++
		}
		passages[i] = verseModel.Passage{
			BookName:     book.Name,
			VerseNoStart: book.VerseNoStart,
			VerseNoEnd:   book.VerseNoEnd,
			Chapters:     chapters,
		}
	}
	r := helper.NewResponse(c, fiber.StatusOK, "")
	response := verseModel.ReponsePassages{
		Response: r,
		Data:     passages,
	}
	return r.WriteResponse(c, response)
}
