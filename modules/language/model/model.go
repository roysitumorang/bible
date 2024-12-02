package model

import (
	"time"

	versionModel "github.com/roysitumorang/bible/modules/version/model"
)

type (
	Language struct {
		ID        int64                  `json:"-"`
		UID       string                 `json:"id"`
		Name      string                 `json:"name"`
		Code      string                 `json:"code"`
		CreatedAt time.Time              `json:"-"`
		UpdatedAt time.Time              `json:"-"`
		Versions  []versionModel.Version `json:"versions"`
	}
)
