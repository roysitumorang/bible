package usecase

import (
	"context"

	verseModel "github.com/roysitumorang/bible/modules/verse/model"
)

type (
	VerseUseCase interface {
		FindVerses(ctx context.Context, filter *verseModel.Filter) (response []verseModel.Verse, err error)
	}
)
