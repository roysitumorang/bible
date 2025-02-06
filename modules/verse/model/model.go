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
		ID          int64     `json:"-"`
		UID         string    `json:"id" example:"07py2kdcg0004"`
		BookUID     string    `json:"-"`
		Chapter     int       `json:"chapter_no,omitempty" example:"1"`
		Number      int       `json:"verse_no" example:"1"`
		Body        string    `json:"body" example:"In the beginning God created the heaven and the earth."`
		CreatedAt   time.Time `json:"-"`
		UpdatedAt   time.Time `json:"-"`
		BookName    string    `json:"book,omitempty" example:"Genesis"`
		VersionCode string    `json:"version,omitempty" example:"KJ21"`
	}

	Chapter struct {
		Number int     `json:"number"`
		Verses []Verse `json:"verses"`
	}

	Passage struct {
		BookName     string  `json:"book" example:"Genesis"`
		Chapter      int     `json:"chapter" example:"1"`
		VerseNoStart int     `json:"verse_start" example:"1"`
		VerseNoEnd   int     `json:"verse_end" example:"20"`
		Verses       []Verse `json:"verses"`
	}

	ResponsePassages struct {
		RequestID  string    `json:"request_id" example:"60068eaa-1f89-4e31-80a5-66cc133e86fc"`
		RequestURL string    `json:"request_url" example:"GET http://localhost:18000/v1/verses?version=KJ21&q=Genesis+1:1-20;Exodus+2:1-20"`
		StatusCode int       `json:"status_code" example:"200"`
		Message    string    `json:"message,omitempty" example:""`
		Status     string    `json:"status" example:"OK"`
		Timestamp  time.Time `json:"timestamp" example:"2025-02-06T16:52:34.01591064+07:00"`
		Latency    string    `json:"latency" example:"110.247731ms"`
		Data       []Passage `json:"data"`
		App        string    `json:"app" example:"bible"`
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
