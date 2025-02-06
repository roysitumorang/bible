package model

import (
	"time"

	bookModel "github.com/roysitumorang/bible/modules/book/model"
)

type (
	Filter struct {
		VersionUID string
	}

	FilterOption func(q *Filter)

	Version struct {
		ID          int64             `json:"-"`
		UID         string            `json:"id" example:"07pz6q6q80c2r"`
		LanguageUID string            `json:"-"`
		Name        string            `json:"name" example:"21st Century King James Version"`
		Code        string            `json:"code" example:"KJ21"`
		Slug        string            `json:"slug" example:"21st-Century-King-James-Version-KJ21-Bible"`
		CreatedAt   time.Time         `json:"-"`
		UpdatedAt   time.Time         `json:"-"`
		Books       []*bookModel.Book `json:"books,omitempty"`
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
