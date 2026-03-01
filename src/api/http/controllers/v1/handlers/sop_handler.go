package handlers

import (
	"fmt"

	"jk-api/api/http/controllers/v1/dto"
	"jk-api/api/http/controllers/v1/mapper"
	"jk-api/internal/database/models"
	"jk-api/internal/service"
)

type SopHandler struct {
	Service service.SopService
}

func NewSopHandler(service service.SopService) *SopHandler {
	return &SopHandler{Service: service}
}

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

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return mapper.SopModelToResponseDto(createdData)
}

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

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedData, nil
}

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

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}

func (h *SopHandler) GetSopByIDHandler(id int64, filter dto.SopFilterDto) (*models.Sop, error) {
	return h.Service.GetSopByID(id, filter)
}

func (h *SopHandler) GetAllSopsHandler(filter dto.SopFilterDto) ([]models.Sop, int64, error) {
	data, err := h.Service.GetAllSops(filter)
	if err != nil {
		return nil, 0, err
	}

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

func (h *SopHandler) BulkCreateSopsHandler(input *dto.BulkCreateSopsDto) ([]*models.Sop, error) {
	var sops []*models.Sop
	for _, createDto := range input.Data {
		sop, err := mapper.CreateSopDtoToModel(createDto)
		if err != nil {
			return nil, err
		}
		sops = append(sops, sop)
	}
	return h.Service.BulkCreateSops(sops)
}

func (h *SopHandler) BulkUpdateSopsHandler(input *dto.BulkUpdateSopDto) ([]*models.Sop, error) {
	updates, associations := mapper.UpdateSopDtoToModel(input.Data)
	if len(updates) == 0 && len(associations) == 0 {
		return nil, fmt.Errorf("update data cannot be empty")
	}

	if err := h.Service.BulkUpdateSops(input.IDs, updates); err != nil {
		return nil, fmt.Errorf("failed to bulk update sops: %w", err)
	}

	updatedSops, err := h.Service.GetSopsByIDs(input.IDs)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated sops: %w", err)
	}
	return updatedSops, nil
}

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

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}

func (h *SopHandler) CountSopsHandler(filter dto.SopFilterDto) (int64, error) {
	return h.Service.CountSops(filter)
}
