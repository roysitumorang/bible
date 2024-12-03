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
	InitHelper    = sync.OnceValue(func() (err error) {
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
		timeZone, err = time.LoadLocation(location)
		return
	})
	GetEnv = sync.OnceValue(func() string {
		env := os.Getenv("ENV")
		if env == "" {
			env = "development"
		}
		return env
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
