package presenter

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/roysitumorang/bible/helper"
	versionModel "github.com/roysitumorang/bible/modules/version/model"
	versionUseCase "github.com/roysitumorang/bible/modules/version/usecase"
	"go.uber.org/zap"
)

type (
	versionHTTPHandler struct {
		versionUseCase versionUseCase.VersionUseCase
	}
)

func New(
	versionUseCase versionUseCase.VersionUseCase,
) *versionHTTPHandler {
	return &versionHTTPHandler{
		versionUseCase: versionUseCase,
	}
}

func (q *versionHTTPHandler) Mount(r fiber.Router) {
	r.Get("/:uid", q.FindVersion)
}

func (q *versionHTTPHandler) FindVersion(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "VersionPresenter-FindVersion"
	filter := versionModel.NewFilter(versionModel.WithVersionUID(c.Params("uid")))
	response, err := q.versionUseCase.FindVersions(ctx, filter)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindVersions")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	if len(response) == 0 {
		return helper.NewResponse(fiber.StatusNotFound, "version not found", nil).WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK, "", response[0]).WriteResponse(c)
}
