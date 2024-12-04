package usecase

import (
	"context"

	bookModel "github.com/roysitumorang/bible/modules/book/model"
)

type (
	BookUseCase interface {
		FindBooks(ctx context.Context, filter *bookModel.Filter) (response []bookModel.Book, err error)
	}
)
