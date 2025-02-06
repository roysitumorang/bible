package model

import (
	"time"

	versionModel "github.com/roysitumorang/bible/modules/version/model"
)

type (
	Language struct {
		ID        int64                   `json:"-"`
		UID       string                  `json:"id" example:"07pz6q6q80c2q"`
		Name      string                  `json:"name" example:"English"`
		Code      string                  `json:"code" example:"EN"`
		CreatedAt time.Time               `json:"-"`
		UpdatedAt time.Time               `json:"-"`
		Versions  []*versionModel.Version `json:"versions"`
	}
)
