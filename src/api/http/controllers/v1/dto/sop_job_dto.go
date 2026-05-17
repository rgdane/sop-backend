// sop_job_dto.go

package dto

import (
	"jk-api/internal/database/models"
)

// CreateSopJobDto is used when creating a new SopJob.
type CreateSopJobDto struct {
	Name        string  `json:"name"`
	Alias       string  `json:"alias"`
	Description *string `json:"description"`
	Type        *string `json:"type"`
	TitleID     *int64  `json:"title_id"`
	SopID       int64   `json:"sop_id"`
	ReferenceID *int64  `json:"reference_id"`
	Url         *string `json:"url"`
	IsPublished *bool   `json:"is_published"`
}

// UpdateSopJobDto is used when updating an existing SopJob.
type UpdateSopJobDto struct {
	Name        *string `json:"name"`
	Alias       *string `json:"alias"`
	Description *string `json:"description"`
	Type        *string `json:"type"`
	TitleID     *int64  `json:"title_id"`
	SopID       *int64  `json:"sop_id"`
	ReferenceID *int64  `json:"reference_id"`
	FlowchartID *int64  `json:"flowchart_id"`
	NextIndex   *int    `json:"next_index"`
	PrevIndex   *int    `json:"prev_index"`
	Url         *string `json:"url"`
	IsPublished *bool   `json:"is_published"`
	IsHide      *bool   `json:"is_hide"`
}

// SopJobFilterDto is used to filter SopJob queries.
type SopJobFilterDto struct {
	Preload       bool
	Type          *string  `json:"type"`           // filter by type if needed
	SopID         int64    `json:"sop_id"`         // filter by sop_id
	SopName       string   `json:"sop_name"`       // filter by SOP name
	TitleID       int64    `json:"title_id"`       // filter by title
	DivisionNames []string `json:"division_names"` // filter by division names (IT, Finance, etc.)
	Page          int64    `json:"page"`           // pagination page
	Limit         int64    `json:"limit"`          // pagination limit
	Name          string   `json:"name"`           // filter by name
	MinIndex      int      `json:"min_index"`      // filter by minimum index (index > X)
	ReferenceID   *int64   `json:"reference_id"`   // filter by reference (SPK/SOP ID)
	ReferenceType string   `json:"reference_type"` // filter by reference type (sop/spk)
	ShowDeleted   bool     `json:"show_deleted"`   // show deleted records
	Sort          string   `json:"sort"`           // sort field
	Order         string   `json:"order"`          // sort order
}

// SopJobResponseDto represents a detailed view of SopJob with related data.
type SopJobResponseDto struct {
	models.SopJob
	HasSop       *models.Sop   `json:"has_sop,omitempty"`
	HasTitle     *models.Title `json:"has_title,omitempty"`
	HasReference interface{}   `json:"has_reference,omitempty"`
}

type BulkCreateSopJobs struct {
	Data []*CreateSopJobDto `json:"data" binding:"required"`
}

type BulkUpdateSopJobDto struct {
	IDs  []int64          `json:"ids" binding:"required"`
	Data *UpdateSopJobDto `json:"data" binding:"required"`
}

type BulkDeleteSopJobDto struct {
	IDs []int64 `json:"ids" binding:"required"`
}

type ReorderSopJobDto struct {
	NewIndex int   `json:"new_index" validate:"required,min=1"`
	SopID    int64 `json:"sop_id" validate:"required"`
}
