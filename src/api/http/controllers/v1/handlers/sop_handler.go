package handlers

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/api/http/controllers/v1/mapper"
	"jk-api/internal/database/models"
	"jk-api/internal/repository/graphdb"
	"jk-api/internal/service"
	"jk-api/internal/shared/helper"
	"time"
)

type SopHandler struct {
	Service service.SopService
}

func NewSopHandler(service service.SopService) *SopHandler {
	return &SopHandler{Service: service}
}

// --- CREATE ---
func (h *SopHandler) CreateSopHandler(input *dto.CreateSopDto) (*dto.SopResponseDto, error) {
	db := h.Service.GetDB().Begin()
	committed := false
	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
			panic(r)
		}
		if !committed {
			db.Rollback()
		}
	}()

	sopService := h.Service.WithTx(db)

	payload, err := mapper.CreateSopDtoToModel(input)
	if err != nil {
		return nil, err
	}

	createdData, err := sopService.CreateSop(payload)
	if err != nil {
		return nil, err
	}

	graphNode := &graphdb.SopNode{
		ID:          createdData.ID,
		Name:        createdData.Name,
		Code:        createdData.Code,
		Description: fmt.Sprintf("%v", createdData.Description),
		ParentJobID: createdData.ParentJobID,
		CreatedAt:   createdData.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   createdData.UpdatedAt.Format(time.RFC3339Nano),
	}
	if err := sopService.InsertGraphSop(graphNode); err != nil {
		return nil, fmt.Errorf("failed to sync to graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return mapper.SopModelToResponseDto(createdData)
}

func (h *SopHandler) CreateSopSqlHandler(input *dto.CreateSopDto) (*dto.SopResponseDto, error) {
	db := h.Service.GetDB().Begin()
	committed := false
	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
			panic(r)
		}
		if !committed {
			db.Rollback()
		}
	}()

	sopService := h.Service.WithTx(db)

	payload, err := mapper.CreateSopDtoToModel(input)
	if err != nil {
		return nil, err
	}

	createdData, err := sopService.CreateSop(payload)
	if err != nil {
		return nil, err
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return mapper.SopModelToResponseDto(createdData)
}

func (h *SopHandler) CreateSopGraphHandler(input *dto.CreateSopDto) (*graphdb.SopNode, error) {
	newID := time.Now().UnixMilli()
	now := time.Now().Format(time.RFC3339Nano)

	graphNode := &graphdb.SopNode{
		ID:          newID,
		Name:        input.Name,
		Code:        input.Code,
		Description: fmt.Sprintf("%v", input.Description),
		ParentJobID: input.ParentJobID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.Service.InsertGraphSop(graphNode); err != nil {
		return nil, fmt.Errorf("failed to create graph SOP: %w", err)
	}

	return graphNode, nil
}

// --- UPDATE ---
func (h *SopHandler) UpdateSopHandler(id int64, input *dto.UpdateSopDto) (*models.Sop, error) {
	db := h.Service.GetDB().Begin()
	committed := false
	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
			panic(r)
		}
		if !committed {
			db.Rollback()
		}
	}()

	sopService := h.Service.WithTx(db)

	payload, associations := mapper.UpdateSopDtoToModel(input)
	updatedData, err := sopService.UpdateSop(id, payload, associations)
	if err != nil {
		return nil, err
	}

	graphNode := &graphdb.SopNode{
		ID:          id,
		Name:        *input.Name,
		Code:        *input.Code,
		Description: fmt.Sprintf("%v", *input.Description),
		ParentJobID: input.ParentJobID,
	}
	if err := sopService.UpdateGraphSop(graphNode); err != nil {
		return nil, fmt.Errorf("failed to update graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedData, nil
}

func (h *SopHandler) UpdateSopSqlHandler(id int64, input *dto.UpdateSopDto) (*models.Sop, error) {
	db := h.Service.GetDB().Begin()
	committed := false
	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
			panic(r)
		}
		if !committed {
			db.Rollback()
		}
	}()

	sopService := h.Service.WithTx(db)

	payload, associations := mapper.UpdateSopDtoToModel(input)
	updatedData, err := sopService.UpdateSop(id, payload, associations)
	if err != nil {
		return nil, err
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedData, nil
}

func (h *SopHandler) UpdateSopGraphHandler(id int64, input *dto.UpdateSopDto) (*graphdb.SopNode, error) {
	graphNode := &graphdb.SopNode{
		ID:        id,
		UpdatedAt: time.Now().Format(time.RFC3339Nano),
	}

	if input.Name != nil {
		graphNode.Name = *input.Name
	}
	if input.Code != nil {
		graphNode.Code = *input.Code
	}
	if input.Description != nil {
		graphNode.Description = fmt.Sprintf("%v", *input.Description)
	}
	if input.ParentJobID != nil {
		graphNode.ParentJobID = input.ParentJobID
	}

	if err := h.Service.UpdateGraphSop(graphNode); err != nil {
		return nil, fmt.Errorf("failed to update graph SOP: %w", err)
	}

	return graphNode, nil
}

// --- DELETE ---
func (h *SopHandler) DeleteSopHandler(id int64, isPermanent bool) error {
	db := h.Service.GetDB().Begin()
	committed := false
	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
			panic(r)
		}
		if !committed {
			db.Rollback()
		}
	}()

	sopService := h.Service.WithTx(db)

	if err := sopService.DeleteSop(id, isPermanent); err != nil {
		return err
	}

	if err := sopService.DeleteGraphSop(id); err != nil {
		return fmt.Errorf("failed to delete graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}

func (h *SopHandler) DeleteSopSqlHandler(id int64, isPermanent bool) error {
	db := h.Service.GetDB().Begin()
	committed := false
	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
			panic(r)
		}
		if !committed {
			db.Rollback()
		}
	}()

	sopService := h.Service.WithTx(db)

	if err := sopService.DeleteSop(id, isPermanent); err != nil {
		return err
	}

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}

