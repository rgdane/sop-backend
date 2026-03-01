package dto

import (
	"jk-api/internal/database/models"
	"time"
)

type CreateDatabaseNodeDto struct {
	Name     string `json:"name" validate:"required"`
	TableRef string `json:"table_ref" validate:"required"`
	GraphRef string `json:"graph_ref" validate:"required"`
}

type UpdateDatabaseNodeDto struct {
	Name      *string    `json:"name"`
	TableRef  *string    `json:"table_ref"`
	GraphRef  *string    `json:"graph_ref"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type DatabaseNodeFilter struct {
	Preload     bool
	Cursor      int64  `query:"cursor"`
	Limit       int64  `query:"limit"`
	Search      string `query:"search"`
	ShowDeleted bool
	Restore     bool
}

type DatabaseNodeResponseDto struct {
	models.DatabaseNode
}
