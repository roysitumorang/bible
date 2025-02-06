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
		ID          int64             `json:"-"`
		UID         string            `json:"id" example:"07pz6q6q80c2r"`
		LanguageUID string            `json:"-"`
		Name        string            `json:"name" example:"21st Century King James Version"`
		Code        string            `json:"code" example:"KJ21"`
		Slug        string            `json:"slug" example:"21st-Century-King-James-Version-KJ21-Bible"`
		CreatedAt   time.Time         `json:"-"`
		UpdatedAt   time.Time         `json:"-"`
		Books       []*bookModel.Book `json:"books,omitempty"`
	}

	ResponseVersion struct {
		RequestID  string    `json:"request_id" example:"c6e430c5-21a5-45bf-aca2-36ae06c1a45d"`
		RequestURL string    `json:"request_url" example:"GET http://localhost:18000/v1/versions/07pz6q6q80c2r"`
		StatusCode int       `json:"status_code" example:"200"`
		Message    string    `json:"message,omitempty" example:""`
		Status     string    `json:"status" example:"OK"`
		Timestamp  time.Time `json:"timestamp" example:"2025-02-06T16:49:47.815304721+07:00"`
		Latency    string    `json:"latency" example:"17.306935ms"`
		Data       *Version  `json:"data,omitempty"`
		App        string    `json:"app" example:"bible"`
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
