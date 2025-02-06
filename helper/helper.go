package helper

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/gofiber/fiber/v2"
	"github.com/vishal-bihani/go-tsid"
)

type (
	contextKey string
)

var (
	timeZone *time.Location
	env,
	elasticUsername,
	elasticPassword string
	elasticAddress *url.URL
	indexPassage   string
	indexBatchSize int
	InitHelper     = sync.OnceValue(func() (err error) {
		location, ok := os.LookupEnv("TIME_ZONE")
		if !ok || location == "" {
			return errors.New("env TIME_ZONE is required")
		}
		if timeZone, err = time.LoadLocation(location); err != nil {
			return
		}
		if env, ok = os.LookupEnv("ENV"); !ok {
			return errors.New("env ENV is required")
		}
		if env == "" {
			env = "development"
		}
		if elasticUsername, ok = os.LookupEnv("ELASTIC_USERNAME"); !ok || elasticUsername == "" {
			return errors.New("env ELASTIC_USERNAME is required")
		}
		if elasticPassword, ok = os.LookupEnv("ELASTIC_PASSWORD"); !ok || elasticPassword == "" {
			return errors.New("env ELASTIC_PASSWORD is required")
		}
		envElasticAddress, ok := os.LookupEnv("ELASTIC_ADDRESS")
		if !ok || envElasticAddress == "" {
			return errors.New("env ELASTIC_ADDRESS is required")
		}
		if elasticAddress, err = url.Parse(envElasticAddress); err != nil {
			return
		}
		if indexPassage, ok = os.LookupEnv("INDEX_PASSAGE"); !ok || indexPassage == "" {
			return errors.New("env INDEX_PASSAGE is required")
		}
		envIndexBatchSize, ok := os.LookupEnv("INDEX_BATCH_SIZE")
		if !ok || envIndexBatchSize == "" {
			err = errors.New("env INDEX_BATCH_SIZE is required")
		}
		if indexBatchSize, _ = strconv.Atoi(envIndexBatchSize); indexBatchSize < 1 {
			err = errors.New("env INDEX_BATCH_SIZE requires a positive integer")
		}
		return
	})
)

func String2ByteSlice(str string) []byte {
	return unsafe.Slice(unsafe.StringData(str), len(str))
}

func ByteSlice2String(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}

func GenerateUniqueID() (internalID int64, externalID string) {
	tsid := tsid.Fast()
	return tsid.ToNumber(), tsid.ToLowerCase()
}

func GetContext(ctx context.Context, c *fiber.Ctx) context.Context {
	if requestID := c.Get(fiber.HeaderXRequestID); requestID != "" {
		ctx = context.WithValue(ctx, contextKey(fiber.HeaderXRequestID), requestID)
	}
	return ctx
}

func LoadTimeZone() *time.Location {
	return timeZone
}

func GetEnv() string {
	return env
}

func GetElasticUsername() string {
	return elasticUsername
}

func GetElasticPassword() string {
	return elasticPassword
}

func GetElasticAddress() *url.URL {
	return elasticAddress
}

func GetIndexPassage() string {
	return indexPassage
}

func GetIndexBatchSize() int {
	return indexBatchSize
}
