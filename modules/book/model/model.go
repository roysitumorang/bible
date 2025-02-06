package model

import (
	"time"

	verseModel "github.com/roysitumorang/bible/modules/verse/model"
)

type (
	Filter struct {
		VersionUID,
		PaginationURL string
		Names []string
	}

	FilterOption func(q *Filter)

	Book struct {
		ID int64 `json:"-"`
		// example: 07py2kdcc0003
		UID        string `json:"id"`
		Testament  string `json:"-"`
		VersionUID string `json:"-"`
		// example: Genesis
		Name          string              `json:"name"`
		Slug          string              `json:"-"`
		ChaptersCount int                 `json:"-"`
		CreatedAt     time.Time           `json:"-"`
		UpdatedAt     time.Time           `json:"-"`
		Chapters      []Chapter           `json:"chapters"`
		Verses        []*verseModel.Verse `json:"-"`
	}

	Chapter struct {
		// example: 1
		Number int `json:"number"`
		// example: http://localhost:18000/v1/verses?q=Genesis+1&version=KJ21
		Link string `json:"link"`
	}
)

func NewFilter(options ...FilterOption) *Filter {
	filter := &Filter{}
	for _, option := range options {
		option(filter)
	}
	return filter
}

func WithVersionUID(versionUID string) FilterOption {
	return func(q *Filter) {
		q.VersionUID = versionUID
	}
}

func WithPaginationURL(paginationURL string) FilterOption {
	return func(q *Filter) {
		q.PaginationURL = paginationURL
	}
}

func WithNames(names ...string) FilterOption {
	return func(q *Filter) {
		q.Names = names
	}
}
