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
		// example:07py2kdb40002
		UID         string `json:"id"`
		LanguageUID string `json:"-"`
		// example: 21st Century King James Version
		Name string `json:"name"`
		// example: KJ21
		Code string `json:"code"`
		// example: 21st-Century-King-James-Version-KJ21-Bible
		Slug      string            `json:"slug"`
		CreatedAt time.Time         `json:"-"`
		UpdatedAt time.Time         `json:"-"`
		Books     []*bookModel.Book `json:"books,omitempty"`
	}

	// swagger:model ResponseVersion
	ResponseVersion struct {
		// example: 2bebf1ec-40e1-4037-9660-d3397594f6bc
		RequestID string `json:"request_id"`
		// example: GET http://localhost:18000/v1/versions/07py2kdb40002
		RequestURL string `json:"request_url"`
		// example: 200
		StatusCode int `json:"status_code"`
		// example:
		Message string `json:"message,omitempty"`
		// example: OK
		Status string `json:"status"`
		// example: 2025-02-06T13:44:22.460901389+07:00
		Timestamp time.Time `json:"timestamp"`
		// example: 11.112834ms
		Latency string   `json:"latency"`
		Data    *Version `json:"data,omitempty"`
		// example: bible
		App string `json:"app"`
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
