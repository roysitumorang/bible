package presenter

import (
	"github.com/gofiber/fiber/v2"
	"github.com/roysitumorang/bible/helper"
	languageModel "github.com/roysitumorang/bible/modules/language/model"
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

// swagger:operation GET /languages Language FindLanguages
// Get languages
// ---
// produces:
//   - "application/json"
// parameters:
// responses:
//   200:
//     description: "successful operation"
//     schema:
//       $ref: "#/definitions/ResponseLanguages"
func (q *languageHTTPHandler) FindLanguages(c *fiber.Ctx) error {
	ctx := c.UserContext()
	ctxt := "LanguagePresenter-FindLanguages"
	languages, err := q.languageUseCase.FindLanguages(ctx)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindLanguages")
		return languageModel.NewResponseLanguages(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	versions, err := q.versionUseCase.FindVersions(ctx, nil)
	if err != nil {
		helper.Log(ctx, zap.ErrorLevel, err.Error(), ctxt, "ErrFindVersions")
		return languageModel.NewResponseLanguages(fiber.StatusBadRequest, err.Error(), nil).WriteResponse(c)
	}
	mapLanguageVersions := map[string][]*versionModel.Version{}
	for _, version := range versions {
		mapLanguageVersions[version.LanguageUID] = append(mapLanguageVersions[version.LanguageUID], version)
	}
	for i, language := range languages {
		if versions, ok := mapLanguageVersions[language.UID]; ok {
			language.Versions = versions
		}
		languages[i] = language
	}
	return languageModel.NewResponseLanguages(fiber.StatusOK, "", languages).WriteResponse(c)
}
