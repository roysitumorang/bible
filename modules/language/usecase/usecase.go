package usecase

import (
	"context"

	languageModel "github.com/roysitumorang/bible/modules/language/model"
)

type (
	LanguageUseCase interface {
		FindLanguages(ctx context.Context) (response []languageModel.Language, err error)
		Sync(ctx context.Context) (err error)
		CreateIndex(ctx context.Context) (err error)
		ReIndex(ctx context.Context) (err error)
		DeleteIndex(ctx context.Context) (err error)
	}
)
