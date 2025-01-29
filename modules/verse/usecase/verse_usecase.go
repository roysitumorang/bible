package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/goccy/go-json"
	"github.com/roysitumorang/bible/helper"
	verseModel "github.com/roysitumorang/bible/modules/verse/model"
	verseQuery "github.com/roysitumorang/bible/modules/verse/query"
	"github.com/roysitumorang/bible/services/elastic"
	"go.uber.org/zap"
)

type (
	verseUseCase struct {
		verseQuery verseQuery.VerseQuery
		elastic    *elastic.Elastic
		indexName  string
	}
)

func New(
	verseQuery verseQuery.VerseQuery,
	elastic *elastic.Elastic,
	indexName string,
) VerseUseCase {
	return &verseUseCase{
		verseQuery: verseQuery,
		elastic:    elastic,
		indexName:  indexName,
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
			},
		}
		if book.ChapterEnd > book.ChapterStart {
			chapterStart := types.Float64(book.ChapterStart)
			chapterEnd := types.Float64(book.ChapterEnd)
			query.Bool.Should[i].Bool.Must = append(
				query.Bool.Should[i].Bool.Must,
				types.Query{
					Range: map[string]types.RangeQuery{
						"chapter_no": types.NumberRangeQuery{
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
						"chapter_no": {Query: strconv.Itoa(book.ChapterStart)},
					},
				},
			)
		}
	}
	limit := 100
	result, err := q.elastic.Search(ctx, q.indexName, &search.Request{
		Query: query,
		Size:  &limit,
	})
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
		return fmt.Errorf("index %s already exists", q.indexName)
	}
	if _, err = q.elastic.CreateIndex(
		ctx,
		q.indexName,
		&create.Request{
			Mappings: &types.TypeMapping{
				Properties: map[string]types.Property{
					"id":         types.NewTextProperty(),
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
	n := len(verses)
	if n == 0 {
		helper.Log(ctx, zap.InfoLevel, "no verses to reindex", ctxt, "")
		return
	}
	helper.Log(ctx, zap.InfoLevel, fmt.Sprintf("%d verses will be reindexed", n), ctxt, "")
	if _, err = q.elastic.ReIndex(ctx, q.indexName, verses); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrReIndex")
		return
	}
	helper.Log(ctx, zap.InfoLevel, fmt.Sprintf("%d verses were reindexed", n), ctxt, "")
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
