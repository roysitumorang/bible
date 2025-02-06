package alkitabtoba

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/roysitumorang/bible/helper"
	"github.com/roysitumorang/bible/models"
	bookModel "github.com/roysitumorang/bible/modules/book/model"
	languageModel "github.com/roysitumorang/bible/modules/language/model"
	verseModel "github.com/roysitumorang/bible/modules/verse/model"
	versionModel "github.com/roysitumorang/bible/modules/version/model"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

type (
	AlkitabToba struct{}
)

const (
	baseURL = "https://alkitabtoba.wordpress.com/"
)

func New() *AlkitabToba {
	return &AlkitabToba{}
}

func (q *AlkitabToba) Sync(ctx context.Context) (response []*languageModel.Language, err error) {
	ctxt := "AlkitabToba-Sync"
	statusCode, body, err := fasthttp.Get(nil, baseURL)
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
	response = []*languageModel.Language{
		{
			Name: "Bahasa Batak Toba",
			Code: "BBC",
			Versions: []*versionModel.Version{
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
		var testament string
		if strings.Contains(href, "1-padan-na-robi") {
			testament = models.OldTestament
		} else if strings.Contains(href, "2-padan-na-imbaru") {
			testament = models.NewTestament
		}
		bookName := s.Text()
		response[0].Versions[0].Books = append(
			response[0].Versions[0].Books,
			&bookModel.Book{
				Testament: testament,
				Name:      bookName,
				Slug:      href,
			},
		)
	})
	chapterNumberPattern := regexp.MustCompile(`(\d+)$`)
	verseNumberPattern := regexp.MustCompile(`^\d+:(\d+)\.? `)
	for i, book := range response[0].Versions[0].Books {
		statusCode, body, err := fasthttp.Get(nil, book.Slug)
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
					&verseModel.Verse{
						Body:    verseBody,
						Chapter: chapterNumber,
						Number:  verseNumber,
					},
				)
			}
		})
		response[0].Versions[0].Books[i] = book
	}
	return
}
