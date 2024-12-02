package usecase

import (
	"context"

	"github.com/roysitumorang/bible/helper"
	versionModel "github.com/roysitumorang/bible/modules/version/model"
	versionQuery "github.com/roysitumorang/bible/modules/version/query"
	"go.uber.org/zap"
)

type (
	versionUseCase struct {
		versionQuery versionQuery.VersionQuery
	}
)

func New(
	versionQuery versionQuery.VersionQuery,
) VersionUseCase {
	return &versionUseCase{
		versionQuery: versionQuery,
	}
}

func (q *versionUseCase) FindVersions(ctx context.Context, filter *versionModel.Filter) (response []versionModel.Version, err error) {
	ctxt := "VersionUseCase-FindVersions"
	if response, err = q.versionQuery.FindVersions(ctx, filter); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrScan")
	}
	return
}
