package usecase

import (
	"context"

	verseModel "github.com/roysitumorang/bible/modules/verse/model"
)

type (
	VerseUseCase interface {
		FindVerses(ctx context.Context, filter *verseModel.Filter) (response []verseModel.Verse, err error)
		CreateIndex(ctx context.Context) (err error)
		ReIndex(ctx context.Context) (err error)
		DeleteIndex(ctx context.Context) (err error)
		SearchVerses(ctx context.Context, filter *verseModel.Filter) (response []verseModel.Verse, err error)
	}
)
