package model

import (
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/roysitumorang/bible/helper"
	bookModel "github.com/roysitumorang/bible/modules/book/model"
)

type (
	Filter struct {
		VersionUID string
	}

	FilterOption func(q *Filter)

	Version struct {
		ID int64 `json:"-"`
		// example: 0194d944-2957-7215-8f15-2b12282fa896
		UID         string `json:"id"`
		LanguageUID string `json:"-"`
		// example: 21st Century King James Version
		Name string `json:"name"`
		// example: KJ21
		Code string `json:"code"`
		// example: 21st-Century-King-James-Version-KJ21-Bible
		Slug      string           `json:"slug"`
		CreatedAt time.Time        `json:"-"`
		UpdatedAt time.Time        `json:"-"`
		Books     []bookModel.Book `json:"books,omitempty"`
	}

	// swagger:model ResponseVersion
	ResponseVersion struct {
		RequestID  string    `json:"request_id"`
		RequestURL string    `json:"request_url"`
		StatusCode int       `json:"status_code"`
		Message    string    `json:"message,omitempty"`
		Status     string    `json:"status"`
		Timestamp  time.Time `json:"timestamp"`
		Latency    string    `json:"latency"`
		Data       *Version  `json:"data,omitempty"`
		App        string    `json:"app"`
	}
)

func NewResponseVersion(statusCode int, message string, data *Version) *ResponseVersion {
	return &ResponseVersion{
		StatusCode: statusCode,
		Message:    message,
		Status:     http.StatusText(statusCode),
		Timestamp:  time.Now(),
		Data:       data,
		App:        helper.APP,
	}
}

func (r *ResponseVersion) WriteResponse(c *fiber.Ctx) error {
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

func WithVersionUID(versionUID string) FilterOption {
	return func(q *Filter) {
		q.VersionUID = versionUID
	}
}
