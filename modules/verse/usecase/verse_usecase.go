package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/goccy/go-json"
	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/bible/helper"
	bookQuery "github.com/roysitumorang/bible/modules/book/query"
	languageModel "github.com/roysitumorang/bible/modules/language/model"
	languageQuery "github.com/roysitumorang/bible/modules/language/query"
	verseModel "github.com/roysitumorang/bible/modules/verse/model"
	verseQuery "github.com/roysitumorang/bible/modules/verse/query"
	versionQuery "github.com/roysitumorang/bible/modules/version/query"
	"github.com/roysitumorang/bible/services/alkitabtoba"
	"github.com/roysitumorang/bible/services/biblegateway"
	"github.com/roysitumorang/bible/services/elastic"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type (
	verseUseCase struct {
		languageQuery languageQuery.LanguageQuery
		versionQuery  versionQuery.VersionQuery
		bookQuery     bookQuery.BookQuery
		verseQuery    verseQuery.VerseQuery
		elastic       *elastic.Elastic
		indexName     string
		biblegateway  *biblegateway.BibleGateway
		alkitabtoba   *alkitabtoba.AlkitabToba
	}
)

func New(
	languageQuery languageQuery.LanguageQuery,
	versionQuery versionQuery.VersionQuery,
	bookQuery bookQuery.BookQuery,
	verseQuery verseQuery.VerseQuery,
	elastic *elastic.Elastic,
	indexName string,
	biblegateway *biblegateway.BibleGateway,
	alkitabtoba *alkitabtoba.AlkitabToba,
) VerseUseCase {
	return &verseUseCase{
		languageQuery: languageQuery,
		versionQuery:  versionQuery,
		bookQuery:     bookQuery,
		verseQuery:    verseQuery,
		elastic:       elastic,
		indexName:     indexName,
		biblegateway:  biblegateway,
		alkitabtoba:   alkitabtoba,
	}
}

func (q *verseUseCase) FindVerses(ctx context.Context, filter *verseModel.Filter) (response []verseModel.Verse, err error) {
	ctxt := "VerseUseCase-FindVerses"
	if response, err = q.verseQuery.FindVerses(ctx, filter); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindVerses")
	}
	return
}

func (q *verseUseCase) SearchVerses(ctx context.Context, filter *verseModel.Filter) (response []verseModel.Verse, err error) {
	ctxt := "VerseUseCase-SearchVerses"
	if filter == nil ||
		filter.VersionCode == "" ||
		len(filter.Books) == 0 {
		return
	}
	query := &types.Query{
		Bool: &types.BoolQuery{
			Should: make([]types.Query, len(filter.Books)),
		},
	}
	for i, book := range filter.Books {
		query.Bool.Should[i].Bool = &types.BoolQuery{
			Must: []types.Query{
				{
					Match: map[string]types.MatchQuery{
						"version": {Query: filter.VersionCode},
					},
				},
				{
					Match: map[string]types.MatchQuery{
						"book": {Query: book.Name},
					},
				},
				{
					Match: map[string]types.MatchQuery{
						"chapter_no": {Query: strconv.Itoa(book.Chapter)},
					},
				},
			},
		}
		if book.VerseNoEnd > book.VerseNoStart {
			chapterStart := types.Float64(book.VerseNoStart)
			chapterEnd := types.Float64(book.VerseNoEnd)
			query.Bool.Should[i].Bool.Must = append(
				query.Bool.Should[i].Bool.Must,
				types.Query{
					Range: map[string]types.RangeQuery{
						"verse_no": types.NumberRangeQuery{
							Gte: &chapterStart,
							Lte: &chapterEnd,
						},
					},
				},
			)
		} else {
			query.Bool.Should[i].Bool.Must = append(
				query.Bool.Should[i].Bool.Must,
				types.Query{
					Match: map[string]types.MatchQuery{
						"verse_no": {Query: strconv.Itoa(book.VerseNoStart)},
					},
				},
			)
		}
	}
	result, err := q.elastic.Search(
		ctx,
		q.indexName,
		&search.Request{
			Query: query,
		},
		0,
		10000,
		"id",
	)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSearch")
		return
	}
	n := len(result.Hits.Hits)
	if result.Hits.Total.Value == 0 ||
		n == 0 {
		return
	}
	var builder strings.Builder
	_, _ = builder.WriteString("[")
	for i, hit := range result.Hits.Hits {
		if i > 0 {
			_, _ = builder.WriteString(",")
		}
		_, _ = builder.Write(hit.Source_)
	}
	_, _ = builder.WriteString("]")
	if err = json.Unmarshal(helper.String2ByteSlice(builder.String()), &response); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrUnmarshal")
	}
	return
}

