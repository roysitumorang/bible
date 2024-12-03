package model

import (
	"time"
)

type (
	Filter struct {
		VersionUID string
	}

	FilterOption func(q *Filter)

	Version struct {
		ID          int64     `json:"-"`
		UID         string    `json:"id"`
		LanguageUID string    `json:"-"`
		Name        string    `json:"name"`
		Code        string    `json:"code"`
		Slug        string    `json:"slug"`
		CreatedAt   time.Time `json:"-"`
		UpdatedAt   time.Time `json:"-"`
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
