package query

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

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
	query :=
		`SELECT COUNT(1)
		FROM verses v1
		JOIN books b ON v1.book_uid = b.uid
		JOIN versions v2 ON b.version_uid = v2.uid`
	var params []interface{}
	if filter != nil {
		conditions := make([]string, 2)
		if filter.VersionCode == "" ||
			len(filter.Books) == 0 {
			return
		}
		params = append(params, filter.VersionCode)
		conditions[0] = fmt.Sprintf("v2.code = $%d", len(params))
		var builder strings.Builder
		subConditions := make([]string, len(filter.Books))
		for i, book := range filter.Books {
			builder.Reset()
			params = append(params, book.Name, book.Chapter, book.VerseNoStart, book.VerseNoEnd)
			n := len(params)
			_, _ = builder.WriteString("(b.name = $")
			_, _ = builder.WriteString(strconv.Itoa(n - 3))
			_, _ = builder.WriteString(" AND v1.chapter = $")
			_, _ = builder.WriteString(strconv.Itoa(n - 2))
			_, _ = builder.WriteString(" AND v1.number BETWEEN $")
			_, _ = builder.WriteString(strconv.Itoa(n - 1))
			_, _ = builder.WriteString(" AND $")
			_, _ = builder.WriteString(strconv.Itoa(n))
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
		, b.name
		, v2.code`,
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
			&verse.VersionCode,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			return
		}
		response[i] = verse
		i++
	}
	return
}

func (q *verseQuery) SaveVerse(ctx context.Context, tx pgx.Tx, request *verseModel.Verse) (err error) {
	ctxt := "VerseQuery-SaveVerse"
	verseID, verseUID := helper.GenerateUniqueID()
	if _, err = tx.Exec(
		ctx,
		`INSERT INTO verses (
			id
			, uid
			, book_uid
			, chapter
			, number
			, body
			, created_at
			, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		ON CONFLICT (number, chapter, book_uid) DO UPDATE SET
			body = $6
			, updated_at = $7`,
		verseID,
		verseUID,
		request.BookUID,
		request.Chapter,
		request.Number,
		request.Body,
		time.Now(),
	); err != nil {
		if errRollback := tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, zap.ErrorLevel, errRollback, ctxt, "ErrRollback")
		}
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
	}
	return
}

func (q *verseQuery) BeginTx(ctx context.Context) (response pgx.Tx, err error) {
	ctxt := "VerseQuery-BeginTx"
	if response, err = q.dbWrite.Begin(ctx); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrBegin")
	}
	return
}
