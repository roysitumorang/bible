package helper

import (
	"github.com/gofiber/fiber/v2"
)

const (
	APP = "bible"
)

type (
	Response struct {
		Code    int         `json:"code"`
		Message string      `json:"message,omitempty"`
		Data    interface{} `json:"data,omitempty"`
		App     string      `json:"app"`
	}
)

func NewResponse(code int, msg string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: msg,
		Data:    data,
		App:     APP,
	}
}

func (r *Response) WriteResponse(c *fiber.Ctx) error {
	if r.Code == fiber.StatusNoContent {
		return c.SendStatus(r.Code)
	}
	return c.Status(r.Code).JSON(r)
}
