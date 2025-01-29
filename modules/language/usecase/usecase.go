package usecase

import (
	"context"

	languageModel "github.com/roysitumorang/bible/modules/language/model"
)

type (
	LanguageUseCase interface {
		FindLanguages(ctx context.Context) (response []languageModel.Language, err error)
	}
)
