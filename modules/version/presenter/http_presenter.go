package presenter

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/roysitumorang/bible/helper"
	languageUseCase "github.com/roysitumorang/bible/modules/language/usecase"
	versionModel "github.com/roysitumorang/bible/modules/version/model"
	versionUseCase "github.com/roysitumorang/bible/modules/version/usecase"
	"go.uber.org/zap"
)

type (
	versionHTTPHandler struct {
		versionUseCase  versionUseCase.VersionUseCase
		languageUseCase languageUseCase.LanguageUseCase
	}
)

func New(
	versionUseCase versionUseCase.VersionUseCase,
	languageUseCase languageUseCase.LanguageUseCase,
) *versionHTTPHandler {
	return &versionHTTPHandler{
		versionUseCase:  versionUseCase,
		languageUseCase: languageUseCase,
	}
}

func (q *versionHTTPHandler) Mount(r fiber.Router) {
	r.Get("", q.FindVersions)
	r.Get("/:uid", q.FindVersion)
}

func (q *versionHTTPHandler) FindVersions(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "VersionPresenter-FindVersions"
	languages, err := q.languageUseCase.FindLanguages(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindLanguages")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	versions, err := q.versionUseCase.FindVersions(ctx, nil)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindVersions")
		return helper.NewResponse(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	mapLanguageVersions := map[string][]versionModel.Version{}
	for _, version := range versions {
		mapLanguageVersions[version.LanguageUID] = append(mapLanguageVersions[version.LanguageUID], version)
	}
	for i, language := range languages {
		if versions, ok := mapLanguageVersions[language.UID]; ok {
			language.Versions = versions
		}
		languages[i] = language
	}
	return helper.NewResponse(fiber.StatusOK, "", languages).WriteResponse(c)
}

func (q *versionHTTPHandler) FindVersion(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "VersionPresenter-FindVersion"
	filter := &versionModel.Filter{
		VersionUID: c.Params("uid"),
	}
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
