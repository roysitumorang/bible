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
		ID        int64                   `json:"-"`
		UID       string                  `json:"id" example:"07pz6q6q80c2q"`
		Name      string                  `json:"name" example:"English"`
		Code      string                  `json:"code" example:"EN"`
		CreatedAt time.Time               `json:"-"`
		UpdatedAt time.Time               `json:"-"`
		Versions  []*versionModel.Version `json:"versions"`
	}

	ResponseLanguages struct {
		RequestID  string     `json:"request_id" example:"add11106-20af-40c2-becb-d1a063e00e77"`
		RequestURL string     `json:"request_url" example:"GET http://localhost:18000/v1/languages"`
		StatusCode int        `json:"status_code" example:"200"`
		Message    string     `json:"message,omitempty" example:""`
		Status     string     `json:"status" example:"OK"`
		Timestamp  time.Time  `json:"timestamp" example:"2025-02-06T16:44:47.444931371+07:00"`
		Latency    string     `json:"latency" example:"7.746177ms"`
		Data       []Language `json:"data"`
		App        string     `json:"app" example:"bible"`
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
