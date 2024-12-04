package biblegateway

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	strip "github.com/grokify/html-strip-tags-go"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/roysitumorang/bible/helper"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type (
	BibleGateway struct {
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
	baseURL = "https://www.biblegateway.com"
)

func New(
	dbRead,
	dbWrite *pgxpool.Pool,
) *BibleGateway {
	return &BibleGateway{
		dbRead:  dbRead,
		dbWrite: dbWrite,
	}
}

func (q *BibleGateway) Sync(ctx context.Context) (err error) {
	ctxt := "BibleGateway-Sync"
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
	var builder strings.Builder
	_, _ = builder.WriteString(baseURL)
	_, _ = builder.WriteString("/versions/")
	statusCode, body, err := fasthttp.Get(nil, builder.String())
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
	codeReplacer := strings.NewReplacer("(", "", ")", "")
	var languages []Language
	doc.Find("span.language-display").Each(func(i int, s *goquery.Selection) {
		spanID, ok := s.Attr("id")
		if !ok {
			return
		}
		languageID := strings.ReplaceAll(spanID, "lang-", "")
		parts := strings.Split(s.Nodes[0].NextSibling.Data, " ")
		language := Language{
			Name: s.Text(),
			Code: codeReplacer.Replace(parts[1]),
		}
		doc.Find(fmt.Sprintf("tr[data-language=%s] > td[data-translation]", languageID)).Children().Each(func(i int, s *goquery.Selection) {
			a := s.Find("a").First()
			if a == nil {
				return
			}
			if versionSlug, ok := a.Attr("href"); ok {
				parts := strings.Split(a.Text(), " ")
				n := len(parts) - 1
				versionCode, parts := codeReplacer.Replace(parts[n]), parts[:n]
				versionName := strings.Join(parts, " ")
				if versionCode == "KJ21" || versionCode == "ERV" {
					language.Versions = append(
						language.Versions,
						Version{
							Name: versionName,
							Code: versionCode,
							Slug: strings.TrimSuffix(strings.TrimPrefix(versionSlug, "/versions/"), "/#booklist"),
						},
					)
				}
			}
		})
		if language.Code == "EN" && len(language.Versions) > 0 {
			languages = append(languages, language)
		}
	})
	for i, language := range languages {
		for j, version := range language.Versions {
			var builder strings.Builder
			_, _ = builder.WriteString(baseURL)
			_, _ = builder.WriteString("/versions/")
			_, _ = builder.WriteString(version.Slug)
			_, _ = builder.WriteString("/#booklist")
			statusCode, body, err := fasthttp.Get(nil, builder.String())
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
			doc.Find("tr.ot-book > td.book-name").Each(func(i int, s *goquery.Selection) {
				chaptersCount, err := strconv.Atoi(s.Children().Last().Text())
				if err != nil {
					helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrAtoi")
					return
				}
				bookName := strings.TrimSpace(s.Children().Nodes[1].NextSibling.Data)
				book := Book{
					TestamentUID:  oldTestamentUID,
					Name:          bookName,
					ChaptersCount: chaptersCount,
				}
				version.Books = append(version.Books, book)
			})
			doc.Find("tr.nt-book > td.book-name").Each(func(i int, s *goquery.Selection) {
				chaptersCount, err := strconv.Atoi(s.Children().Last().Text())
				if err != nil {
					helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrAtoi")
					return
				}
				bookName := strings.TrimSpace(s.Children().Nodes[1].NextSibling.Data)
				book := Book{
					TestamentUID:  newTestamentUID,
					Name:          bookName,
					ChaptersCount: chaptersCount,
				}
				version.Books = append(version.Books, book)
			})
			language.Versions[j] = version
		}
		languages[i] = language
	}
	for i, language := range languages {
		for j, version := range language.Versions {
			for k, book := range version.Books {
				chapters := make([]int, book.ChaptersCount)
				for l := 0; l < book.ChaptersCount; l++ {
					chapters[l] = l + 1
				}
				for chunk := range slices.Chunk(chapters, 20) {
					firstChapter := chunk[0]
					var builder strings.Builder
					_, _ = builder.WriteString(book.Name)
					_, _ = builder.WriteString(" ")
					_, _ = builder.WriteString(strconv.Itoa(firstChapter))
					if n := len(chunk); n > 1 {
						lastChapter := chunk[n-1]
						_, _ = builder.WriteString("-")
						_, _ = builder.WriteString(strconv.Itoa(lastChapter))
					}
					query := url.Values{}
					query.Set("search", builder.String())
					query.Set("version", version.Code)
					builder.Reset()
					_, _ = builder.WriteString(baseURL)
					_, _ = builder.WriteString("/passage/?")
					_, _ = builder.WriteString(query.Encode())
					statusCode, body, err := fasthttp.Get(nil, builder.String())
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
					chapterNumber := firstChapter
					doc.Find("p.verse").Each(func(i int, s *goquery.Selection) {
						verseNumber := 1
						chapterNumberRaw := strings.ReplaceAll(s.Find("span.chapternum").Text(), "\u00a0", "")
						if chapterNumberRaw != "" {
							if chapterNumber, err = strconv.Atoi(chapterNumberRaw); err != nil {
								helper.Capture(ctx, zap.ErrorLevel, err, ctxt, builder.String())
								return
							}
						}
						verseNumberRaw := strings.ReplaceAll(s.Find("sup.versenum").Text(), "\u00a0", "")
						if verseNumberRaw != "" {
							if verseNumber, err = strconv.Atoi(verseNumberRaw); err != nil {
								helper.Capture(ctx, zap.ErrorLevel, err, ctxt, builder.String())
								return
							}
						}
						verseBody := strip.StripTags(s.Text())
						if verseNumber > 1 {
							verseBody = strings.TrimPrefix(verseBody, strconv.Itoa(verseNumber))
						} else {
							verseBody = strings.TrimPrefix(verseBody, strconv.Itoa(chapterNumber))
						}
						verseBody = strings.TrimSpace(verseBody)
						book.Verses = append(
							book.Verses,
							Verse{
								Chapter: chapterNumber,
								Number:  verseNumber,
								Body:    verseBody,
							},
						)
					})
				}
				version.Books[k] = book
			}
			language.Versions[j] = version
		}
		languages[i] = language
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
