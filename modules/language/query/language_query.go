package query

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/bible/helper"
	languageModel "github.com/roysitumorang/bible/modules/language/model"
	versionModel "github.com/roysitumorang/bible/modules/version/model"
	"go.uber.org/zap"
)

type (
	languageQuery struct {
		dbRead, dbWrite *pgxpool.Pool
	}
)

func New(
	dbRead,
	dbWrite *pgxpool.Pool,
) LanguageQuery {
	return &languageQuery{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *languageQuery) FindLanguages(ctx context.Context) (response []languageModel.Language, err error) {
	ctxt := "LanguageQuery-FindLanguages"
	response = make([]languageModel.Language, 0)
	var count int
	if err = q.dbRead.QueryRow(ctx, "SELECT COUNT(1) FROM languages").Scan(&count); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
		return
	}
	if count == 0 {
		return
	}
	response = make([]languageModel.Language, count)
	rows, err := q.dbRead.Query(
		ctx,
		`SELECT
			id
			, uid
			, name
			, code
			, created_at
			, updated_at
		FROM languages
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
	var i int
	for rows.Next() {
		language := response[i]
		language.Versions = make([]*versionModel.Version, 0)
		if err = rows.Scan(
			&language.ID,
			&language.UID,
			&language.Name,
			&language.Code,
			&language.CreatedAt,
			&language.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			return
		}
		response[i] = language
		i++
	}
	return
}

func (q *languageQuery) SaveLanguage(ctx context.Context, tx pgx.Tx, request *languageModel.Language) (response string, err error) {
	ctxt := "LanguageQuery-SaveLanguage"
	languageID, languageUID := helper.GenerateUniqueID()
	if err = tx.QueryRow(
		ctx,
		`INSERT INTO languages (
			id
			, uid
			, name
			, code
			, created_at
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (code) DO UPDATE SET
			name = $3
			, updated_at = $5
		RETURNING uid`,
		languageID,
		languageUID,
		request.Name,
		request.Code,
		time.Now(),
	).Scan(&response); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, zap.ErrorLevel, errRollback, ctxt, "ErrRollback")
		}
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
	}
	return
}
