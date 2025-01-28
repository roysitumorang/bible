package model

import (
	"time"
)

type (
	Testament struct {
		ID        int64     `json:"-"`
		UID       string    `json:"id"`
		Name      string    `json:"name"`
		Code      string    `json:"code"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
)
