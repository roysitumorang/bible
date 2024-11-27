package migration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/roysitumorang/bible/helper"
	"go.uber.org/zap"
)

type (
	Migration struct {
		tx pgx.Tx
	}
)

var (
	Migrations = map[int64]func(ctx context.Context, tx pgx.Tx) error{}
)

func NewMigration(tx pgx.Tx) *Migration {
	return &Migration{
		tx: tx,
	}
}

func (m *Migration) Migrate(ctx context.Context) error {
	ctxt := "Migration-Migrate"
	var exists int
	err := m.tx.QueryRow(
		ctx,
		`SELECT COUNT(1)
		FROM information_schema.tables
		WHERE table_name = 'migrations'`,
	).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
		if errRollback := m.tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrRollback")
		}
		return err
	}
	if exists == 0 {
		if _, err := m.tx.Exec(
			ctx,
			`CREATE TABLE migrations (
				"version" bigint NOT NULL PRIMARY KEY
			)`,
		); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			if errRollback := m.tx.Rollback(ctx); errRollback != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrRollback")
			}
			return err
		}
	}
	rows, err := m.tx.Query(ctx, `SELECT "version" FROM "migrations" ORDER BY "version"`)
	if errors.Is(err, pgx.ErrNoRows) {
		err = nil
	}
	if err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrQuery")
		if errRollback := m.tx.Rollback(ctx); errRollback != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrRollback")
		}
		return err
	}
	defer rows.Close()
	mapVersions := map[int64]int{}
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrScan")
			if errRollback := m.tx.Rollback(ctx); errRollback != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrRollback")
			}
			return err
		}
		mapVersions[version] = 1
	}
	sortedVersions := make([]int64, len(Migrations))
	var i int
	for version := range Migrations {
		sortedVersions[i] = version
		i++
	}
	if len(sortedVersions) > 0 {
		sort.Slice(
			sortedVersions,
			func(i, j int) bool {
				return sortedVersions[i] < sortedVersions[j]
			},
		)
	}
	for _, version := range sortedVersions {
		if _, ok := mapVersions[version]; ok {
			continue
		}
		function, ok := Migrations[version]
		if !ok {
			err := fmt.Errorf("migration function for version %d not found", version)
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrOK")
			if errRollback := m.tx.Rollback(ctx); errRollback != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrRollback")
			}
			return err
		}
		if err := function(ctx, m.tx); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrFunction")
			if errRollback := m.tx.Rollback(ctx); errRollback != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrRollback")
			}
			return err
		}
		if _, err := m.tx.Exec(ctx, `INSERT INTO "migrations" ("version") VALUES ($1)`, version); err != nil {
			helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExec")
			if errRollback := m.tx.Rollback(ctx); errRollback != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrRollback")
			}
			return err
		}
	}
	if err := m.tx.Commit(ctx); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrCommit")
		return err
	}
	return nil
}

func (m *Migration) CreateMigrationFile(_ context.Context) error {
	now := time.Now().UTC().UnixNano()
	filepath := fmt.Sprintf("./migration/%d.go", now)
	content := fmt.Sprintf(
		`package migration

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func init() {
	Migrations[%d] = func(ctx context.Context, tx pgx.Tx) (err error) {
		ctxt := "Migration-%d"
		return
	}
}`,
		now,
		now,
	)
	return os.WriteFile(
		filepath,
		helper.String2ByteSlice(content),
		0600,
	)
}
