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
	languageHTTPHandler struct {
		languageUseCase languageUseCase.LanguageUseCase
		versionUseCase  versionUseCase.VersionUseCase
	}
)

func New(
	languageUseCase languageUseCase.LanguageUseCase,
	versionUseCase versionUseCase.VersionUseCase,

) *languageHTTPHandler {
	return &languageHTTPHandler{
		languageUseCase: languageUseCase,
		versionUseCase:  versionUseCase,
	}
}

func (q *languageHTTPHandler) Mount(r fiber.Router) {
	r.Get("", q.FindLanguages)
}

func (q *languageHTTPHandler) FindLanguages(c *fiber.Ctx) error {
	ctx := context.Background()
	ctxt := "LanguagePresenter-FindLanguages"
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
