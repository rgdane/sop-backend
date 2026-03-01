package handlers

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/api/http/controllers/v1/mapper"
	"jk-api/internal/database/models"
	"jk-api/internal/service"
)

type TitleHandler struct {
	Service service.TitleService
}

func NewTitleHandler(service service.TitleService) *TitleHandler {
	return &TitleHandler{Service: service}
}

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

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return mapper.TitleModelToResponseDto(createdData)
}

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

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedData, nil
}

func (h *TitleHandler) DeleteTitleHandler(id int64) error {
	return h.Service.DeleteTitle(id)
}

func (h *TitleHandler) GetTitleByIDHandler(id int64, filter dto.TitleFilterDto) (*models.Title, error) {
	return h.Service.GetTitleByID(id, filter)
}

func (h *TitleHandler) GetAllTitlesHandler(filter dto.TitleFilterDto) ([]models.Title, int64, error) {
	data, err := h.Service.GetAllTitles(filter)
	if err != nil {
		return nil, 0, err
	}
	var total int64
	db := h.Service.GetDB()
	if filter.Name != "" {
		db = db.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if err := db.Model(&models.Title{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	return data, total, nil
}

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

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return createdTitles, nil
}

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

	if err := titleService.BulkUpdateTitles(input.IDs, updates); err != nil {
		return nil, fmt.Errorf("failed to bulk update titles: %w", err)
	}

	updatedTitles, err := titleService.GetTitlesByIDs(input.IDs)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated titles: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedTitles, nil
}

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

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}
