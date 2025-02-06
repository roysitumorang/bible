package query

import (
	"context"

	"github.com/jackc/pgx/v5"
	versionModel "github.com/roysitumorang/bible/modules/version/model"
)

type (
	VersionQuery interface {
		FindVersions(ctx context.Context, filter *versionModel.Filter) (response []*versionModel.Version, err error)
		SaveVersion(ctx context.Context, tx pgx.Tx, request *versionModel.Version) (response string, err error)
	}
)
