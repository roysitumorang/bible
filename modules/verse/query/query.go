package query

import (
	"context"

	"github.com/jackc/pgx/v5"
	verseModel "github.com/roysitumorang/bible/modules/verse/model"
)

type (
	VerseQuery interface {
		FindVerses(ctx context.Context, filter *verseModel.Filter) (response []verseModel.Verse, err error)
		SaveVerse(ctx context.Context, tx pgx.Tx, request verseModel.Verse) (err error)
		BeginTx(ctx context.Context) (response pgx.Tx, err error)
	}
)
