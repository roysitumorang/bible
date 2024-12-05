package query

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/bible/helper"
	bookModel "github.com/roysitumorang/bible/modules/book/model"
	"go.uber.org/zap"
)

type (
	bookQuery struct {
		dbRead, dbWrite *pgxpool.Pool
	}
)

func New(
	dbRead,
	dbWrite *pgxpool.Pool,
) BookQuery {
	return &bookQuery{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *bookQuery) FindBooks(ctx context.Context, filter *bookModel.Filter) (response []bookModel.Book, err error) {
	ctxt := "BookQuery-FindBooks"
	response = make([]bookModel.Book, 0)
	if filter == nil {
		return
	}
	query := "SELECT COUNT(1) FROM books b JOIN versions v ON b.version_uid = v.uid"
	var (
		params     []interface{}
		builder    strings.Builder
		conditions []string
	)
	if filter.VersionUID != "" {
		params = append(params, filter.VersionUID)
		n := len(params)
		conditions = append(
			conditions,
			fmt.Sprintf("b.version_uid = $%d", n),
		)
	}
	if n := len(filter.Names); n > 0 {
		builder.Reset()
		_, _ = builder.WriteString("b.name IN (")
		for i, name := range filter.Names {
			params = append(params, name)
			if i > 0 {
				_, _ = builder.WriteString(",")
			}
			_, _ = builder.WriteString("$")
			_, _ = builder.WriteString(strconv.Itoa(len(params)))
		}
		_, _ = builder.WriteString(")")
		conditions = append(
			conditions,
			builder.String(),
		)
	}
	query = fmt.Sprintf("%s WHERE %s", query, strings.Join(conditions, " AND "))
	var count int
	if err = q.dbRead.QueryRow(ctx, query, params...).Scan(&count); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
		return
	}
	if count == 0 {
		return
	}
	response = make([]bookModel.Book, count)
	query = strings.ReplaceAll(
		query,
		"COUNT(1)",
		`b.id
		, b.uid
		, b.testament_uid
		, b.version_uid
		, b.name
		, b.chapters_count
		, b.created_at
		, b.updated_at
		, v.code`,
	)
	query = fmt.Sprintf("%s ORDER BY b.id", query)
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
		var (
			book        = response[i]
			versionCode string
		)
		if err = rows.Scan(
			&book.ID,
			&book.UID,
			&book.TestamentUID,
			&book.VersionUID,
			&book.Name,
			&book.ChaptersCount,
			&book.CreatedAt,
			&book.UpdatedAt,
			&versionCode,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			return
		}
		book.Chapters = make([]bookModel.Chapter, book.ChaptersCount)
		for j := 0; j < book.ChaptersCount; j++ {
			k := j + 1
			builder.Reset()
			_, _ = builder.WriteString(book.Name)
			_, _ = builder.WriteString(" ")
			_, _ = builder.WriteString(strconv.Itoa(k))
			query := url.Values{}
			query.Set("version", versionCode)
			query.Set("q", builder.String())
			builder.Reset()
			_, _ = builder.WriteString(filter.PaginationURL)
			_, _ = builder.WriteString(query.Encode())
			chapter := book.Chapters[j]
			chapter.Number = k
			chapter.Link = builder.String()
			book.Chapters[j] = chapter
		}
		response[i] = book
		i++
	}
	return
}