func (q *verseUseCase) CreateIndex(ctx context.Context) (err error) {
	ctxt := "VerseUseCase-CreateIndex"
	exists, err := q.elastic.IndexExists(ctx, q.indexName)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrIndexExists")
		return
	}
	if exists {
		return
	}
	if _, err = q.elastic.CreateIndex(
		ctx,
		q.indexName,
		&create.Request{
			Mappings: &types.TypeMapping{
				Properties: map[string]types.Property{
					"id":         types.NewKeywordProperty(),
					"version":    types.NewTextProperty(),
					"book":       types.NewTextProperty(),
					"chapter_no": types.NewIntegerNumberProperty(),
					"verse_no":   types.NewIntegerNumberProperty(),
					"verse":      types.NewTextProperty(),
				},
			},
		},
	); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrCreateIndex")
	}
	return
}

func (q *verseUseCase) ReIndex(ctx context.Context) (err error) {
	ctxt := "VerseUseCase-ReIndex"
	if err = q.CreateIndex(ctx); err != nil {
		return
	}
	verses, err := q.verseQuery.FindVerses(ctx, nil)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindVerses")
		return
	}
	total := len(verses)
	if total == 0 {
		helper.Log(ctx, zap.InfoLevel, "no verses to reindex", ctxt, "")
		return
	}
	helper.Log(ctx, zap.InfoLevel, fmt.Sprintf("%d verses will be reindexed", total), ctxt, "")
	batchSize := helper.GetIndexBatchSize()
	batches := int(math.Ceil(float64(total) / float64(batchSize)))
	var g errgroup.Group
	g.SetLimit(5)
	for i := 0; i < batches; i++ {
		g.Go(func() error {
			offsetMin := i * batchSize
			offsetMax := min(offsetMin+batchSize, total)
			if _, err = q.elastic.ReIndex(ctx, q.indexName, verses[offsetMin:offsetMax]); err != nil {
				helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrReIndex")
				return err
			}
			helper.Log(ctx, zap.InfoLevel, fmt.Sprintf("%d verses[%d:%d] were reindexed", offsetMax-offsetMin, offsetMin, offsetMax), ctxt, "")
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrWait")
		return
	}
	helper.Log(ctx, zap.InfoLevel, fmt.Sprintf("%d verses were reindexed", total), ctxt, "")
	return
}

func (q *verseUseCase) DeleteIndex(ctx context.Context) (err error) {
	ctxt := "VerseUseCase-DeleteIndex"
	exists, err := q.elastic.IndexExists(ctx, q.indexName)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrIndexExists")
		return
	}
	if !exists {
		return fmt.Errorf("index %s not found", q.indexName)
	}
	if _, err = q.elastic.DeleteIndex(ctx, q.indexName); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrDeleteIndex")
	}
	return
}

func (q *verseUseCase) Sync(ctx context.Context) (err error) {
	ctxt := "VerseUseCase-Sync"
	mapLanguages := sync.Map{}
	var g errgroup.Group
	g.Go(func() error {
		languages, err := q.biblegateway.Sync(ctx)
		if err != nil {
			helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSync")
			return err
		}
		for _, language := range languages {
			mapLanguages.Store(language.Code, language)
		}
		return nil
	})
	g.Go(func() error {
		languages, err := q.alkitabtoba.Sync(ctx)
		if err != nil {
			helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrSync")
			return err
		}
		for _, language := range languages {
			mapLanguages.Store(language.Code, language)
		}
		return nil
	})
	if err = g.Wait(); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrWait")
		return
	}
	var languages []*languageModel.Language
	mapLanguages.Range(func(_, value any) bool {
		languages = append(languages, value.(*languageModel.Language))
		return true
	})
	tx, err := q.verseQuery.BeginTx(ctx)
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
