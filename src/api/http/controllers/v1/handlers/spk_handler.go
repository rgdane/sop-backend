package handlers

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/api/http/controllers/v1/mapper"
	"jk-api/internal/database/models"
	"jk-api/internal/repository/graphdb"
	"jk-api/internal/service"
	"time"
)

type SpkHandler struct {
	Service service.SpkService
}

func NewSpkHandler(service service.SpkService) *SpkHandler {
	return &SpkHandler{Service: service}
}

func (h *SpkHandler) CreateSpkHandler(input *dto.CreateSpkDto) (*dto.SpkResponseDto, error) {
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

	spkService := h.Service.WithTx(db)

	payload, err := mapper.CreateSpkDtoToModel(input)
	if err != nil {
		return nil, err
	}

	createdData, err := spkService.CreateSpk(payload)
	if err != nil {
		return nil, err
	}

	graphNode := &graphdb.SpkNode{
		ID:          createdData.ID,
		Name:        createdData.Name,
		Code:        createdData.Code,
		Description: fmt.Sprintf("%v", createdData.Description),
		CreatedAt:   createdData.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   createdData.UpdatedAt.Format(time.RFC3339Nano),
	}
	if err := spkService.InsertGraphSpk(graphNode); err != nil {
		return nil, fmt.Errorf("failed to sync to graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return mapper.SpkModelToResponseDto(createdData)
}

func (h *SpkHandler) CreateSpkSqlHandler(input *dto.CreateSpkDto) (*dto.SpkResponseDto, error) {
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

	spkService := h.Service.WithTx(db)

	payload, err := mapper.CreateSpkDtoToModel(input)
	if err != nil {
		return nil, err
	}

	createdData, err := spkService.CreateSpk(payload)
	if err != nil {
		return nil, err
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return mapper.SpkModelToResponseDto(createdData)
}

func (h *SpkHandler) CreateSpkGraphHandler(input *dto.CreateSpkDto) (*graphdb.SpkNode, error) {
	newID := time.Now().UnixMilli()
	now := time.Now().Format(time.RFC3339Nano)

	graphNode := &graphdb.SpkNode{
		ID:          newID,
		Name:        input.Name,
		Code:        input.Code,
		Description: fmt.Sprintf("%v", input.Description),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.Service.InsertGraphSpk(graphNode); err != nil {
		return nil, fmt.Errorf("failed to create graph SPK: %w", err)
	}

	return graphNode, nil
}

func (h *SpkHandler) UpdateSpkHandler(id int64, input *dto.UpdateSpkDto) (*models.Spk, error) {
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

	spkService := h.Service.WithTx(db)

	payload, associations := mapper.UpdateSpkDtoToModel(input)
	updatedData, err := spkService.UpdateSpk(id, payload, associations)
	if err != nil {
		return nil, err
	}

	graphNode := &graphdb.SpkNode{
		ID:          id,
		Name:        *input.Name,
		Code:        *input.Code,
		Description: fmt.Sprintf("%v", *input.Description),
	}
	if err := spkService.UpdateGraphSpk(graphNode); err != nil {
		return nil, fmt.Errorf("failed to update graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedData, nil
}

func (h *SpkHandler) UpdateSpkSqlHandler(id int64, input *dto.UpdateSpkDto) (*models.Spk, error) {
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

	spkService := h.Service.WithTx(db)

	payload, associations := mapper.UpdateSpkDtoToModel(input)
	updatedData, err := spkService.UpdateSpk(id, payload, associations)
	if err != nil {
		return nil, err
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedData, nil
}

func (h *SpkHandler) UpdateSpkGraphHandler(id int64, input *dto.UpdateSpkDto) (*graphdb.SpkNode, error) {
	graphNode := &graphdb.SpkNode{
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

	if err := h.Service.UpdateGraphSpk(graphNode); err != nil {
		return nil, fmt.Errorf("failed to update graph SPK: %w", err)
	}

	return graphNode, nil
}

func (h *SpkHandler) DeleteSpkHandler(id int64, isPermanent bool) error {
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

	spkService := h.Service.WithTx(db)

	if err := spkService.DeleteSpk(id, isPermanent); err != nil {
		return err
	}

	if err := spkService.DeleteGraphSpk(id); err != nil {
		return fmt.Errorf("failed to delete graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}

func (h *SpkHandler) DeleteSpkSqlHandler(id int64, isPermanent bool) error {
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

	spkService := h.Service.WithTx(db)

	if err := spkService.DeleteSpk(id, isPermanent); err != nil {
		return err
	}

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}

func (h *SpkHandler) DeleteSpkGraphHandler(id int64) error {
	if err := h.Service.DeleteGraphSpk(id); err != nil {
		return fmt.Errorf("failed to delete graph SPK: %w", err)
	}

	return nil
}

func (h *SpkHandler) GetSpkByIDHandler(id int64, filter dto.SpkFilterDto) (*models.Spk, error) {
	return h.Service.GetSpkByID(id, filter)
}

func (h *SpkHandler) GetSpkByIdGraphHandler(id int64) (*graphdb.SpkNode, error) {
	return h.Service.GetGraphSpkByID(id)
}

func (h *SpkHandler) GetAllSpksHandler(filter dto.SpkFilterDto) ([]models.Spk, int64, error) {
	data, err := h.Service.GetAllSpks(filter)
	if err != nil {
		return nil, 0, err
	}

	var total int64
	db := h.Service.GetDB()
	if filter.Name != "" {
		db = db.Where("name ILIKE ?", "%"+filter.Name+"%")
	}
	if filter.ShowDeleted {
		db = db.Unscoped().Where("deleted_at IS NOT NULL")
	}
	if err := db.Model(&models.Spk{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (h *SpkHandler) GetAllSpksGraphHandler(filter dto.SpkFilterDto) ([]*graphdb.SpkNode, int64, error) {
	data, err := h.Service.GetAllGraphSpks(filter)
	if err != nil {
		return nil, 0, err
	}

	total, err := h.Service.CountGraphSpks(filter)
	if err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (h *SpkHandler) BulkCreateSpksHandler(input *dto.BulkCreateSpksDto) ([]*models.Spk, error) {
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

	spkService := h.Service.WithTx(db)
	var spks []*models.Spk

	for _, createDto := range input.Data {
		spk, err := mapper.CreateSpkDtoToModel(createDto)
		if err != nil {
			return nil, err
		}
		spks = append(spks, spk)
	}

	createdSpks, err := spkService.BulkCreateSpks(spks)
	if err != nil {
		return nil, err
	}

	var graphNodes []*graphdb.SpkNode
	for _, sqlData := range createdSpks {
		graphNodes = append(graphNodes, &graphdb.SpkNode{
			ID:          sqlData.ID,
			Name:        sqlData.Name,
			Code:        sqlData.Code,
			Description: fmt.Sprintf("%v", sqlData.Description),
		})
	}
	if err := spkService.BulkInsertGraphSpks(graphNodes); err != nil {
		return nil, fmt.Errorf("failed to bulk insert graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return createdSpks, nil
}

func (h *SpkHandler) BulkUpdateSpksHandler(input *dto.BulkUpdateSpkDto) ([]*models.Spk, error) {
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

	spkService := h.Service.WithTx(db)

	updates, associations := mapper.UpdateSpkDtoToModel(input.Data)
	if len(updates) == 0 && len(associations) == 0 {
		return nil, fmt.Errorf("update data cannot be empty")
	}

	if err := spkService.BulkUpdateSpks(input.IDs, updates); err != nil {
		return nil, fmt.Errorf("failed to bulk update Spks: %w", err)
	}

	updatedSpks, err := spkService.GetSpksByIDs(input.IDs)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated Spks: %w", err)
	}

	var graphNodes []*graphdb.SpkNode
	for _, sqlData := range updatedSpks {
		graphNodes = append(graphNodes, &graphdb.SpkNode{
			ID:          sqlData.ID,
			Name:        sqlData.Name,
			Code:        sqlData.Code,
			Description: fmt.Sprintf("%v", sqlData.Description),
		})
	}
	if err := spkService.BulkUpdateGraphSpks(graphNodes); err != nil {
		return nil, fmt.Errorf("failed to bulk update graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedSpks, nil
}

func (h *SpkHandler) BulkDeleteSpksHandler(input *dto.BulkDeleteSpkDto, isPermanent bool) error {
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

	spkService := h.Service.WithTx(db)

	if err := spkService.BulkDeleteSpks(input.IDs, isPermanent); err != nil {
		return err
	}

	if err := spkService.BulkDeleteGraphSpks(input.IDs); err != nil {
		return fmt.Errorf("failed to bulk delete graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}

func (h *SpkHandler) CountSpksHandler(filter dto.SpkFilterDto) (int64, error) {
	return h.Service.CountSpks(filter)
}
