package query

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/bible/helper"
	testamentModel "github.com/roysitumorang/bible/modules/testament/model"
	"go.uber.org/zap"
)

type (
	testamentQuery struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}
)

func New(
	dbRead,
	dbWrite *pgxpool.Pool,
) TestamentQuery {
	return &testamentQuery{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *testamentQuery) FindTestaments(ctx context.Context) (response []testamentModel.Testament, err error) {
	ctxt := "TestamentQuery-FindTestaments"
	response = make([]testamentModel.Testament, 0)
	rows, err := q.dbRead.Query(
		ctx,
		`SELECT
			id
			, uid
			, name
			, code
			, created_at
			, updated_at
		FROM testaments
		ORDER BY id`,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrQuery")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var testament testamentModel.Testament
		if err = rows.Scan(
			&testament.ID,
			&testament.UID,
			&testament.Name,
			&testament.Code,
			&testament.CreatedAt,
			&testament.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			return
		}
		response = append(response, testament)
	}
	return
}
