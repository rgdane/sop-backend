package dto

import (
	"jk-api/internal/database/models"
	"time"
)

// CreateSopDto is used when creating a new SOP.
type CreateSopDto struct {
	Name         string  `json:"name" binding:"required"`
	Description  *string `json:"description" binding:"required"`
	Code         string  `json:"code" binding:"required"`
	HasTitles    []int64 `json:"has_titles"`
	HasDivisions []int64 `json:"has_divisions"`
	ParentJobID  *int64  `json:"parent_job_id"`
}

// UpdateSopDto is used when updating an existing SOP.
type UpdateSopDto struct {
	Name         *string    `json:"name" binding:"required"`
	Description  *string    `json:"description" binding:"required"`
	Code         *string    `json:"code" binding:"required"`
	HasTitles    *[]int64   `json:"has_titles"`
	HasDivisions *[]int64   `json:"has_divisions"`
	DeletedAt    *time.Time `json:"deleted_at"`
	ParentJobID  *int64     `json:"parent_job_id"`
}

// SopResponseDto represents a detailed view of SOP with related data.
type SopResponseDto struct {
	models.Sop
}

type SopFilterDto struct {
	TitleID     int64
	DivisionID  int64
	DivisionIDs []int64
	Preload     bool
	Cursor      int64
	ShowDeleted bool
	Limit       int64
	Code        *string
	Name        string
	Restore     bool
	ExcludeID   int64
}

type BulkCreateSopsDto struct {
	Data []*CreateSopDto `json:"data" binding:"required"`
}

// BulkUpdateSopDto is used for bulk updating SOPs.
type BulkUpdateSopDto struct {
	IDs  []int64       `json:"ids" binding:"required"`
	Data *UpdateSopDto `json:"data"`
}

// BulkDeleteSopDto is used for bulk deleting SOPs.
type BulkDeleteSopDto struct {
	IDs []int64 `json:"ids" binding:"required"`
}
