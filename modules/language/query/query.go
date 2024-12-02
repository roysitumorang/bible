package query

import (
	"context"

	languageModel "github.com/roysitumorang/bible/modules/language/model"
)

type (
	LanguageQuery interface {
		FindLanguages(ctx context.Context) (response []languageModel.Language, err error)
	}
)
