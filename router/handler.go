package router

import (
	"context"
	"errors"

	"github.com/roysitumorang/bible/config"
	"github.com/roysitumorang/bible/helper"
	"github.com/roysitumorang/bible/migration"
	bookQuery "github.com/roysitumorang/bible/modules/book/query"
	bookUseCase "github.com/roysitumorang/bible/modules/book/usecase"
	languageQuery "github.com/roysitumorang/bible/modules/language/query"
	languageUseCase "github.com/roysitumorang/bible/modules/language/usecase"
	testamentQuery "github.com/roysitumorang/bible/modules/testament/query"
	verseQuery "github.com/roysitumorang/bible/modules/verse/query"
	verseUseCase "github.com/roysitumorang/bible/modules/verse/usecase"
	versionQuery "github.com/roysitumorang/bible/modules/version/query"
	versionUseCase "github.com/roysitumorang/bible/modules/version/usecase"
	"github.com/roysitumorang/bible/services/alkitabtoba"
	"github.com/roysitumorang/bible/services/biblegateway"
	"github.com/roysitumorang/bible/services/elastic"
	"go.uber.org/zap"
)

type (
	Service struct {
		Migration       *migration.Migration
		LanguageUseCase languageUseCase.LanguageUseCase
		VersionUseCase  versionUseCase.VersionUseCase
		BookUseCase     bookUseCase.BookUseCase
		VerseUseCase    verseUseCase.VerseUseCase
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
	bibleGateway := biblegateway.New()
	alkitabToba := alkitabtoba.New()
	elastic, err := elastic.New(helper.GetElasticAddress(), helper.GetElasticUsername(), helper.GetElasticPassword())
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrNew")
		return nil, err
	}
	ok, err := elastic.Ping(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrPing")
		return nil, err
	}
	if !ok {
		return nil, errors.New("failure to ping elastic")
	}
	languageQuery := languageQuery.New(dbRead, dbWrite)
	versionQuery := versionQuery.New(dbRead, dbWrite)
	bookQuery := bookQuery.New(dbRead, dbWrite)
	verseQuery := verseQuery.New(dbRead, dbWrite)
	testamentQuery := testamentQuery.New(dbRead, dbWrite)
	languageUseCase := languageUseCase.New(testamentQuery, languageQuery, versionQuery, bookQuery, verseQuery, bibleGateway, alkitabToba)
	versionUseCase := versionUseCase.New(versionQuery)
	bookUseCase := bookUseCase.New(bookQuery)
	verseUseCase := verseUseCase.New(verseQuery, elastic, helper.GetIndexPassage())
	return &Service{
		Migration:       migration,
		LanguageUseCase: languageUseCase,
		VersionUseCase:  versionUseCase,
		BookUseCase:     bookUseCase,
		VerseUseCase:    verseUseCase,
	}, nil
}
