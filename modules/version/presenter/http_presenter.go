package presenter

import (
	"strings"

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

// swagger:operation GET /versions/{versionUid} Version FindVersion
// Get version by UID
// ---
// produces:
//   - "application/json"
// parameters:
//   - name: versionUid
//     in: path
//     required: true
//     type: string
// responses:
//   200:
//     description: "successful operation"
//     schema:
//       $ref: "#/definitions/ResponseVersion"
func (q *versionHTTPHandler) FindVersion(c *fiber.Ctx) error {
	ctx := c.UserContext()
	ctxt := "VersionPresenter-FindVersion"
	var builder strings.Builder
	_, _ = builder.WriteString(c.BaseURL())
	_, _ = builder.WriteString("/v1/verses")
	paginationURL := builder.String()
	filter := versionModel.NewFilter(versionModel.WithVersionUID(c.Params("uid")))
	versions, err := q.versionUseCase.FindVersions(ctx, filter)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindVersions")
		return helper.NewResponse(c, fiber.StatusBadRequest, err.Error()).WriteResponse(c, nil)
	}
	if len(versions) == 0 {
		return helper.NewResponse(c, fiber.StatusNotFound, "version not found").WriteResponse(c, nil)
	}
	version := versions[0]
	if version.Books, err = q.bookUseCase.FindBooks(
		ctx,
		bookModel.NewFilter(
			bookModel.WithVersionUID(version.UID),
			bookModel.WithPaginationURL(paginationURL),
		),
	); err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindBooks")
		return helper.NewResponse(c, fiber.StatusBadRequest, err.Error()).WriteResponse(c, nil)
	}
	r := helper.NewResponse(c, fiber.StatusOK, "")
	response := versionModel.ResponseVersion{
		Response: r,
		Data:     version,
	}
	return r.WriteResponse(c, response)
}
