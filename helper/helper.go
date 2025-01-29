package helper

import (
	"context"
	"errors"
	"math"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/bwmarrin/snowflake"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sqids/sqids-go"
)

type (
	contextKey string
)

const (
	letters                = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerCaseAlphanumerics = "abcdefghijklmnopqrstuvwxyz0123456789"
)

var (
	snowflakeNode *snowflake.Node
	sqIDs         *sqids.Sqids
	timeZone      *time.Location
	env,
	elasticUsername,
	elasticPassword string
	elasticAddress *url.URL
	indexPassage   string
	indexBatchSize int
	InitHelper     = sync.OnceValue(func() (err error) {
		if snowflakeNode, err = snowflake.NewNode(1); err != nil {
			return
		}
		envSqIDsMinLength, ok := os.LookupEnv("SQIDS_MIN_LENGTH")
		if !ok || envSqIDsMinLength == "" {
			return errors.New("env SQIDS_MIN_LENGTH requires a positive integer")
		}
		sqIDsMinLength, err := strconv.Atoi(envSqIDsMinLength)
		if err != nil {
			return
		}
		sqIDs, err = sqids.New(sqids.Options{
			Alphabet:  lowerCaseAlphanumerics,
			MinLength: min(uint8(sqIDsMinLength), math.MaxUint8),
		})
		if err != nil {
			return
		}
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

func EncodeSqIDs(numbers ...uint64) (string, error) {
	return sqIDs.Encode(numbers)
}

func GenerateUniqueID() (internalID int64, externalID string, err error) {
	uuidV7, err := uuid.NewV7()
	if err != nil {
		return
	}
	return snowflakeNode.Generate().Int64(), uuidV7.String(), nil
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
