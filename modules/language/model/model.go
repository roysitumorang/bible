package model

import (
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/roysitumorang/bible/helper"
	versionModel "github.com/roysitumorang/bible/modules/version/model"
)

type (
	Language struct {
		ID int64 `json:"-"`
		// example: 0194d944-2954-7a1a-97f4-707afa02238b
		UID string `json:"id"`
		// example: English
		Name string `json:"name"`
		// example: EN
		Code      string                 `json:"code"`
		CreatedAt time.Time              `json:"-"`
		UpdatedAt time.Time              `json:"-"`
		Versions  []versionModel.Version `json:"versions"`
	}

	// swagger:model ResponseLanguages
	ResponseLanguages struct {
		RequestID  string     `json:"request_id"`
		RequestURL string     `json:"request_url"`
		StatusCode int        `json:"status_code"`
		Message    string     `json:"message,omitempty"`
		Status     string     `json:"status"`
		Timestamp  time.Time  `json:"timestamp"`
		Latency    string     `json:"latency"`
		Data       []Language `json:"data"`
		App        string     `json:"app"`
	}
)

func NewResponseLanguages(statusCode int, message string, data []Language) *ResponseLanguages {
	return &ResponseLanguages{
		StatusCode: statusCode,
		Message:    message,
		Status:     http.StatusText(statusCode),
		Timestamp:  time.Now(),
		Data:       data,
		App:        helper.APP,
	}
}

func (r *ResponseLanguages) WriteResponse(c *fiber.Ctx) error {
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
