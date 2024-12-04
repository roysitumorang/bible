package model

import (
	"time"
)

type (
	Filter struct {
		VersionCode string
		Books       []Book
	}

	FilterOption func(q *Filter)

	Book struct {
		Name string
		ChapterStart,
		ChapterEnd int
	}

	Verse struct {
		ID        int64     `json:"-"`
		UID       string    `json:"id"`
		BookUID   string    `json:"-"`
		Chapter   int       `json:"-"`
		Number    int       `json:"number"`
		Body      string    `json:"body"`
		CreatedAt time.Time `json:"-"`
		UpdatedAt time.Time `json:"-"`
		BookName  string    `json:"-"`
	}

	Chapter struct {
		Number int     `json:"number"`
		Verses []Verse `json:"verses"`
	}

	Passage struct {
		BookName     string    `json:"book"`
		ChapterStart int       `json:"chapter_start"`
		ChapterEnd   int       `json:"chapter_end"`
		Chapters     []Chapter `json:"chapters"`
	}
)

func NewFilter(options ...FilterOption) *Filter {
	filter := &Filter{}
	for _, option := range options {
		option(filter)
	}
	return filter
}

func WithVersionCode(versionCode string) FilterOption {
	return func(q *Filter) {
		q.VersionCode = versionCode
	}
}

func WithBook(name string, chapterStart, chapterEnd int) FilterOption {
	return func(q *Filter) {
		q.Books = append(
			q.Books,
			Book{
				Name:         name,
				ChapterStart: chapterStart,
				ChapterEnd:   chapterEnd,
			},
		)
	}
}
