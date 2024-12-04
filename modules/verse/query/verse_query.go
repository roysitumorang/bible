package query

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/bible/helper"
	verseModel "github.com/roysitumorang/bible/modules/verse/model"
	"go.uber.org/zap"
)

type (
	verseQuery struct {
		dbRead, dbWrite *pgxpool.Pool
	}
)

func New(
	dbRead,
	dbWrite *pgxpool.Pool,
) VerseQuery {
	return &verseQuery{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *verseQuery) FindVerses(ctx context.Context, filter *verseModel.Filter) (response []verseModel.Verse, err error) {
	ctxt := "VerseQuery-FindVerses"
	response = make([]verseModel.Verse, 0)
	query := "SELECT COUNT(1) FROM verses v1 JOIN books b ON v1.book_uid = b.uid"
	var params []interface{}
	if filter != nil {
		conditions := make([]string, 2)
		if filter.VersionCode == "" ||
			len(filter.Books) == 0 {
			return
		}
		params = append(params, filter.VersionCode)
		n := len(params)
		conditions[0] = fmt.Sprintf(
			`EXISTS(
					SELECT 1
					FROM versions v2
					WHERE v2.uid = b.version_uid
						AND v2.code = $%d
			)`,
			n,
		)
		var builder strings.Builder
		subConditions := make([]string, len(filter.Books))
		for i, book := range filter.Books {
			builder.Reset()
			params = append(params, book.Name, book.ChapterStart)
			n = len(params)
			_, _ = builder.WriteString("(b.name = $")
			_, _ = builder.WriteString(strconv.Itoa(n - 1))
			_, _ = builder.WriteString(" AND v1.chapter ")
			if book.ChapterEnd > 0 {
				params = append(params, book.ChapterEnd)
				n = len(params)
				_, _ = builder.WriteString("BETWEEN $")
				_, _ = builder.WriteString(strconv.Itoa(n - 1))
				_, _ = builder.WriteString(" AND $")
				_, _ = builder.WriteString(strconv.Itoa(n))
				_, _ = builder.WriteString(")")
			} else {
				_, _ = builder.WriteString("= $")
				_, _ = builder.WriteString(strconv.Itoa(n))
				_, _ = builder.WriteString(")")
			}
			subConditions[i] = builder.String()
		}
		conditions[1] = fmt.Sprintf("(%s)", strings.Join(subConditions, " OR "))
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
	response = make([]verseModel.Verse, count)
	query = strings.ReplaceAll(
		query,
		"COUNT(1)",
		`v1.id
		, v1.uid
		, v1.book_uid
		, v1.chapter
		, v1.number
		, v1.body
		, v1.created_at
		, v1.updated_at
		, b.name`,
	)
	query = fmt.Sprintf("%s ORDER BY v1.id", query)
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
		verse := response[i]
		if err = rows.Scan(
			&verse.ID,
			&verse.UID,
			&verse.BookUID,
			&verse.Chapter,
			&verse.Number,
			&verse.Body,
			&verse.CreatedAt,
			&verse.UpdatedAt,
			&verse.BookName,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			return
		}
		response[i] = verse
		i++
	}
	return
}
