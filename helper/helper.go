package helper

import (
	"context"
	"errors"
	"math"
	"os"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/bwmarrin/snowflake"
	"github.com/gofiber/fiber/v2"
	"github.com/sqids/sqids-go"
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
	paginationHost string
	InitHelper = sync.OnceValue(func() (err error) {
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
		if paginationHost, ok = os.LookupEnv("PAGINATION_HOST"); !ok || paginationHost == "" {
			err = errors.New("env PAGINATION_HOST is required")
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

func GenerateUniqueID() (numericID int64, alphaNumericID string, err error) {
	numericID = snowflakeNode.Generate().Int64()
	alphaNumericID, err = EncodeSqIDs(uint64(numericID))
	return
}

func GetContext(ctx context.Context, c *fiber.Ctx) context.Context {
	type contextKey string
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

func GetPaginationHost() string {
	return paginationHost
}
