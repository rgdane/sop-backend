package dto

import (
	"jk-api/internal/database/models"
	"time"
)

// CreateDivisionDto is used when creating a new Division.
type CreateDivisionDto struct {
	Name         string `json:"name"`
	Code         string `json:"code"`
	DepartmentID int64  `json:"department_id"`
}

// UpdateDivisionDto is used when updating an existing Division.
type UpdateDivisionDto struct {
	Name         *string    `json:"name"`
	Code         *string    `json:"code"`
	DepartmentID *int64     `json:"department_id"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

type BulkCreateDivisionDto struct {
	Data []*CreateDivisionDto `json:"data" binding:"required"`
}

type BulkUpdateDivisionDto struct {
	IDs  []int64            `json:"ids" binding:"required"`
	Data *UpdateDivisionDto `json:"data" binding:"required"`
}

type BulkDeleteDivisionDto struct {
	IDs []int64 `json:"ids" binding:"required"`
}

// DivisionResponseDto represents a detailed view of Division with related data.
type DivisionResponseDto struct {
	models.Division
}

type DivisionFilterDto struct {
	DepartmentID int64
	SopId        int64
	Preload      bool
	Name         string
	Sort         string
	Order        string
	Limit        int64
	Cursor       int64
	ShowDeleted  bool
	Restore      bool
}
