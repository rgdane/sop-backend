package dto

import (
	"jk-api/internal/database/models"
)

type CreateSpkJobDto struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	SpkID       int64   `json:"spk_id"`
	SopID       *int64  `json:"sop_id"`
	TitleID     *int64  `json:"title_id"`
	Index       int     `json:"index"`
	Url         *string `json:"url"`
}

type UpdateSpkJobDto struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	SopID       *int64  `json:"sop_id"`
	TitleID     *int64  `json:"title_id"`
	Index       *int64  `json:"index"`
	FlowchartID *int64  `json:"flowchart_id"`
	NextIndex   *int    `json:"next_index"`
	PrevIndex   *int    `json:"prev_index"`
	Url         *string `json:"url"`
}

type BulkCreateSpkJobsDto struct {
	Data []*CreateSpkJobDto `json:"data" binding:"required"`
}

type BulkUpdateSpkJobDto struct {
	IDs  []int64          `json:"ids" binding:"required"`
	Data *UpdateSpkJobDto `json:"data" binding:"required"`
}

type BulkDeleteSpkJobDto struct {
	IDs []int64 `json:"ids" binding:"required"`
}

type SpkJobFilterDto struct {
	Preload bool
	SpkID   int64 `json:"spk_id"`
	TitleID int64 `json:"title_id"`
}

type SpkJobResponseDto struct {
	models.SpkJob
}

type ReorderSpkJobDto struct {
	NewIndex int   `json:"new_index" validate:"required,min=1"`
	SpkID    int64 `json:"spk_id" validate:"required"`
}
