package presenter

import (
	"context"

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

func (q *verseHTTPHandler) FindVerses(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "VersePresenter-FindVerses"
	versesFilter, err := sanitizer.FindVerses(ctx, c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindVerses")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
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
			return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
		}
		for _, book := range books {
			mapBookChaptersCount[book.Name] = book.ChaptersCount
		}
	}
	verses, err := q.verseUseCase.FindVerses(ctx, versesFilter)
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
	response := make([]verseModel.Passage, len(versesFilter.Books))
	for i, book := range versesFilter.Books {
		chaptersCount, ok := mapBookChaptersCount[book.Name]
		if !ok {
			continue
		}
		chapterEnd := book.ChapterStart
		if book.ChapterEnd > 0 {
			chapterEnd = min(book.ChapterEnd, chaptersCount)
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
