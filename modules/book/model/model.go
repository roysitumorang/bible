package model

import (
	"time"
)

type (
	Filter struct {
		VersionUID string
	}

	FilterOption func(q *Filter)

	Book struct {
		ID            int64     `json:"-"`
		UID           string    `json:"id"`
		TestamentUID  string    `json:"-"`
		VersionUID    string    `json:"-"`
		Name          string    `json:"name"`
		ChaptersCount int       `json:"-"`
		CreatedAt     time.Time `json:"-"`
		UpdatedAt     time.Time `json:"-"`
		Chapters      []Chapter `json:"chapters"`
	}

	Chapter struct {
		Number int    `json:"number"`
		Link   string `json:"link"`
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
