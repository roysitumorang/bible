package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/roysitumorang/bible/helper"
)

func BasicAuth() func(c *fiber.Ctx) error {
	return basicauth.New(basicauth.Config{
		Users: map[string]string{
			os.Getenv("BASIC_AUTH_USERNAME"): os.Getenv("BASIC_AUTH_PASSWORD"),
		},
		Unauthorized: func(c *fiber.Ctx) error {
			return helper.NewResponse(fiber.StatusUnauthorized, "Unauthorized", nil).WriteResponse(c)
		},
	})
}
