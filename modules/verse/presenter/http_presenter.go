package presenter

import (
	"github.com/gofiber/fiber/v2"
	"github.com/roysitumorang/bible/helper"
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

// @Summary		Finding verses
// @Description	finding verses
// @Tags			Verse
// @ID				FindVerses
// @Accept			json
// @Produce		json
// @Param			version	query		string	true	"Version code"
// @Param			q		query		string	true	"BookName chapterNo:verseNoStart-verseNoEnd"
// @Success		200		{object}	model.ResponsePassages
// @Router			/verses [get]
func (q *verseHTTPHandler) FindVerses(c *fiber.Ctx) error {
	ctx := c.UserContext()
	ctxt := "VersePresenter-FindVerses"
	versesFilter, err := sanitizer.FindVerses(ctx, c)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindVerses")
		return verseModel.NewResponsePassages(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	verses, err := q.verseUseCase.SearchVerses(ctx, versesFilter)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSearchVerses")
		return verseModel.NewResponsePassages(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	mapBookChapterVerses := map[string]map[int][]verseModel.Verse{}
	for _, verse := range verses {
		mapChapterVerses, ok := mapBookChapterVerses[verse.BookName]
		if !ok {
			mapChapterVerses = map[int][]verseModel.Verse{}
		}
		mapChapterVerses[verse.Chapter] = append(mapChapterVerses[verse.Chapter], verse)
		mapBookChapterVerses[verse.BookName] = mapChapterVerses
	}
	passages := make([]verseModel.Passage, len(versesFilter.Books))
	for i, book := range versesFilter.Books {
		verses, ok := mapBookChapterVerses[book.Name][book.Chapter]
		if !ok {
			continue
		}
		passage := verseModel.Passage{
			BookName:     book.Name,
			Chapter:      book.Chapter,
			VerseNoStart: book.VerseNoStart,
			VerseNoEnd:   book.VerseNoEnd,
			Verses:       []verseModel.Verse{},
		}
		for _, verse := range verses {
			if verse.Number >= book.VerseNoStart && verse.Number <= book.VerseNoEnd {
				passage.Verses = append(passage.Verses, verse.Doc())
			}
		}
		passages[i] = passage
	}
	return verseModel.NewResponsePassages(fiber.StatusOK, "", passages).WriteResponse(c)
}
