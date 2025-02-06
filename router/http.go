package router

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/middleware/rewrite"
	"github.com/joho/godotenv"
	"github.com/roysitumorang/bible/config"
	"github.com/roysitumorang/bible/helper"
	"github.com/roysitumorang/bible/middleware"
	languagePresenter "github.com/roysitumorang/bible/modules/language/presenter"
	versePresenter "github.com/roysitumorang/bible/modules/verse/presenter"
	versionPresenter "github.com/roysitumorang/bible/modules/version/presenter"
	"go.uber.org/zap"
)

const (
	DefaultPort uint16 = 8080
)

func (q *Service) HTTPServerMain(ctx context.Context) error {
	ctxt := "Router-HTTPServerMain"
	app := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return helper.NewResponse(ctx, code, err.Error()).WriteResponse(ctx, nil)
		},
	})
	app.Use(
		recover.New(recover.Config{
			EnableStackTrace: true,
		}),
		fiberzap.New(fiberzap.Config{
			Logger: helper.GetLogger(),
		}),
		requestid.New(),
		compress.New(),
		rewrite.New(rewrite.Config{
			Rules: map[string]string{},
		}),
		cors.New(),
	)
	if helper.GetEnv() == "development" {
		app.Use(swagger.New(swagger.Config{
			BasePath: "/",
			FilePath: "./swagger.json",
			Path:     "docs",
			Title:    "API documentation",
			CacheAge: 0,
		}))
	}
	v1 := app.Group("/v1")
	languagePresenter.New(q.LanguageUseCase, q.VersionUseCase).Mount(v1.Group("/languages"))
	versionPresenter.New(q.VersionUseCase, q.BookUseCase).Mount(v1.Group("/versions"))
	versePresenter.New(q.BookUseCase, q.VerseUseCase).Mount(v1.Group("/verses"))
	v1.Get("/ping", func(c *fiber.Ctx) error {
		return helper.NewResponse(c, fiber.StatusOK, "").WriteResponse(c, map[string]interface{}{
			"version": config.Version,
			"commit":  config.Commit,
			"build":   config.Build,
			"upsince": config.Now.Format(time.RFC3339),
			"uptime":  time.Since(config.Now).String(),
		})
	})
	v1.Use(middleware.BasicAuth()).
		Get("/metrics", monitor.New(monitor.Config{
			APIOnly: true,
		})).
		Get("/env", func(c *fiber.Ctx) error {
			envMap, err := godotenv.Read(".env")
			if err != nil {
				helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrRead")
				return helper.NewResponse(c, fiber.StatusBadRequest, err.Error()).WriteResponse(c, nil)
			}
			envMap["GO_VERSION"] = runtime.Version()
			return helper.NewResponse(c, fiber.StatusOK, "").WriteResponse(c, envMap)
		})
	app.Use(func(c *fiber.Ctx) error {
		return helper.NewResponse(c, fiber.StatusNotFound, "").WriteResponse(c, nil)
	})
	port := DefaultPort
	if envPort, ok := os.LookupEnv("PORT"); ok && envPort != "" {
		if portInt, _ := strconv.Atoi(envPort); portInt >= 0 && portInt <= math.MaxUint16 {
			port = uint16(portInt)
		}
	}
	listenerPort := fmt.Sprintf(":%d", port)
	err := app.Listen(listenerPort)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrListen")
	}
	return err
}
