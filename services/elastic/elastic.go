package elastic

import (
	"context"
	"fmt"
	"net/url"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/bulk"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/get"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/update"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/delete"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/goccy/go-json"
	"github.com/roysitumorang/bible/helper"
	verseModel "github.com/roysitumorang/bible/modules/verse/model"
	"go.uber.org/zap"
)

type (
	Elastic struct {
		client *elasticsearch.TypedClient
	}
)

func New(
	address *url.URL,
	username,
	password string,
) (*Elastic, error) {
	client, err := elasticsearch.NewTypedClient(
		elasticsearch.Config{
			Addresses: []string{address.String()},
			Username:  username,
			Password:  password,
		},
	)
	if err != nil {
		return nil, err
	}
	return &Elastic{
		client: client,
	}, nil
}

func (q *Elastic) Ping(ctx context.Context) (response bool, err error) {
	ctxt := "ElasticService-Ping"
	if response, err = q.client.Ping().Do(ctx); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrDo")
	}
	return
}

func (q *Elastic) IndexExists(ctx context.Context, indexName string) (bool, error) {
	ctxt := "ElasticService-IndexExists"
	exists, err := q.client.Indices.Exists(indexName).IsSuccess(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrDo")
	}
	return exists, err
}

func (q *Elastic) CreateIndex(ctx context.Context, indexName string, mappings *create.Request) (*create.Response, error) {
	ctxt := "ElasticService-CreateIndex"
	response, err := q.client.Indices.Create(indexName).Request(mappings).Do(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrDo")
	}
	return response, err
}

func (q *Elastic) ReIndex(ctx context.Context, indexName string, verses []verseModel.Verse) (*bulk.Response, error) {
	ctxt := "ElasticService-ReIndex"
	bulkIndexer := q.client.Bulk()
	for _, verse := range verses {
		if err := bulkIndexer.CreateOp(types.CreateOperation{Id_: &verse.UID}, verse); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrCreateOp")
			return nil, err
		}
	}
	response, err := bulkIndexer.Index(indexName).Do(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrDo")
	}
	return response, err
}

func (q *Elastic) FindDocByID(ctx context.Context, indexName, docID string) (*get.Response, error) {
	ctxt := "ElasticService-FindDocByID"
	response, err := q.client.Get(indexName, docID).Do(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrDo")
	}
	return response, err
}

func (q *Elastic) Search(ctx context.Context, indexName string, request *search.Request) (*search.Response, error) {
	ctxt := "ElasticService-Search"
	response, err := q.client.Search().Index(indexName).Request(request).Do(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrDo")
	}
	return response, err
}

func (q *Elastic) Update(ctx context.Context, indexName, docID string, doc interface{}) (*update.Response, error) {
	ctxt := "ElasticService-Update"
	docByte, err := json.Marshal(doc)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrMarshal")
		return nil, err
	}
	response, err := q.client.Update(indexName, docID).Request(
		&update.Request{
			Doc: docByte,
		},
	).Do(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrDo")
	}
	return response, err
}

func (q *Elastic) Delete(ctx context.Context, indexName string, docIDs ...string) (*bulk.Response, error) {
	ctxt := "ElasticService-Delete"
	n := len(docIDs)
	if n == 0 {
		return nil, nil
	}
	bulkIndexer := q.client.Bulk()
	for _, docID := range docIDs {
		if err := bulkIndexer.DeleteOp(types.DeleteOperation{Id_: &docID}); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrDeleteOp")
			return nil, err
		}
	}
	response, err := bulkIndexer.Index(indexName).Do(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrDo")
	}
	return response, err
}

func (q *Elastic) DeleteIndex(ctx context.Context, indexName string) (*delete.Response, error) {
	ctxt := "ElasticService-Delete"
	exists, err := q.client.Indices.Exists(indexName).IsSuccess(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrDo")
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("index %s not found", indexName)
	}
	response, err := q.client.Indices.Delete(indexName).Do(ctx)
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrDo")
	}
	return response, err
}
