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
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/middleware/rewrite"
	"github.com/joho/godotenv"
	"github.com/roysitumorang/bible/config"
	"github.com/roysitumorang/bible/helper"
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
	r := fiber.New(fiber.Config{
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return helper.NewResponse(code, err.Error(), nil).WriteResponse(ctx)
		},
	})
	r.Use(
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
	v1 := r.Group("/v1")
	languagePresenter.New(q.LanguageUseCase, q.VersionUseCase).Mount(v1.Group("/languages"))
	versionPresenter.New(q.VersionUseCase, q.BookUseCase).Mount(v1.Group("/versions"))
	versePresenter.New(q.BookUseCase, q.VerseUseCase).Mount(v1.Group("/verses"))
	v1.Get("/ping", func(c *fiber.Ctx) error {
		return helper.NewResponse(
			fiber.StatusOK,
			"",
			map[string]interface{}{
				"version": config.Version,
				"commit":  config.Commit,
				"build":   config.Build,
				"upsince": config.Now.Format(time.RFC3339),
				"uptime":  time.Since(config.Now).String(),
			},
		).WriteResponse(c)
	})
	v1.Use(basicauth.New(basicauth.Config{
		Users: map[string]string{
			os.Getenv("BASIC_AUTH_USERNAME"): os.Getenv("BASIC_AUTH_PASSWORD"),
		},
		Unauthorized: func(c *fiber.Ctx) error {
			return helper.NewResponse(fiber.StatusUnauthorized, "Unauthorized", nil).WriteResponse(c)
		},
	})).
		Get("/metrics", monitor.New(monitor.Config{
			APIOnly: true,
		})).
		Get("/env", func(c *fiber.Ctx) error {
			envMap, err := godotenv.Read(".env")
			if err != nil {
				helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrRead")
				return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
			}
			envMap["GO_VERSION"] = runtime.Version()
			return helper.NewResponse(fiber.StatusOK, "", envMap).WriteResponse(c)
		})
	port := DefaultPort
	if envPort, ok := os.LookupEnv("PORT"); ok && envPort != "" {
		if portInt, _ := strconv.Atoi(envPort); portInt >= 0 && portInt <= math.MaxUint16 {
			port = uint16(portInt)
		}
	}
	listenerPort := fmt.Sprintf(":%d", port)
	err := r.Listen(listenerPort)
	if err != nil {
		helper.Log(ctx, zap.FatalLevel, err.Error(), ctxt, "ErrListen")
	}
	return err
}
