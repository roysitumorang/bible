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
		// example: edc014eb-514b-4b87-87f9-4f568de7025f
		RequestID string `json:"request_id"`
		// example: "GET http://localhost:18000/v1/languages"
		RequestURL string `json:"request_url"`
		// example: 200
		StatusCode int `json:"status_code"`
		// example:
		Message string `json:"message,omitempty"`
		// example: OK
		Status string `json:"status"`
		// example: 2025-02-06T09:07:00.509782488+07:00
		Timestamp time.Time `json:"timestamp"`
		// example: 6.029109ms
		Latency string      `json:"latency"`
		Data    interface{} `json:"data,omitempty"`
		// example: bible
		App string `json:"app"`
	}
)

func NewResponse(c *fiber.Ctx, statusCode int, message string) *Response {
	var builder strings.Builder
	_, _ = builder.WriteString(c.Method())
	_, _ = builder.WriteString(" ")
	_, _ = builder.Write(c.Request().URI().FullURI())
	r := &Response{
		RequestID:  ByteSlice2String(c.Response().Header.Peek(fiber.HeaderXRequestID)),
		RequestURL: builder.String(),
		StatusCode: statusCode,
		Message:    message,
		Status:     http.StatusText(statusCode),
		Timestamp:  time.Now(),
		Latency:    time.Since(c.Context().Time()).String(),
		App:        APP,
	}
	return r
}

func (r *Response) WriteResponse(c *fiber.Ctx, data interface{}) error {
	if r.StatusCode == fiber.StatusNoContent {
		return c.SendStatus(r.StatusCode)
	}
	c.Status(r.StatusCode)
	if data == nil {
		return c.JSON(r)
	}
	return c.JSON(data)
}
