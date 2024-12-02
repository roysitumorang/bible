package usecase

import (
	"context"

	versionModel "github.com/roysitumorang/bible/modules/version/model"
)

type (
	VersionUseCase interface {
		FindVersions(ctx context.Context, filter *versionModel.Filter) (response []versionModel.Version, err error)
	}
)
