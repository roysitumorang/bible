package model

import (
	"time"

	"github.com/roysitumorang/bible/helper"
	bookModel "github.com/roysitumorang/bible/modules/book/model"
)

type (
	Filter struct {
		VersionUID string
	}

	FilterOption func(q *Filter)

	Version struct {
		ID int64 `json:"-"`
		// example: 0194d944-2957-7215-8f15-2b12282fa896
		UID         string `json:"id"`
		LanguageUID string `json:"-"`
		// example: 21st Century King James Version
		Name string `json:"name"`
		// example: KJ21
		Code string `json:"code"`
		// example: 21st-Century-King-James-Version-KJ21-Bible
		Slug      string           `json:"slug"`
		CreatedAt time.Time        `json:"-"`
		UpdatedAt time.Time        `json:"-"`
		Books     []bookModel.Book `json:"books,omitempty"`
	}

	// swagger:model ResponseVersion
	ResponseVersion struct {
		*helper.Response
		Data Version `json:"data"`
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
