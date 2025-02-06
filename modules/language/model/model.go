package model

import (
	"time"

	"github.com/roysitumorang/bible/helper"
	versionModel "github.com/roysitumorang/bible/modules/version/model"
)

type (
	Language struct {
		ID int64 `json:"-"`
		// example: 0194d944-2954-7a1a-97f4-707afa02238b
		UID string `json:"id"`
		// example: English
		Name string `json:"name"`
		// example: EN
		Code      string                 `json:"code"`
		CreatedAt time.Time              `json:"-"`
		UpdatedAt time.Time              `json:"-"`
		Versions  []versionModel.Version `json:"versions"`
	}

	// swagger:model ResponseLanguages
	ResponseLanguages struct {
		*helper.Response
		Data []Language `json:"data"`
	}
)
