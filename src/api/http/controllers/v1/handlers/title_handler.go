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

type TitleHandler struct {
	Service service.TitleService
}

func NewTitleHandler(service service.TitleService) *TitleHandler {
	return &TitleHandler{Service: service}
}

// --- CREATE ---
func (h *TitleHandler) CreateTitleHandler(input *dto.CreateTitleDto) (*dto.TitleResponseDto, error) {
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

	titleService := h.Service.WithTx(db)

	payload, err := mapper.CreateTitleDtoToModel(input)
	if err != nil {
		return nil, err
	}

	createdData, err := titleService.CreateTitle(payload)
	if err != nil {
		return nil, err
	}

	graphNode := &graphdb.TitleNode{
		ID:        createdData.ID,
		Name:      createdData.Name,
		Code:      createdData.Code,
		Color:     createdData.Color,
		CreatedAt: createdData.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: createdData.UpdatedAt.Format(time.RFC3339Nano),
	}
	if err := titleService.InsertGraphTitle(graphNode); err != nil {
		return nil, fmt.Errorf("failed to sync to graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return mapper.TitleModelToResponseDto(createdData)
}

func (h *TitleHandler) CreateTitleSqlHandler(input *dto.CreateTitleDto) (*dto.TitleResponseDto, error) {
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

	titleService := h.Service.WithTx(db)

	payload, err := mapper.CreateTitleDtoToModel(input)
	if err != nil {
		return nil, err
	}

	createdData, err := titleService.CreateTitle(payload)
	if err != nil {
		return nil, err
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return mapper.TitleModelToResponseDto(createdData)
}

func (h *TitleHandler) CreateTitleGraphHandler(input *dto.CreateTitleDto) (*graphdb.TitleNode, error) {
	newID := time.Now().UnixMilli()
	now := time.Now().Format(time.RFC3339Nano)

	graphNode := &graphdb.TitleNode{
		ID:        newID,
		Name:      input.Name,
		Code:      input.Code,
		Color:     input.Color,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.Service.InsertGraphTitle(graphNode); err != nil {
		return nil, fmt.Errorf("failed to create graph title: %w", err)
	}

	return graphNode, nil
}

// --- UPDATE ---
func (h *TitleHandler) UpdateTitleHandler(id int64, input *dto.UpdateTitleDto) (*models.Title, error) {
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

	titleService := h.Service.WithTx(db)

	payload, err := mapper.UpdateTitleDtoToModel(input)
	if err != nil {
		return nil, err
	}

	updatedData, err := titleService.UpdateTitle(id, payload)
	if err != nil {
		return nil, err
	}

	graphNode := &graphdb.TitleNode{
		ID:    id,
		Name:  *input.Name,
		Code:  *input.Code,
		Color: *input.Color,
	}
	if err := titleService.UpdateGraphTitle(graphNode); err != nil {
		return nil, fmt.Errorf("failed to update graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedData, nil
}

func (h *TitleHandler) UpdateTitleSqlHandler(id int64, input *dto.UpdateTitleDto) (*models.Title, error) {
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

	titleService := h.Service.WithTx(db)

	payload, err := mapper.UpdateTitleDtoToModel(input)
	if err != nil {
		return nil, err
	}

	updatedData, err := titleService.UpdateTitle(id, payload)
	if err != nil {
		return nil, err
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedData, nil
}

func (h *TitleHandler) UpdateTitleGraphHandler(id int64, input *dto.UpdateTitleDto) (*graphdb.TitleNode, error) {
	graphNode := &graphdb.TitleNode{
		ID:        id,
		UpdatedAt: time.Now().Format(time.RFC3339Nano),
	}

	if input.Name != nil {
		graphNode.Name = *input.Name
	}
	if input.Code != nil {
		graphNode.Code = *input.Code
	}
	if input.Color != nil {
		graphNode.Color = *input.Color
	}

	if err := h.Service.UpdateGraphTitle(graphNode); err != nil {
		return nil, fmt.Errorf("failed to update graph title: %w", err)
	}

	return graphNode, nil
}

// --- DELETE ---
func (h *TitleHandler) DeleteTitleHandler(id int64) error {
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

	titleService := h.Service.WithTx(db)

	if err := titleService.DeleteTitle(id); err != nil {
		return err
	}

	if err := titleService.DeleteGraphTitle(id); err != nil {
		return fmt.Errorf("failed to delete graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}

func (h *TitleHandler) DeleteTitleSqlHandler(id int64) error {
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

	titleService := h.Service.WithTx(db)

	if err := titleService.DeleteTitle(id); err != nil {
		return err
	}

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}

func (h *TitleHandler) DeleteTitleGraphHandler(id int64) error {
	if err := h.Service.DeleteGraphTitle(id); err != nil {
		return fmt.Errorf("failed to delete graph title: %w", err)
	}

	return nil
}

// --- READ ---
func (h *TitleHandler) GetTitleByIDHandler(id int64, filter dto.TitleFilterDto) (*models.Title, error) {
	return h.Service.GetTitleByID(id, filter)
}

func (h *TitleHandler) GetTitleByIdGraphHandler(id int64) (*graphdb.TitleNode, error) {
	return h.Service.GetGraphTitleByID(id)
}

func (h *TitleHandler) GetAllTitlesHandler(filter dto.TitleFilterDto) ([]models.Title, int64, error) {
	data, err := h.Service.GetAllTitles(filter)
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
	if err := db.Model(&models.Title{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (h *TitleHandler) GetAllTitlesGraphHandler(filter dto.TitleFilterDto) ([]*graphdb.TitleNode, int64, error) {
	data, err := h.Service.GetAllGraphTitles(filter)
	if err != nil {
		return nil, 0, err
	}

	total, err := h.Service.CountGraphTitles(filter)
	if err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

// --- BULK CREATE ---
func (h *TitleHandler) BulkCreateHandler(input *dto.BulkCreateTitleDto) ([]*models.Title, error) {
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

	titleService := h.Service.WithTx(db)
	var titles []*models.Title

	for _, createDto := range input.Data {
		title, err := mapper.CreateTitleDtoToModel(createDto)
		if err != nil {
			return nil, err
		}
		if title != nil {
			titles = append(titles, title)
		}
	}

	createdTitles, err := titleService.BulkCreateTitles(titles)
	if err != nil {
		return nil, err
	}

	var graphNodes []*graphdb.TitleNode
	for _, sqlData := range createdTitles {
		graphNodes = append(graphNodes, &graphdb.TitleNode{
			ID:    sqlData.ID,
			Name:  sqlData.Name,
			Code:  sqlData.Code,
			Color: sqlData.Color,
		})
	}
	if err := titleService.BulkInsertGraphTitles(graphNodes); err != nil {
		return nil, fmt.Errorf("failed to bulk insert graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return createdTitles, nil
}

// --- BULK UPDATE ---
func (h *TitleHandler) BulkUpdateHandler(input *dto.BulkUpdateTitleDto) ([]*models.Title, error) {
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

	titleService := h.Service.WithTx(db)

	updates, err := mapper.UpdateTitleDtoToModel(input.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to map update data: %w", err)
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("update data cannot be empty")
	}

	err = titleService.BulkUpdateTitles(input.IDs, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk update titles: %w", err)
	}

	updatedTitles, err := titleService.GetTitlesByIDs(input.IDs)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated titles: %w", err)
	}

	var graphNodes []*graphdb.TitleNode
	for _, sqlData := range updatedTitles {
		graphNodes = append(graphNodes, &graphdb.TitleNode{
			ID:    sqlData.ID,
			Name:  sqlData.Name,
			Code:  sqlData.Code,
			Color: sqlData.Color,
		})
	}
	if err := titleService.BulkUpdateGraphTitles(graphNodes); err != nil {
		return nil, fmt.Errorf("failed to bulk update graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedTitles, nil
}

// --- BULK DELETE ---
func (h *TitleHandler) BulkDeleteHandler(input *dto.BulkDeleteTitleDto) error {
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

	titleService := h.Service.WithTx(db)

	if err := titleService.BulkDeleteTitles(input.IDs); err != nil {
		return err
	}

	if err := titleService.BulkDeleteGraphTitles(input.IDs); err != nil {
		return fmt.Errorf("failed to bulk delete graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}
