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
		Chapter,
		VerseNoStart,
		VerseNoEnd int
	}

	Verse struct {
		ID          int64     `json:"-"`
		UID         string    `json:"id" example:"07py2kdcg0004"`
		BookUID     string    `json:"-"`
		Chapter     int       `json:"chapter_no,omitempty" example:"1"`
		Number      int       `json:"verse_no" example:"1"`
		Body        string    `json:"body" example:"In the beginning God created the heaven and the earth."`
		CreatedAt   time.Time `json:"-"`
		UpdatedAt   time.Time `json:"-"`
		BookName    string    `json:"book,omitempty" example:"Genesis"`
		VersionCode string    `json:"version,omitempty" example:"KJ21"`
	}

	Chapter struct {
		Number int     `json:"number"`
		Verses []Verse `json:"verses"`
	}

	Passage struct {
		BookName     string  `json:"book" example:"Genesis"`
		Chapter      int     `json:"chapter" example:"1"`
		VerseNoStart int     `json:"verse_start" example:"1"`
		VerseNoEnd   int     `json:"verse_end" example:"20"`
		Verses       []Verse `json:"verses"`
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

func WithBook(name string, chapter, verseNoStart, verseNoEnd int) FilterOption {
	return func(q *Filter) {
		q.Books = append(
			q.Books,
			Book{
				Name:         name,
				Chapter:      chapter,
				VerseNoStart: verseNoStart,
				VerseNoEnd:   verseNoEnd,
			},
		)
	}
}

func (q Verse) Doc() Verse {
	q.Chapter = 0
	q.BookName = ""
	q.VersionCode = ""
	return q
}
