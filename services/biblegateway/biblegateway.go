package biblegateway

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	strip "github.com/grokify/html-strip-tags-go"
	"github.com/roysitumorang/bible/helper"
	bookModel "github.com/roysitumorang/bible/modules/book/model"
	languageModel "github.com/roysitumorang/bible/modules/language/model"
	testamentModel "github.com/roysitumorang/bible/modules/testament/model"
	verseModel "github.com/roysitumorang/bible/modules/verse/model"
	versionModel "github.com/roysitumorang/bible/modules/version/model"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type (
	BibleGateway struct {
	}
)

const (
	baseURL = "https://www.biblegateway.com"
)

func New() *BibleGateway {
	return &BibleGateway{}
}

func (q *BibleGateway) Sync(ctx context.Context, testaments []testamentModel.Testament) (response []languageModel.Language, err error) {
	ctxt := "BibleGateway-Sync"
	var oldTestamentUID, newTestamentUID string
	for _, testament := range testaments {
		switch testament.Code {
		case "OT":
			oldTestamentUID = testament.UID
		case "NT":
			newTestamentUID = testament.UID
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
		return nil, fmt.Errorf("status code error: %d", statusCode)
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrNewDocumentFromReader")
		return
	}
	codeReplacer := strings.NewReplacer("(", "", ")", "")
	doc.Find("span.language-display").Each(func(i int, s *goquery.Selection) {
		spanID, ok := s.Attr("id")
		if !ok {
			return
		}
		languageID := strings.ReplaceAll(spanID, "lang-", "")
		parts := strings.Split(s.Nodes[0].NextSibling.Data, " ")
		language := languageModel.Language{
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
						versionModel.Version{
							Name: versionName,
							Code: versionCode,
							Slug: strings.TrimSuffix(strings.TrimPrefix(versionSlug, "/versions/"), "/#booklist"),
						},
					)
				}
			}
		})
		if language.Code == "EN" && len(language.Versions) > 0 {
			response = append(response, language)
		}
	})
	for i, language := range response {
		for j, version := range language.Versions {
			var builder strings.Builder
			_, _ = builder.WriteString(baseURL)
			_, _ = builder.WriteString("/versions/")
			_, _ = builder.WriteString(version.Slug)
			_, _ = builder.WriteString("/#booklist")
			statusCode, body, err := fasthttp.Get(nil, builder.String())
			if err != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrGet")
				return nil, err
			}
			if statusCode != fasthttp.StatusOK {
				return nil, fmt.Errorf("status code error: %d", statusCode)
			}
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
			if err != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrNewDocumentFromReader")
				return nil, err
			}
			doc.Find("tr.ot-book > td.book-name").Each(func(i int, s *goquery.Selection) {
				chaptersCount, err := strconv.Atoi(s.Children().Last().Text())
				if err != nil {
					helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrAtoi")
					return
				}
				bookName := strings.TrimSpace(s.Children().Nodes[1].NextSibling.Data)
				book := bookModel.Book{
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
				book := bookModel.Book{
					TestamentUID:  newTestamentUID,
					Name:          bookName,
					ChaptersCount: chaptersCount,
				}
				version.Books = append(version.Books, book)
			})
			language.Versions[j] = version
		}
		response[i] = language
	}
	for i, language := range response {
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
						return nil, err
					}
					if statusCode != fasthttp.StatusOK {
						return nil, fmt.Errorf("status code error: %d", statusCode)
					}
					doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
					if err != nil {
						helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrNewDocumentFromReader")
						return nil, err
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
							verseModel.Verse{
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
		response[i] = language
	}
	return
}