func (h *SopHandler) DeleteSopGraphHandler(id int64) error {
	if err := h.Service.DeleteGraphSop(id); err != nil {
		return fmt.Errorf("failed to delete graph SOP: %w", err)
	}

	return nil
}

// --- READ ---
func (h *SopHandler) GetSopByIDHandler(id int64, filter dto.SopFilterDto) (*models.Sop, error) {
	return h.Service.GetSopByID(id, filter)
}

func (h *SopHandler) GetSopByIdGraphHandler(id int64) (*graphdb.SopNode, error) {
	return h.Service.GetGraphSopByID(id)
}

func (h *SopHandler) GetAllSopsHandler(filter dto.SopFilterDto) ([]models.Sop, int64, error) {
	start := time.Now()
	data, err := h.Service.GetAllSops(filter)
	if err != nil {
		return nil, 0, err
	}
	helper.RecordDBLatency(time.Since(start))

	var total int64
	db := h.Service.GetDB()
	if filter.Name != "" {
		db = db.Where("name ILIKE ?", "%"+filter.Name+"%")
	}
	if filter.Code != nil {
		db = db.Where("code ILIKE ?", "%"+*filter.Code+"%")
	}
	if filter.DivisionID != 0 {
		db = db.Joins("JOIN sop_divisions ON sop_divisions.sop_id = sops.id").Where("sop_divisions.division_id = ?", filter.DivisionID)
	}
	if len(filter.DivisionIDs) > 0 {
		db = db.Joins("JOIN sop_divisions ON sop_divisions.sop_id = sops.id").Where("sop_divisions.division_id IN ?", filter.DivisionIDs)
	}
	if filter.ShowDeleted {
		db = db.Unscoped().Where("deleted_at IS NOT NULL")
	}
	if err := db.Model(&models.Sop{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (h *SopHandler) GetAllSopsGraphHandler(filter dto.SopFilterDto) ([]*graphdb.SopNode, int64, error) {
	start := time.Now()
	data, err := h.Service.GetAllGraphSops(filter)
	if err != nil {
		return nil, 0, err
	}
	helper.RecordDBLatency(time.Since(start))

	total, err := h.Service.CountGraphSops(filter)
	if err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

// --- BULK CREATE ---
func (h *SopHandler) BulkCreateSopsHandler(input *dto.BulkCreateSopsDto) ([]*models.Sop, error) {
	db := h.Service.GetDB().Begin()
	committed := false
	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
			panic(r)
		}
		if !committed {
			db.Rollback()
		}
	}()

	sopService := h.Service.WithTx(db)
	var sops []*models.Sop

	for _, createDto := range input.Data {
		sop, err := mapper.CreateSopDtoToModel(createDto)
		if err != nil {
			return nil, err
		}
		sops = append(sops, sop)
	}

	createdSops, err := sopService.BulkCreateSops(sops)
	if err != nil {
		return nil, err
	}

	var graphNodes []*graphdb.SopNode
	for _, sqlData := range createdSops {
		graphNodes = append(graphNodes, &graphdb.SopNode{
			ID:          sqlData.ID,
			Name:        sqlData.Name,
			Code:        sqlData.Code,
			Description: fmt.Sprintf("%v", sqlData.Description),
			ParentJobID: sqlData.ParentJobID,
		})
	}
	if err := sopService.BulkInsertGraphSops(graphNodes); err != nil {
		return nil, fmt.Errorf("failed to bulk insert graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return createdSops, nil
}

// --- BULK UPDATE ---
func (h *SopHandler) BulkUpdateSopsHandler(input *dto.BulkUpdateSopDto) ([]*models.Sop, error) {
	db := h.Service.GetDB().Begin()
	committed := false
	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
			panic(r)
		}
		if !committed {
			db.Rollback()
		}
	}()

	sopService := h.Service.WithTx(db)

	updates, associations := mapper.UpdateSopDtoToModel(input.Data)
	if len(updates) == 0 && len(associations) == 0 {
		return nil, fmt.Errorf("update data cannot be empty")
	}

	if err := sopService.BulkUpdateSops(input.IDs, updates); err != nil {
		return nil, fmt.Errorf("failed to bulk update SOPs: %w", err)
	}

	updatedSops, err := sopService.GetSopsByIDs(input.IDs)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated SOPs: %w", err)
	}

	var graphNodes []*graphdb.SopNode
	for _, sqlData := range updatedSops {
		graphNodes = append(graphNodes, &graphdb.SopNode{
			ID:          sqlData.ID,
			Name:        sqlData.Name,
			Code:        sqlData.Code,
			Description: fmt.Sprintf("%v", sqlData.Description),
			ParentJobID: sqlData.ParentJobID,
		})
	}
	if err := sopService.BulkUpdateGraphSops(graphNodes); err != nil {
		return nil, fmt.Errorf("failed to bulk update graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedSops, nil
}

// --- BULK DELETE ---
func (h *SopHandler) BulkDeleteSopsHandler(input *dto.BulkDeleteSopDto, isPermanent bool) error {
	db := h.Service.GetDB().Begin()
	committed := false
	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
			panic(r)
		}
		if !committed {
			db.Rollback()
		}
	}()

	sopService := h.Service.WithTx(db)

	if err := sopService.BulkDeleteSops(input.IDs, isPermanent); err != nil {
		return err
	}

	if err := sopService.BulkDeleteGraphSops(input.IDs); err != nil {
		return fmt.Errorf("failed to bulk delete graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}

func (h *SopHandler) CountSopsHandler(filter dto.SopFilterDto) (int64, error) {
	return h.Service.CountSops(filter)
}
