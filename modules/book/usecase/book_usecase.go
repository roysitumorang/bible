package usecase

import (
	"context"

	"github.com/roysitumorang/bible/helper"
	bookModel "github.com/roysitumorang/bible/modules/book/model"
	bookQuery "github.com/roysitumorang/bible/modules/book/query"
	"go.uber.org/zap"
)

type (
	bookUseCase struct {
		bookQuery bookQuery.BookQuery
	}
)

func New(
	bookQuery bookQuery.BookQuery,
) BookUseCase {
	return &bookUseCase{
		bookQuery: bookQuery,
	}
}

func (q *bookUseCase) FindBooks(ctx context.Context, filter *bookModel.Filter) (response []bookModel.Book, err error) {
	ctxt := "BookUseCase-FindBooks"
	if response, err = q.bookQuery.FindBooks(ctx, filter); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindBooks")
	}
	return
}
