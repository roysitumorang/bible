package presenter

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/roysitumorang/bible/helper"
	bookModel "github.com/roysitumorang/bible/modules/book/model"
	bookUseCase "github.com/roysitumorang/bible/modules/book/usecase"
	versionModel "github.com/roysitumorang/bible/modules/version/model"
	versionUseCase "github.com/roysitumorang/bible/modules/version/usecase"
	"go.uber.org/zap"
)

type (
	versionHTTPHandler struct {
		versionUseCase versionUseCase.VersionUseCase
		bookUseCase    bookUseCase.BookUseCase
	}
)

func New(
	versionUseCase versionUseCase.VersionUseCase,
	bookUseCase bookUseCase.BookUseCase,
) *versionHTTPHandler {
	return &versionHTTPHandler{
		versionUseCase: versionUseCase,
		bookUseCase:    bookUseCase,
	}
}

func (q *versionHTTPHandler) Mount(r fiber.Router) {
	r.Get("/:uid", q.FindVersion)
}

func (q *versionHTTPHandler) FindVersion(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "VersionPresenter-FindVersion"
	filter := versionModel.NewFilter(versionModel.WithVersionUID(c.Params("uid")))
	versions, err := q.versionUseCase.FindVersions(ctx, filter)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindVersions")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	if len(versions) == 0 {
		return helper.NewResponse(fiber.StatusNotFound, "version not found", nil).WriteResponse(c)
	}
	response := versions[0]
	if response.Books, err = q.bookUseCase.FindBooks(ctx, bookModel.NewFilter(bookModel.WithVersionUID(response.UID))); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindBooks")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	return helper.NewResponse(fiber.StatusOK, "", response).WriteResponse(c)
}
