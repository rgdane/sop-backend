package dto

import (
	"jk-api/internal/database/models"
	"time"
)

type CreateSpkDto struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description" binding:"required"`
	Code        string  `json:"code" binding:"required"`
	HasTitles   []int64 `json:"has_title"`
}

type UpdateSpkDto struct {
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	Code        *string    `json:"code" binding:"required"`
	HasTitles   *[]int64   `json:"has_title"`
	DeletedAt   *time.Time `json:"deleted_at"`
}

type BulkCreateSpksDto struct {
	Data []*CreateSpkDto `json:"data" binding:"required"`
}

type BulkUpdateSpkDto struct {
	IDs  []int64       `json:"ids" binding:"required"`
	Data *UpdateSpkDto `json:"data" binding:"required"`
}

type BulkDeleteSpkDto struct {
	IDs []int64 `json:"ids" binding:"required"`
}

type SpkFilterDto struct {
	TitleIDs    int64
	Preload     bool
	Limit       int64
	Cursor      int64
	Name        string
	ShowDeleted bool
	Restore     bool
}

type SpkResponseDto struct {
	models.Spk
}
