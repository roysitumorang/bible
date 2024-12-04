package alkitabtoba

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/bible/helper"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type (
	AlkitabToba struct {
		dbRead,
		dbWrite *pgxpool.Pool
	}

	Language struct {
		ID       int64     `json:"-"`
		UID      string    `json:"id"`
		Name     string    `json:"name"`
		Code     string    `json:"code"`
		Versions []Version `json:"versions"`
	}

	Version struct {
		ID    int64  `json:"-"`
		UID   string `json:"id"`
		Name  string `json:"name"`
		Code  string `json:"code"`
		Slug  string `json:"slug"`
		Books []Book `json:"books"`
	}

	Book struct {
		ID            int64   `json:"-"`
		UID           string  `json:"id"`
		TestamentUID  string  `json:"testament_id"`
		Name          string  `json:"name"`
		Slug          string  `json:"slug"`
		ChaptersCount int     `json:"chapters_count"`
		Verses        []Verse `json:"verses"`
	}

	Verse struct {
		ID      int64  `json:"-"`
		UID     string `json:"id"`
		Body    string `json:"body"`
		Chapter int    `json:"chapter"`
		Number  int    `json:"number"`
	}
)

const (
	baseURL = "https://alkitabtoba.wordpress.com/"
)

func New(
	dbRead,
	dbWrite *pgxpool.Pool,
) *AlkitabToba {
	return &AlkitabToba{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *AlkitabToba) Sync(ctx context.Context) (err error) {
	ctxt := "AlkitabToba-Sync"
	var oldTestamentUID, newTestamentUID string
	rows, err := q.dbRead.Query(ctx, "SELECT uid, code FROM testaments ORDER BY id")
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrQuery")
		return
	}
	defer rows.Close()
	for rows.Next() {
		var testamentUID, testamentCode string
		if err = rows.Scan(&testamentUID, &testamentCode); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			return
		}
		switch testamentCode {
		case "OT":
			oldTestamentUID = testamentUID
		case "NT":
			newTestamentUID = testamentUID
		}
	}
	statusCode, body, err := fasthttp.Get(nil, baseURL)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrGet")
		return
	}
	if statusCode != fasthttp.StatusOK {
		return fmt.Errorf("status code error: %d", statusCode)
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrNewDocumentFromReader")
		return
	}
	languages := []Language{
		{
			Name: "Bahasa Batak Toba",
			Code: "BBC",
			Versions: []Version{
				{
					Name: "Bahasa Batak Toba",
					Code: "BBC",
					Slug: "Bahasa-Batak-Toba",
				},
			},
		},
	}
	doc.Find("div[class='entry entry-content'] > p > a").Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}
		var testamentUID string
		if strings.Contains(href, "1-padan-na-robi") {
			testamentUID = oldTestamentUID
		} else if strings.Contains(href, "2-padan-na-imbaru") {
			testamentUID = newTestamentUID
		}
		bookName := s.Text()
		languages[0].Versions[0].Books = append(
			languages[0].Versions[0].Books,
			Book{
				TestamentUID: testamentUID,
				Name:         bookName,
				Slug:         href,
			},
		)
	})
	chapterNumberPattern := regexp.MustCompile(`(\d+)$`)
	verseNumberPattern := regexp.MustCompile(`^\d+:(\d+)\.? `)
	for i, book := range languages[0].Versions[0].Books {
		statusCode, body, err := fasthttp.Get(nil, book.Slug)
		if err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrGet")
			return err
		}
		if statusCode != fasthttp.StatusOK {
			return fmt.Errorf("status code error: %d", statusCode)
		}
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrNewDocumentFromReader")
			return err
		}
		doc.Find("h2.entry-title").Each(func(j int, s *goquery.Selection) {
			book.ChaptersCount++
			chapterNumberRaw := chapterNumberPattern.FindString(s.Text())
			chapterNumber, err := strconv.Atoi(chapterNumberRaw)
			if err != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrAtoi")
				return
			}
			versesRaw := s.Next().Text()
			verses := strings.Split(versesRaw, "\n")
			if len(verses) == 1 && verses[0] == "" {
				return
			}
			for _, verse := range verses {
				verseBody := strings.TrimSpace(verse)
				if verseBody == "" {
					continue
				}
				matches := verseNumberPattern.FindStringSubmatch(verseBody)
				if len(matches) != 2 {
					continue
				}
				verseNumber, err := strconv.Atoi(matches[1])
				if err != nil {
					helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrAtoi")
					return
				}
				verseBody = strings.TrimPrefix(verseBody, matches[0])
				book.Verses = append(
					book.Verses,
					Verse{
						Body:    verseBody,
						Chapter: chapterNumber,
						Number:  verseNumber,
					},
				)
			}
		})
		languages[0].Versions[0].Books[i] = book
	}
	tx, err := q.dbWrite.Begin(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrBegin")
		return err
	}
	defer func() {
		errRollback := tx.Rollback(ctx)
		if errors.Is(errRollback, pgx.ErrTxClosed) {
			errRollback = nil
		}
		if errRollback != nil {
			helper.Capture(ctx, zap.ErrorLevel, errRollback, ctxt, "ErrRollback")
		}
	}()
	for _, language := range languages {
		if language.ID, language.UID, err = helper.GenerateUniqueID(); err != nil {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				helper.Capture(ctx, zap.ErrorLevel, errRollback, ctxt, "ErrRollback")
			}
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrGenerateUniqueID")
			return err
		}
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
			language.ID,
			language.UID,
			language.Name,
			language.Code,
			time.Now(),
		).Scan(&language.UID); err != nil {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				helper.Capture(ctx, zap.ErrorLevel, errRollback, ctxt, "ErrRollback")
			}
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return err
		}
		for _, version := range language.Versions {
			if version.ID, version.UID, err = helper.GenerateUniqueID(); err != nil {
				if errRollback := tx.Rollback(ctx); errRollback != nil {
					helper.Capture(ctx, zap.ErrorLevel, errRollback, ctxt, "ErrRollback")
				}
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrGenerateUniqueID")
				return err
			}
			if err = tx.QueryRow(
				ctx,
				`INSERT INTO versions (
					id
					, uid
					, language_uid
					, name
					, code
					, slug
					, created_at
					, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
				ON CONFLICT (code) DO UPDATE SET
					language_uid = $3
					, name = $4
					, slug = $6
					, updated_at = $7
				RETURNING uid`,
				version.ID,
				version.UID,
				language.UID,
				version.Name,
				version.Code,
				version.Slug,
				time.Now(),
			).Scan(&version.UID); err != nil {
				if errRollback := tx.Rollback(ctx); errRollback != nil {
					helper.Capture(ctx, zap.ErrorLevel, errRollback, ctxt, "ErrRollback")
				}
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
				return err
			}
			for _, book := range version.Books {
				if book.ID, book.UID, err = helper.GenerateUniqueID(); err != nil {
					if errRollback := tx.Rollback(ctx); errRollback != nil {
						helper.Capture(ctx, zap.ErrorLevel, errRollback, ctxt, "ErrRollback")
					}
					helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrGenerateUniqueID")
					return err
				}
				if err = tx.QueryRow(
					ctx,
					`INSERT INTO books (
						id
						, uid
						, testament_uid
						, version_uid
						, name
						, chapters_count
						, created_at
						, updated_at
					) VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
					ON CONFLICT (name, version_uid) DO UPDATE SET
						testament_uid = $3
						, chapters_count = $6
						, updated_at = $7
					RETURNING uid`,
					book.ID,
					book.UID,
					book.TestamentUID,
					version.UID,
					book.Name,
					book.ChaptersCount,
					time.Now(),
				).Scan(&book.UID); err != nil {
					if errRollback := tx.Rollback(ctx); errRollback != nil {
						helper.Capture(ctx, zap.ErrorLevel, errRollback, ctxt, "ErrRollback")
					}
					helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
					return err
				}
				for _, verse := range book.Verses {
					if verse.ID, verse.UID, err = helper.GenerateUniqueID(); err != nil {
						if errRollback := tx.Rollback(ctx); errRollback != nil {
							helper.Capture(ctx, zap.ErrorLevel, errRollback, ctxt, "ErrRollback")
						}
						helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrGenerateUniqueID")
						return err
					}
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
						verse.ID,
						verse.UID,
						book.UID,
						verse.Chapter,
						verse.Number,
						verse.Body,
						time.Now(),
					); err != nil {
						if errRollback := tx.Rollback(ctx); errRollback != nil {
							helper.Capture(ctx, zap.ErrorLevel, errRollback, ctxt, "ErrRollback")
						}
						helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
						return err
					}
				}
			}
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrCommit")
	}
	return
}
