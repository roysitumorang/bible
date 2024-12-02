package query

import (
	"context"

	versionModel "github.com/roysitumorang/bible/modules/version/model"
)

type (
	VersionQuery interface {
		FindVersions(ctx context.Context, filter *versionModel.Filter) (response []versionModel.Version, err error)
	}
)
