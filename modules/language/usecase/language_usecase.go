package usecase

import (
	"context"

	"github.com/roysitumorang/bible/helper"
	languageModel "github.com/roysitumorang/bible/modules/language/model"
	languageQuery "github.com/roysitumorang/bible/modules/language/query"
	"go.uber.org/zap"
)

type (
	languageUseCase struct {
		languageQuery languageQuery.LanguageQuery
	}
)

func New(
	languageQuery languageQuery.LanguageQuery,
) LanguageUseCase {
	return &languageUseCase{
		languageQuery: languageQuery,
	}
}

func (q *languageUseCase) FindLanguages(ctx context.Context) (response []languageModel.Language, err error) {
	ctxt := "LanguageUseCase-FindLanguages"
	if response, err = q.languageQuery.FindLanguages(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindLanguages")
	}
	return
}
