package query

import (
	"context"

	"github.com/jackc/pgx/v5"
	languageModel "github.com/roysitumorang/bible/modules/language/model"
)

type (
	LanguageQuery interface {
		FindLanguages(ctx context.Context) (response []languageModel.Language, err error)
		SaveLanguage(ctx context.Context, tx pgx.Tx, request *languageModel.Language) (response string, err error)
	}
)
