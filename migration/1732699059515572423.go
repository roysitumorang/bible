package migration

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/bible/helper"
	"go.uber.org/zap"
)

func init() {
	Migrations[1732699059515572423] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-1732699059515572423"
		// languages
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE languages (
				id bigint NOT NULL PRIMARY KEY
				, uid character varying NOT NULL UNIQUE
				, name character varying NOT NULL
				, code character varying NOT NULL
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON languages (name);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON languages (code);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// testaments
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE testaments (
				id bigint NOT NULL PRIMARY KEY
				, uid character varying NOT NULL UNIQUE
				, name character varying NOT NULL
				, code character varying NOT NULL
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON testaments (name);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON testaments (code);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		testaments := [][]string{{"Old", "OT"}, {"New", "NT"}}
		for _, testament := range testaments {
			testamentName, testamentCode := testament[0], testament[1]
			testamentID, testamentUID, err := helper.GenerateUniqueID()
			if err != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrGenerateUniqueID")
				return err
			}
			if _, err = tx.Exec(
				ctx,
				`INSERT INTO testaments (
					id
					, uid
					, name
					, code
					, created_at
					, updated_at
				) VALUES ($1, $2, $3, $4, $5, $5);`,
				testamentID,
				testamentUID,
				testamentName,
				testamentCode,
				time.Now(),
			); err != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
				return err
			}
		}
		// versions
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE versions (
				id bigint NOT NULL PRIMARY KEY
				, uid character varying NOT NULL UNIQUE
				, language_uid character varying NOT NULL REFERENCES languages (uid) ON UPDATE CASCADE ON DELETE CASCADE
				, name character varying NOT NULL
				, code character varying NOT NULL
				, slug character varying NOT NULL
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON versions (language_uid);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON versions (name);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON versions (code);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON versions (slug);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// books
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE books (
				id bigint NOT NULL PRIMARY KEY
				, uid character varying NOT NULL UNIQUE
				, testament_uid character varying NOT NULL REFERENCES testaments (uid) ON UPDATE CASCADE ON DELETE CASCADE
				, version_uid character varying NOT NULL REFERENCES versions (uid) ON UPDATE CASCADE ON DELETE CASCADE
				, name character varying NOT NULL
				, chapters_count integer NOT NULL
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON books (testament_uid);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON books (version_uid);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON books (name, version_uid);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		// verses
		if _, err = tx.Exec(
			ctx,
			`CREATE TABLE verses (
				id bigint NOT NULL PRIMARY KEY
				, uid character varying NOT NULL UNIQUE
				, book_uid character varying NOT NULL REFERENCES books (uid)
				, chapter integer NOT NULL
				, number integer NOT NULL
				, body character varying NOT NULL
				, created_at timestamp with time zone NOT NULL
				, updated_at timestamp with time zone NOT NULL
			);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON verses (book_uid);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE INDEX ON verses (chapter);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			return
		}
		if _, err = tx.Exec(
			ctx,
			`CREATE UNIQUE INDEX ON verses (number, chapter, book_uid);`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
		}
		return
	}
}
