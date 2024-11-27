package helper

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	topic   = "bible-service-log"
	service = "bible"
)

var (
	logger     *zap.Logger
	InitLogger = sync.OnceFunc(func() {
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.TimeKey = "timestamp"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		logger = zap.Must(zap.Config{
			Level:             zap.NewAtomicLevelAt(zap.InfoLevel),
			Development:       false,
			DisableCaller:     true,
			DisableStacktrace: true,
			Sampling:          nil,
			Encoding:          "json",
			EncoderConfig:     encoderCfg,
			OutputPaths: []string{
				"stderr",
			},
			ErrorOutputPaths: []string{
				"stderr",
			},
			InitialFields: map[string]interface{}{},
		}.Build())
	})
)

func GetLogger() *zap.Logger {
	return logger
}

func logContext(_ context.Context, context, scope string) *zap.Logger {
	defer func() {
		_ = logger.Sync()
	}()
	return logger.With(
		zap.String("topic", topic),
		zap.String("context", context),
		zap.String("scope", scope),
		zap.String("service", service),
	)
}

func Log(ctx context.Context, level zapcore.Level, message, context, scope string) {
	entry := logContext(ctx, context, scope)
	switch level {
	case zap.DebugLevel:
		entry.Debug(message)
	case zap.InfoLevel:
		entry.Info(message)
	case zap.WarnLevel:
		entry.Warn(message)
	case zap.ErrorLevel:
		var name string
		pc, file, line, _ := runtime.Caller(1)
		if fn := runtime.FuncForPC(pc); fn != nil {
			name = fn.Name()
		}
		entry.Error(
			message,
			zap.String("func", name),
			zap.String("file", file),
			zap.Int("line", line),
		)
	case zap.FatalLevel:
		entry.Fatal(message)
	case zap.PanicLevel:
		entry.Panic(message)
	}
}

func Capture(ctx context.Context, level zapcore.Level, err error, context, scope string) {
	entry := logContext(ctx, context, scope)
	switch level {
	case zap.DebugLevel:
		entry.Debug(err.Error())
	case zap.InfoLevel:
		entry.Info(err.Error())
	case zap.WarnLevel:
		entry.Warn(err.Error())
	case zap.ErrorLevel:
		// ignoring pgx.ErrNoRows
		if errors.Is(err, pgx.ErrNoRows) {
			return
		}
		var (
			name   string
			pgxErr *pgconn.PgError
		)
		pc, file, line, _ := runtime.Caller(1)
		if !errors.As(err, &pgxErr) &&
			!strings.HasSuffix(file, "_query.go") {
			pc, file, line, _ = runtime.Caller(4)
		}
		if fn := runtime.FuncForPC(pc); fn != nil {
			name = fn.Name()
		}
		entry.Error(
			err.Error(),
			zap.String("func", name),
			zap.String("file", file),
			zap.Int("line", line),
		)
	case zap.FatalLevel:
		entry.Fatal(err.Error())
	case zap.PanicLevel:
		entry.Panic(err.Error())
	}
}
