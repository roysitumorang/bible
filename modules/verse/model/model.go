package model

import (
	"time"

	"github.com/roysitumorang/bible/helper"
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
		ID int64 `json:"-"`
		// example: 0194b25b-bb70-7385-a6c4-bd4a748b7064
		UID     string `json:"id"`
		BookUID string `json:"-"`
		Chapter int    `json:"chapter_no,omitempty"`
		// example: 1
		Number int `json:"verse_no"`
		// example: In the beginning God created the heaven and the earth.
		Body        string    `json:"body"`
		CreatedAt   time.Time `json:"-"`
		UpdatedAt   time.Time `json:"-"`
		BookName    string    `json:"book,omitempty"`
		VersionCode string    `json:"version,omitempty"`
	}

	Chapter struct {
		// example: 1
		Number int     `json:"number"`
		Verses []Verse `json:"verses"`
	}

	Passage struct {
		BookName     string  `json:"book"`
		Chapter      int     `json:"chapter"`
		VerseNoStart int     `json:"verse_start"`
		VerseNoEnd   int     `json:"verse_end"`
		Verses       []Verse `json:"verses"`
	}

	// swagger:model ResponsePassages
	ReponsePassages struct {
		*helper.Response
		Data []Passage `json:"data"`
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
