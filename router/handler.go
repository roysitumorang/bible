package router

import (
	"context"

	"github.com/roysitumorang/bible/config"
	"github.com/roysitumorang/bible/helper"
	"github.com/roysitumorang/bible/migration"
	languageQuery "github.com/roysitumorang/bible/modules/language/query"
	languageUseCase "github.com/roysitumorang/bible/modules/language/usecase"
	versionQuery "github.com/roysitumorang/bible/modules/version/query"
	versionUseCase "github.com/roysitumorang/bible/modules/version/usecase"
	"github.com/roysitumorang/bible/services/biblegateway"
	"go.uber.org/zap"
)

type (
	Service struct {
		Migration       *migration.Migration
		BibleGateway    *biblegateway.BibleGateway
		LanguageUseCase languageUseCase.LanguageUseCase
		VersionUseCase  versionUseCase.VersionUseCase
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
	languageQuery := languageQuery.New(dbRead, dbWrite)
	versionQuery := versionQuery.New(dbRead, dbWrite)
	languageUseCase := languageUseCase.New(languageQuery)
	versionUseCase := versionUseCase.New(versionQuery)
	return &Service{
		Migration:       migration,
		BibleGateway:    bibleGateway,
		LanguageUseCase: languageUseCase,
		VersionUseCase:  versionUseCase,
	}, nil
}
