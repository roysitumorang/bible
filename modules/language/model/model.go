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
		// example: 07py2kdaw0001
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
		// example: 28ebf1ec-40e1-4037-9660-d3397594f6bc
		RequestID string `json:"request_id"`
		// example: GET http://localhost:18000/v1/languages
		RequestURL string `json:"request_url"`
		// example: 200
		StatusCode int `json:"status_code"`
		// example:
		Message string `json:"message,omitempty"`
		// example: OK
		Status string `json:"status"`
		// example: 2025-02-06T13:36:54.822354385+07:00
		Timestamp time.Time `json:"timestamp"`
		// example: 20.06289ms
		Latency string     `json:"latency"`
		Data    []Language `json:"data"`
		// example: bible
		App string `json:"app"`
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
