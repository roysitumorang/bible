package query

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/bible/helper"
	versionModel "github.com/roysitumorang/bible/modules/version/model"
	"go.uber.org/zap"
)

type (
	versionQuery struct {
		dbRead, dbWrite *pgxpool.Pool
	}
)

func New(
	dbRead,
	dbWrite *pgxpool.Pool,
) VersionQuery {
	return &versionQuery{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *versionQuery) FindVersions(ctx context.Context, filter *versionModel.Filter) (response []versionModel.Version, err error) {
	ctxt := "VersionQuery-FindVersions"
	response = make([]versionModel.Version, 0)
	query := "SELECT COUNT(1) FROM versions"
	var params []interface{}
	if filter != nil {
		var conditions []string
		if filter.VersionUID != "" {
			params = append(params, filter.VersionUID)
			n := len(params)
			conditions = append(
				conditions,
				fmt.Sprintf("(uid = $%d OR slug =$%d)", n, n),
			)
		}
		query = fmt.Sprintf("%s WHERE %s", query, strings.Join(conditions, " AND "))
	}
	var count int
	if err = q.dbRead.QueryRow(ctx, query, params...).Scan(&count); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
		return
	}
	if count == 0 {
		return
	}
	response = make([]versionModel.Version, count)
	query = strings.ReplaceAll(
		query,
		"COUNT(1)",
		`id
		, uid
		, language_uid
		, name
		, code
		, slug
		, created_at
		, updated_at`,
	)
	query = fmt.Sprintf("%s ORDER BY id", query)
	rows, err := q.dbRead.Query(ctx, query, params...)
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
		version := response[i]
		if err = rows.Scan(
			&version.ID,
			&version.UID,
			&version.LanguageUID,
			&version.Name,
			&version.Code,
			&version.Slug,
			&version.CreatedAt,
			&version.UpdatedAt,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			return
		}
		response[i] = version
		i++
	}
	return
}
