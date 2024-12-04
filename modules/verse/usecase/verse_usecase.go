package usecase

import (
	"context"

	"github.com/roysitumorang/bible/helper"
	verseModel "github.com/roysitumorang/bible/modules/verse/model"
	verseQuery "github.com/roysitumorang/bible/modules/verse/query"
	"go.uber.org/zap"
)

type (
	verseUseCase struct {
		verseQuery verseQuery.VerseQuery
	}
)

func New(
	verseQuery verseQuery.VerseQuery,
) VerseUseCase {
	return &verseUseCase{
		verseQuery: verseQuery,
	}
}

func (q *verseUseCase) FindVerses(ctx context.Context, filter *verseModel.Filter) (response []verseModel.Verse, err error) {
	ctxt := "VerseUseCase-FindVerses"
	if response, err = q.verseQuery.FindVerses(ctx, filter); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindVerses")
	}
	return
}
