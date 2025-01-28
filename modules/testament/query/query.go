package query

import (
	"context"

	testamentModel "github.com/roysitumorang/bible/modules/testament/model"
)

type (
	TestamentQuery interface {
		FindTestaments(ctx context.Context) (response []testamentModel.Testament, err error)
	}
)
