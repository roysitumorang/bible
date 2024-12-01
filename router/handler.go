package router

import (
	"context"

	"github.com/roysitumorang/bible/config"
	"github.com/roysitumorang/bible/helper"
	"github.com/roysitumorang/bible/migration"
	"github.com/roysitumorang/bible/services/biblegateway"
	"go.uber.org/zap"
)

type (
	Service struct {
		Migration    *migration.Migration
		BibleGateway *biblegateway.BibleGateway
	}
)

func MakeHandler(ctx context.Context) (*Service, error) {
	ctxt := "Router-MakeHandler"
	dbRead, err := config.GetDbReadOnly(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrGetDbReadOnly")
		return nil, err
	}
	dbWrite, err := config.GetDbWriteOnly(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrGetDbWriteOnly")
		return nil, err
	}
	migration := migration.New(dbRead, dbWrite)
	bibleGateway := biblegateway.New(dbRead, dbWrite)
	return &Service{
		Migration:    migration,
		BibleGateway: bibleGateway,
	}, nil
}
