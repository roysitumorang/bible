package model

import (
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/roysitumorang/bible/helper"
)

type (
	Filter struct {
		VersionCode string
		Books       []Book
	}

	FilterOption func(q *Filter)

	Book struct {
		Name string
		Chapter,
		VerseNoStart,
		VerseNoEnd int
	}

	Verse struct {
		ID int64 `json:"-"`
		// example: 07py2kdcg0004
		UID     string `json:"id"`
		BookUID string `json:"-"`
		// example: 1
		Chapter int `json:"chapter_no,omitempty"`
		// example: 1
		Number int `json:"verse_no"`
		// example: In the beginning God created the heaven and the earth.
		Body      string    `json:"body"`
		CreatedAt time.Time `json:"-"`
		UpdatedAt time.Time `json:"-"`
		// example: Genesis
		BookName string `json:"book,omitempty"`
		// example: KJ21
		VersionCode string `json:"version,omitempty"`
	}

	Chapter struct {
		// example: 1
		Number int     `json:"number"`
		Verses []Verse `json:"verses"`
	}

	Passage struct {
		// example: Genesis
		BookName string `json:"book"`
		// example: 1
		Chapter int `json:"chapter"`
		// example: 1
		VerseNoStart int `json:"verse_start"`
		// example: 20
		VerseNoEnd int     `json:"verse_end"`
		Verses     []Verse `json:"verses"`
	}

	// swagger:model ResponsePassages
	ResponsePassages struct {
		// example: 2aebf1ec-40e1-4037-9660-d3397594f6bc
		RequestID string `json:"request_id"`
		// example: GET http://localhost:18000/v1/verses?version=KJ21&q=Genesis+1:1-20;Exodus+2:1-20
		RequestURL string `json:"request_url"`
		// example: 200
		StatusCode int `json:"status_code"`
		// example:
		Message string `json:"message,omitempty"`
		// example: OK
		Status string `json:"status"`
		// example: 2025-02-06T13:40:53.144679821+07:00
		Timestamp time.Time `json:"timestamp"`
		// example: 106.403552ms
		Latency string    `json:"latency"`
		Data    []Passage `json:"data"`
		// example: bible
		App string `json:"app"`
	}
)

func NewResponsePassages(statusCode int, message string, data []Passage) *ResponsePassages {
	return &ResponsePassages{
		StatusCode: statusCode,
		Message:    message,
		Status:     http.StatusText(statusCode),
		Timestamp:  time.Now(),
		Data:       data,
		App:        helper.APP,
	}
}

func (r *ResponsePassages) WriteResponse(c *fiber.Ctx) error {
	if r.StatusCode == fiber.StatusNoContent {
		return c.SendStatus(r.StatusCode)
	}
	var builder strings.Builder
	_, _ = builder.WriteString(c.Method())
	_, _ = builder.WriteString(" ")
	_, _ = builder.Write(c.Request().URI().FullURI())
	r.RequestURL = builder.String()
	r.RequestID = helper.ByteSlice2String(c.Response().Header.Peek(fiber.HeaderXRequestID))
	r.Latency = time.Since(c.Context().Time()).String()
	return c.Status(r.StatusCode).JSON(r)
}

func NewFilter(options ...FilterOption) *Filter {
	filter := &Filter{}
	for _, option := range options {
		option(filter)
	}
	return filter
}

func WithVersionCode(versionCode string) FilterOption {
	return func(q *Filter) {
		q.VersionCode = versionCode
	}
}

func WithBook(name string, chapter, verseNoStart, verseNoEnd int) FilterOption {
	return func(q *Filter) {
		q.Books = append(
			q.Books,
			Book{
				Name:         name,
				Chapter:      chapter,
				VerseNoStart: verseNoStart,
				VerseNoEnd:   verseNoEnd,
			},
		)
	}
}

func (q Verse) Doc() Verse {
	q.Chapter = 0
	q.BookName = ""
	q.VersionCode = ""
	return q
}
