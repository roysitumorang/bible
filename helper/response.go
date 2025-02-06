package helper

import (
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	APP = "bible"
)

type (
	Response struct {
		RequestID  string      `json:"request_id" example:"add11106-20af-40c2-becb-d1a063e00e77"`
		RequestURL string      `json:"request_url" example:"GET http://localhost:18000/v1/languages"`
		StatusCode int         `json:"status_code" example:"200"`
		Message    string      `json:"message,omitempty" example:""`
		Status     string      `json:"status" example:"OK"`
		Timestamp  time.Time   `json:"timestamp" example:"2025-02-06T16:44:47.444931371+07:00"`
		Latency    string      `json:"latency" example:"7.746177ms"`
		Data       interface{} `json:"data,omitempty"`
		App        string      `json:"app" example:"bible"`
	}
)

func NewResponse(statusCode int, message string, data interface{}) *Response {
	return &Response{
		StatusCode: statusCode,
		Message:    message,
		Status:     http.StatusText(statusCode),
		Timestamp:  time.Now(),
		Data:       data,
		App:        APP,
	}
}

func (r *Response) WriteResponse(c *fiber.Ctx) error {
	if r.StatusCode == fiber.StatusNoContent {
		return c.SendStatus(r.StatusCode)
	}
	var builder strings.Builder
	_, _ = builder.WriteString(c.Method())
	_, _ = builder.WriteString(" ")
	_, _ = builder.Write(c.Request().URI().FullURI())
	r.RequestURL = builder.String()
	r.RequestID = ByteSlice2String(c.Response().Header.Peek(fiber.HeaderXRequestID))
	r.Latency = time.Since(c.Context().Time()).String()
	return c.Status(r.StatusCode).JSON(r)
}
