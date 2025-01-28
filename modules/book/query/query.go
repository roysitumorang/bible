package query

import (
	"context"

	"github.com/jackc/pgx/v5"
	bookModel "github.com/roysitumorang/bible/modules/book/model"
)

type (
	BookQuery interface {
		FindBooks(ctx context.Context, filter *bookModel.Filter) (response []bookModel.Book, err error)
		SaveBook(ctx context.Context, tx pgx.Tx, request bookModel.Book) (response string, err error)
	}
)
