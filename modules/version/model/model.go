package model

import (
	"time"
)

type (
	Filter struct {
		VersionUID string
	}

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

func NewFilter() *Filter {
	return &Filter{}
}

func (q *Filter) WithVersionUID(versionUID string) *Filter {
	q.VersionUID = versionUID
	return q
}
