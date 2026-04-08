package handlers

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/api/http/controllers/v1/mapper"
	"jk-api/internal/database/models"
	"jk-api/internal/repository/graphdb"
	"jk-api/internal/service"
	"strings"
	"time"
)

type SpkJobHandler struct {
	Service service.SpkJobService
}

func NewSpkJobHandler(service service.SpkJobService) *SpkJobHandler {
	return &SpkJobHandler{Service: service}
}

func (h *SpkJobHandler) CreateSpkJobHandler(input *dto.CreateSpkJobDto) (*dto.SpkJobResponseDto, error) {
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

	spkJobService := h.Service.WithTx(db)

	if input.SopID != nil && input.Url != nil {
		url := fmt.Sprintf("%s/dashboard/master/sop-view/%d", *input.Url, *input.SopID)
		linkHTML := fmt.Sprintf(`<p><strong>Link SOP : </strong><a href="%s" target="_blank" rel="noopener noreferrer nofollow">%s</a></p>`, url, url)

		if input.Description != nil && len(*input.Description) > 0 {
			newDesc := linkHTML + *input.Description
			input.Description = &newDesc
		} else {
			input.Description = &linkHTML
		}
	}

	payload, err := mapper.CreateSpkJobDtoToModel(input)
	if err != nil {
		return nil, err
	}

	createdData, err := spkJobService.CreateSpkJob(payload)
	if err != nil {
		return nil, err
	}

	graphNode := &graphdb.SpkJobNode{
		ID:          createdData.ID,
		Name:        createdData.Name,
		Description: fmt.Sprintf("%v", createdData.Description),
		SpkID:       createdData.SpkID,
		SopID:       createdData.SopID,
		TitleID:     createdData.TitleID,
		Index:       createdData.Index,
		FlowchartID: createdData.FlowchartID,
		NextIndex:   createdData.NextIndex,
		PrevIndex:   createdData.PrevIndex,
		CreatedAt:   createdData.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   createdData.UpdatedAt.Format(time.RFC3339Nano),
	}
	if err := spkJobService.InsertGraphSpkJob(graphNode); err != nil {
		return nil, fmt.Errorf("failed to sync to graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return mapper.SpkJobModelToResponseDto(createdData)
}

func (h *SpkJobHandler) CreateSpkJobSqlHandler(input *dto.CreateSpkJobDto) (*dto.SpkJobResponseDto, error) {
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

	spkJobService := h.Service.WithTx(db)

	if input.SopID != nil && input.Url != nil {
		url := fmt.Sprintf("%s/dashboard/master/sop-view/%d", *input.Url, *input.SopID)
		linkHTML := fmt.Sprintf(`<p><strong>Link SOP : </strong><a href="%s" target="_blank" rel="noopener noreferrer nofollow">%s</a></p>`, url, url)

		if input.Description != nil && len(*input.Description) > 0 {
			newDesc := linkHTML + *input.Description
			input.Description = &newDesc
		} else {
			input.Description = &linkHTML
		}
	}

	payload, err := mapper.CreateSpkJobDtoToModel(input)
	if err != nil {
		return nil, err
	}

	createdData, err := spkJobService.CreateSpkJob(payload)
	if err != nil {
		return nil, err
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return mapper.SpkJobModelToResponseDto(createdData)
}

func (h *SpkJobHandler) CreateSpkJobGraphHandler(input *dto.CreateSpkJobDto) (*graphdb.SpkJobNode, error) {
	newID := time.Now().UnixMilli()
	now := time.Now().Format(time.RFC3339Nano)

	graphNode := &graphdb.SpkJobNode{
		ID:          newID,
		Name:        input.Name,
		Description: fmt.Sprintf("%v", input.Description),
		SpkID:       input.SpkID,
		SopID:       input.SopID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.Service.InsertGraphSpkJob(graphNode); err != nil {
		return nil, fmt.Errorf("failed to create graph SPK Job: %w", err)
	}

	return graphNode, nil
}

func (h *SpkJobHandler) UpdateSpkJobHandler(id int64, input *dto.UpdateSpkJobDto) (*models.SpkJob, error) {
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

	spkJobService := h.Service.WithTx(db)

	if input.SopID != nil && input.Url != nil {
		url := fmt.Sprintf("%s/dashboard/master/sop-view/%d", *input.Url, *input.SopID)
		linkHTML := fmt.Sprintf(`<p><strong>Link SOP : </strong><a href="%s" target="_blank" rel="noopener noreferrer nofollow">%s</a></p>`, url, url)

		if input.Description != nil && len(*input.Description) > 0 {
			desc := *input.Description
			if strings.HasPrefix(desc, "<p><strong>Link SOP : </strong><a href=") {
				endIdx := strings.Index(desc, "</p>")
				if endIdx != -1 {
					newDesc := linkHTML + desc[endIdx+4:]
					input.Description = &newDesc
				}
			} else {
				newDesc := linkHTML + desc
				input.Description = &newDesc
			}
		} else {
			input.Description = &linkHTML
		}
	}

	payload, err := mapper.UpdateSpkJobDtoToModel(input)
	if err != nil {
		return nil, err
	}

	updatedData, err := spkJobService.UpdateSpkJob(id, payload)
	if err != nil {
		return nil, err
	}

	graphNode := &graphdb.SpkJobNode{
		ID:          id,
		Description: fmt.Sprintf("%v", updatedData.Description),
		UpdatedAt:   time.Now().Format(time.RFC3339Nano),
	}

	if input.Name != nil {
		graphNode.Name = *input.Name
	}
	if input.Description != nil {
		graphNode.Description = fmt.Sprintf("%v", *input.Description)
	}

	if err := spkJobService.UpdateGraphSpkJob(graphNode); err != nil {
		return nil, fmt.Errorf("failed to update graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedData, nil
}

func (h *SpkJobHandler) UpdateSpkJobSqlHandler(id int64, input *dto.UpdateSpkJobDto) (*models.SpkJob, error) {
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

	spkJobService := h.Service.WithTx(db)

	if input.SopID != nil && input.Url != nil {
		url := fmt.Sprintf("%s/dashboard/master/sop-view/%d", *input.Url, *input.SopID)
		linkHTML := fmt.Sprintf(`<p><strong>Link SOP : </strong><a href="%s" target="_blank" rel="noopener noreferrer nofollow">%s</a></p>`, url, url)

		if input.Description != nil && len(*input.Description) > 0 {
			desc := *input.Description
			if strings.HasPrefix(desc, "<p><strong>Link SOP : </strong><a href=") {
				endIdx := strings.Index(desc, "</p>")
				if endIdx != -1 {
					newDesc := linkHTML + desc[endIdx+4:]
					input.Description = &newDesc
				}
			} else {
				newDesc := linkHTML + desc
				input.Description = &newDesc
			}
		} else {
			input.Description = &linkHTML
		}
	}

	payload, err := mapper.UpdateSpkJobDtoToModel(input)
	if err != nil {
		return nil, err
	}

	updatedData, err := spkJobService.UpdateSpkJob(id, payload)
	if err != nil {
		return nil, err
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedData, nil
}

func (h *SpkJobHandler) UpdateSpkJobGraphHandler(id int64, input *dto.UpdateSpkJobDto) (*graphdb.SpkJobNode, error) {
	graphNode := &graphdb.SpkJobNode{
		ID:        id,
		UpdatedAt: time.Now().Format(time.RFC3339Nano),
	}

	if input.Name != nil {
		graphNode.Name = *input.Name
	}
	if input.Description != nil {
		graphNode.Description = fmt.Sprintf("%v", *input.Description)
	}

	if err := h.Service.UpdateGraphSpkJob(graphNode); err != nil {
		return nil, fmt.Errorf("failed to update graph SPK Job: %w", err)
	}

	return graphNode, nil
}

func (h *SpkJobHandler) DeleteSpkJobHandler(id int64, isPermanent bool) error {
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

	spkJobService := h.Service.WithTx(db)

	if err := spkJobService.DeleteSpkJob(id, isPermanent); err != nil {
		return err
	}

	if err := spkJobService.DeleteGraphSpkJob(id); err != nil {
		return fmt.Errorf("failed to delete graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}

func (h *SpkJobHandler) DeleteSpkJobSqlHandler(id int64, isPermanent bool) error {
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

	spkJobService := h.Service.WithTx(db)

	if err := spkJobService.DeleteSpkJob(id, isPermanent); err != nil {
		return err
	}

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}

func (h *SpkJobHandler) DeleteSpkJobGraphHandler(id int64) error {
	if err := h.Service.DeleteGraphSpkJob(id); err != nil {
		return fmt.Errorf("failed to delete graph SPK Job: %w", err)
	}

	return nil
}

func (h *SpkJobHandler) GetSpkJobByIDHandler(id int64, filter dto.SpkJobFilterDto) (*models.SpkJob, error) {
	return h.Service.GetSpkJobByID(id, filter)
}

func (h *SpkJobHandler) GetSpkJobByIdGraphHandler(id int64) (*graphdb.SpkJobNode, error) {
	return h.Service.GetGraphSpkJobByID(id)
}

func (h *SpkJobHandler) GetAllSpkJobsHandler(filter dto.SpkJobFilterDto) ([]models.SpkJob, int64, error) {
	data, err := h.Service.GetAllSpkJobs(filter)
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
	if err := db.Model(&models.SpkJob{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (h *SpkJobHandler) GetAllSpkJobsGraphHandler(filter dto.SpkJobFilterDto) ([]*graphdb.SpkJobNode, int64, error) {
	data, err := h.Service.GetAllGraphSpkJobs(filter)
	if err != nil {
		return nil, 0, err
	}

	total, err := h.Service.CountGraphSpkJobs(filter)
	if err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (h *SpkJobHandler) BulkCreateSpkJobsHandler(input *dto.BulkCreateSpkJobsDto) ([]*models.SpkJob, error) {
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

	spkJobService := h.Service.WithTx(db)
	var spkJobs []*models.SpkJob

	for _, createDto := range input.Data {
		spkJob, err := mapper.CreateSpkJobDtoToModel(createDto)
		if err != nil {
			return nil, err
		}
		if spkJob != nil {
			spkJobs = append(spkJobs, spkJob)
		}
	}

	createdSpkJobs, err := spkJobService.BulkCreateSpkJobs(spkJobs)
	if err != nil {
		return nil, err
	}

	var graphNodes []*graphdb.SpkJobNode
	for _, sqlData := range createdSpkJobs {
		graphNodes = append(graphNodes, &graphdb.SpkJobNode{
			ID:   sqlData.ID,
			Name: sqlData.Name,
		})
	}
	if err := spkJobService.BulkInsertGraphSpkJobs(graphNodes); err != nil {
		return nil, fmt.Errorf("failed to bulk insert graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return createdSpkJobs, nil
}

func (h *SpkJobHandler) BulkUpdateSpkJobsHandler(input *dto.BulkUpdateSpkJobDto) ([]*models.SpkJob, error) {
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

	spkJobService := h.Service.WithTx(db)

	updates, err := mapper.UpdateSpkJobDtoToModel(input.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to map update data: %w", err)
	}

	if err := spkJobService.BulkUpdateSpkJobs(input.IDs, updates); err != nil {
		return nil, fmt.Errorf("failed to bulk update SpkJobs: %w", err)
	}

	updatedSpkJobs, err := spkJobService.GetSpkJobsByIDs(input.IDs)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated SpkJobs: %w", err)
	}

	var graphNodes []*graphdb.SpkJobNode
	for _, sqlData := range updatedSpkJobs {
		graphNodes = append(graphNodes, &graphdb.SpkJobNode{
			ID:   sqlData.ID,
			Name: sqlData.Name,
		})
	}
	if err := spkJobService.BulkUpdateGraphSpkJobs(graphNodes); err != nil {
		return nil, fmt.Errorf("failed to bulk update graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedSpkJobs, nil
}

func (h *SpkJobHandler) BulkDeleteSpkJobsHandler(input *dto.BulkDeleteSpkJobDto, isPermanent bool) error {
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

	spkJobService := h.Service.WithTx(db)

	if err := spkJobService.BulkDeleteSpkJobs(input.IDs, isPermanent); err != nil {
		return err
	}

	if err := spkJobService.BulkDeleteGraphSpkJobs(input.IDs); err != nil {
		return fmt.Errorf("failed to bulk delete graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}

func (h *SpkJobHandler) ReorderSpkJobHandler(id int64, input *dto.ReorderSpkJobDto) error {
	return h.Service.ReorderSpkJob(id, input.NewIndex, input.SpkID)
}

func (h *SpkJobHandler) CountSpkJobsHandler(filter dto.SpkJobFilterDto) (int64, error) {
	return h.Service.CountSpkJobs(filter)
}
