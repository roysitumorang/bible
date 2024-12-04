package query

import (
	"context"

	bookModel "github.com/roysitumorang/bible/modules/book/model"
)

type (
	BookQuery interface {
		FindBooks(ctx context.Context, filter *bookModel.Filter) (response []bookModel.Book, err error)
	}
)
