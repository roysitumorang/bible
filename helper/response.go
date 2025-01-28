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
		RequestID  string      `json:"request_id"`
		RequestURL string      `json:"request_url"`
		StatusCode int         `json:"status_code"`
		Message    string      `json:"message,omitempty"`
		Status     string      `json:"status"`
		Timestamp  time.Time   `json:"timestamp"`
		Data       interface{} `json:"data,omitempty"`
		App        string      `json:"app"`
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
	return c.Status(r.StatusCode).JSON(r)
}
