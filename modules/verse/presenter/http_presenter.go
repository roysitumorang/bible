package presenter

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/roysitumorang/bible/helper"
	verseModel "github.com/roysitumorang/bible/modules/verse/model"
	"github.com/roysitumorang/bible/modules/verse/sanitizer"
	verseUseCase "github.com/roysitumorang/bible/modules/verse/usecase"
	"go.uber.org/zap"
)

type (
	verseHTTPHandler struct {
		verseUseCase verseUseCase.VerseUseCase
	}
)

func New(
	verseUseCase verseUseCase.VerseUseCase,
) *verseHTTPHandler {
	return &verseHTTPHandler{
		verseUseCase: verseUseCase,
	}
}

func (q *verseHTTPHandler) Mount(r fiber.Router) {
	r.Get("", q.FindVerses)
}

func (q *verseHTTPHandler) FindVerses(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "VersePresenter-FindVerses"
	filter, err := sanitizer.FindVerses(ctx, c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindVerses")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	verses, err := q.verseUseCase.FindVerses(ctx, filter)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindVerses")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	mapBookChapterVerses := map[string]map[int][]verseModel.Verse{}
	for _, verse := range verses {
		if _, ok := mapBookChapterVerses[verse.BookName]; !ok {
			mapBookChapterVerses[verse.BookName] = map[int][]verseModel.Verse{}
		}
		mapBookChapterVerses[verse.BookName][verse.Chapter] = append(mapBookChapterVerses[verse.BookName][verse.Chapter], verse)
	}
	response := make([]verseModel.Passage, len(filter.Books))
	for i, book := range filter.Books {
		chapterEnd := book.ChapterStart
		if book.ChapterEnd > 0 {
			chapterEnd = book.ChapterEnd
		}
		var k int
		chapters := make([]verseModel.Chapter, chapterEnd-book.ChapterStart+1)
		for j := book.ChapterStart; j <= chapterEnd; j++ {
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
		response[i] = verseModel.Passage{
			BookName:     book.Name,
			ChapterStart: book.ChapterStart,
			ChapterEnd:   book.ChapterEnd,
			Chapters:     chapters,
		}
	}
	return helper.NewResponse(fiber.StatusOK, "", response).WriteResponse(c)
}
