package usecase

import (
	"context"
	"errors"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/bible/helper"
	bookQuery "github.com/roysitumorang/bible/modules/book/query"
	languageModel "github.com/roysitumorang/bible/modules/language/model"
	languageQuery "github.com/roysitumorang/bible/modules/language/query"
	testamentQuery "github.com/roysitumorang/bible/modules/testament/query"
	verseQuery "github.com/roysitumorang/bible/modules/verse/query"
	versionQuery "github.com/roysitumorang/bible/modules/version/query"
	"github.com/roysitumorang/bible/services/alkitabtoba"
	"github.com/roysitumorang/bible/services/biblegateway"
	"go.uber.org/zap"
)

type (
	languageUseCase struct {
		testamentQuery testamentQuery.TestamentQuery
		languageQuery  languageQuery.LanguageQuery
		versionQuery   versionQuery.VersionQuery
		bookQuery      bookQuery.BookQuery
		verseQuery     verseQuery.VerseQuery
		biblegateway   *biblegateway.BibleGateway
		alkitabtoba    *alkitabtoba.AlkitabToba
	}
)

func New(
	testamentQuery testamentQuery.TestamentQuery,
	languageQuery languageQuery.LanguageQuery,
	versionQuery versionQuery.VersionQuery,
	bookQuery bookQuery.BookQuery,
	verseQuery verseQuery.VerseQuery,
	biblegateway *biblegateway.BibleGateway,
	alkitabtoba *alkitabtoba.AlkitabToba,
) LanguageUseCase {
	return &languageUseCase{
		testamentQuery: testamentQuery,
		languageQuery:  languageQuery,
		versionQuery:   versionQuery,
		bookQuery:      bookQuery,
		verseQuery:     verseQuery,
		biblegateway:   biblegateway,
		alkitabtoba:    alkitabtoba,
	}
}

func (q *languageUseCase) FindLanguages(ctx context.Context) (response []languageModel.Language, err error) {
	ctxt := "LanguageUseCase-FindLanguages"
	if response, err = q.languageQuery.FindLanguages(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindLanguages")
	}
	return
}

func (q *languageUseCase) Sync(ctx context.Context) (err error) {
	ctxt := "LanguageUseCase-Sync"
	testaments, err := q.testamentQuery.FindTestaments(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindTestaments")
		return
	}
	biblegatewayLanguages, err := q.biblegateway.Sync(ctx, testaments)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSync")
		return
	}
	alkitabtobaLanguages, err := q.alkitabtoba.Sync(ctx, testaments)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSync")
		return
	}
	languages := slices.Concat(biblegatewayLanguages, alkitabtobaLanguages)
	tx, err := q.languageQuery.BeginTx(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrBeginTx")
		return
	}
	defer func() {
		errRollback := tx.Rollback(ctx)
		if errors.Is(errRollback, pgx.ErrTxClosed) {
			errRollback = nil
		}
		if errRollback != nil {
			helper.Log(ctx, zap.ErrorLevel, errRollback.Error(), ctxt, "ErrRollback")
		}
	}()
	for _, language := range languages {
		languageUID, err := q.languageQuery.SaveLanguage(ctx, tx, language)
		if err != nil {
			helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSaveLanguage")
			return err
		}
		for _, version := range language.Versions {
			version.LanguageUID = languageUID
			versionUID, err := q.versionQuery.SaveVersion(ctx, tx, version)
			if err != nil {
				helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSaveVersion")
				return err
			}
			for _, book := range version.Books {
				book.VersionUID = versionUID
				bookUID, err := q.bookQuery.SaveBook(ctx, tx, book)
				if err != nil {
					helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSaveBook")
					return err
				}
				for _, verse := range book.Verses {
					verse.BookUID = bookUID
					if err = q.verseQuery.SaveVerse(ctx, tx, verse); err != nil {
						helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSaveVerse")
						return err
					}
				}
			}
		}
	}
	if err = tx.Commit(ctx); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCommit")
	}
	return
}
