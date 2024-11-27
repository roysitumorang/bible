package router

import (
	"context"

	"github.com/roysitumorang/bible/config"
	"github.com/roysitumorang/bible/helper"
	"github.com/roysitumorang/bible/migration"
	"go.uber.org/zap"
)

type (
	Service struct {
		Migration *migration.Migration
	}
)

func MakeHandler(ctx context.Context) (*Service, error) {
	ctxt := "Router-MakeHandler"
	_, err := config.GetDbReadOnly(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrGetDbReadOnly")
		return nil, err
	}
	dbWrite, err := config.GetDbWriteOnly(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrGetDbWriteOnly")
		return nil, err
	}
	tx, err := dbWrite.Begin(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrBegin")
		return nil, err
	}
	migration := migration.NewMigration(tx)
	return &Service{
		Migration: migration,
	}, nil
}
