package dto

import (
	"jk-api/internal/database/models"
)

type CreateSopMenuDto struct {
	Name       string `json:"name"`
	SopID      *int64 `json:"sop_id"`
	DivisionID *int64 `json:"division_id"`
	Type       string `json:"type"`
	Multiple   bool   `json:"multiple"`
	ParentID   *int64 `json:"parent_id"`
	MasterID   *int64 `json:"master_id"`
	IsMaster   bool   `json:"is_master"`
}

type UpdateSopMenuDto struct {
	Name       *string `json:"name"`
	SopID      *int64  `json:"sop_id"`
	DivisionID *int64  `json:"division_id"`
	Type       *string `json:"type"`
	Multiple   *bool   `json:"multiple"`
	ParentID   *int64  `json:"parent_id"`
	MasterID   *int64  `json:"master_id"`
	IsMaster   *bool   `json:"is_master"`
}

type SopMenuFilterDto struct {
	IsCreateGraph bool
	ProjectID     int64
	DivisionID    int64
	Parent        bool
	MasterID      int64
	IsMaster      bool
	Preload       bool
	Type          string
	Name          string
}

type SopMenuResponseDto struct {
	models.SopMenu
}
