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

type SopJobHandler struct {
	Service service.SopJobService
}

func NewSopJobHandler(service service.SopJobService) *SopJobHandler {
	return &SopJobHandler{Service: service}
}

func (h *SopJobHandler) CreateSopJobHandler(input *dto.CreateSopJobDto) (*dto.SopJobResponseDto, error) {
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

	sopJobService := h.Service.WithTx(db)

	if input.Type != nil && input.ReferenceID != nil && input.Url != nil {
		var viewType string
		switch *input.Type {
		case "spk":
			viewType = "spk-view"
		case "sop":
			viewType = "sop-view"
		default:
			viewType = "sop-view"
		}

		url := fmt.Sprintf("%s/dashboard/master/%s/%d", *input.Url, viewType, *input.ReferenceID)
		linkHTML := fmt.Sprintf(`<p><strong>Link SOP/SPK : </strong><a href="%s" target="_blank" rel="noopener noreferrer nofollow">%s</a></p>`, url, url)

		if input.Description != nil && len(*input.Description) > 0 {
			newDesc := linkHTML + *input.Description
			input.Description = &newDesc
		} else {
			input.Description = &linkHTML
		}
	}

	payload, err := mapper.CreateSopJobDtoToModel(input)
	if err != nil {
		return nil, err
	}

	createdData, err := sopJobService.CreateSopJob(payload)
	if err != nil {
		return nil, err
	}

	graphNode := &graphdb.SopJobNode{
		ID:          createdData.ID,
		Name:        createdData.Name,
		Alias:       createdData.Alias,
		Type:        "",
		Code:        createdData.Code,
		Description: fmt.Sprintf("%v", createdData.Description),
		TitleID:     createdData.TitleID,
		SopID:       createdData.SopID,
		ReferenceID: createdData.ReferenceID,
		Index:       createdData.Index,
		FlowchartID: createdData.FlowchartID,
		NextIndex:   createdData.NextIndex,
		PrevIndex:   createdData.PrevIndex,
		IsPublished: createdData.IsPublished,
		IsHide:      createdData.IsHide,
		CreatedAt:   createdData.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   createdData.UpdatedAt.Format(time.RFC3339Nano),
	}
	if createdData.Type != nil {
		graphNode.Type = *createdData.Type
	}
	if err := sopJobService.InsertGraphSopJob(graphNode); err != nil {
		return nil, fmt.Errorf("failed to sync to graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return mapper.SopJobModelToResponseDto(createdData)
}

func (h *SopJobHandler) CreateSopJobSqlHandler(input *dto.CreateSopJobDto) (*dto.SopJobResponseDto, error) {
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

	sopJobService := h.Service.WithTx(db)

	if input.Type != nil && input.ReferenceID != nil && input.Url != nil {
		var viewType string
		switch *input.Type {
		case "spk":
			viewType = "spk-view"
		case "sop":
			viewType = "sop-view"
		default:
			viewType = "sop-view"
		}

		url := fmt.Sprintf("%s/dashboard/master/%s/%d", *input.Url, viewType, *input.ReferenceID)
		linkHTML := fmt.Sprintf(`<p><strong>Link SOP/SPK : </strong><a href="%s" target="_blank" rel="noopener noreferrer nofollow">%s</a></p>`, url, url)

		if input.Description != nil && len(*input.Description) > 0 {
			newDesc := linkHTML + *input.Description
			input.Description = &newDesc
		} else {
			input.Description = &linkHTML
		}
	}

	payload, err := mapper.CreateSopJobDtoToModel(input)
	if err != nil {
		return nil, err
	}

	createdData, err := sopJobService.CreateSopJob(payload)
	if err != nil {
		return nil, err
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return mapper.SopJobModelToResponseDto(createdData)
}

func (h *SopJobHandler) CreateSopJobGraphHandler(input *dto.CreateSopJobDto) (*graphdb.SopJobNode, error) {
	newID := time.Now().UnixMilli()
	now := time.Now().Format(time.RFC3339Nano)

	graphNode := &graphdb.SopJobNode{
		ID:          newID,
		Name:        input.Name,
		Alias:       input.Alias,
		Type:        "",
		Description: fmt.Sprintf("%v", input.Description),
		SopID:       input.SopID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if input.Type != nil {
		graphNode.Type = *input.Type
	}

	if err := h.Service.InsertGraphSopJob(graphNode); err != nil {
		return nil, fmt.Errorf("failed to create graph SOP Job: %w", err)
	}

	return graphNode, nil
}

func (h *SopJobHandler) UpdateSopJobHandler(id int64, input *dto.UpdateSopJobDto) (*models.SopJob, error) {
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

	sopJobService := h.Service.WithTx(db)

	if input.Type != nil && input.ReferenceID != nil && input.Url != nil {
		var viewType string
		switch *input.Type {
		case "spk":
			viewType = "spk-view"
		case "sop":
			viewType = "sop-view"
		default:
			viewType = "sop-view"
		}

		url := fmt.Sprintf("%s/dashboard/master/%s/%d", *input.Url, viewType, *input.ReferenceID)
		linkHTML := fmt.Sprintf(`<p><strong>Link SOP/SPK : </strong><a href="%s" target="_blank" rel="noopener noreferrer nofollow">%s</a></p>`, url, url)

		if input.Description != nil && len(*input.Description) > 0 {
			desc := *input.Description
			if strings.HasPrefix(desc, "<p><strong>Link SOP/SPK : </strong><a href=") {
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

	payload, err := mapper.UpdateSopJobDtoToModel(input)
	if err != nil {
		return nil, err
	}

	updatedData, err := sopJobService.UpdateSopJob(id, payload)
	if err != nil {
		return nil, err
	}

	graphNode := &graphdb.SopJobNode{
		ID:        id,
		UpdatedAt: time.Now().Format(time.RFC3339Nano),
	}

	if input.Name != nil {
		graphNode.Name = *input.Name
	}
	if input.Alias != nil {
		graphNode.Alias = *input.Alias
	}
	if input.Type != nil {
		graphNode.Type = *input.Type
	}
	if input.Description != nil {
		graphNode.Description = fmt.Sprintf("%v", *input.Description)
	}

	if err := sopJobService.UpdateGraphSopJob(graphNode); err != nil {
		return nil, fmt.Errorf("failed to update graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedData, nil
}

func (h *SopJobHandler) UpdateSopJobSqlHandler(id int64, input *dto.UpdateSopJobDto) (*models.SopJob, error) {
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

	sopJobService := h.Service.WithTx(db)

	if input.Type != nil && input.ReferenceID != nil && input.Url != nil {
		var viewType string
		switch *input.Type {
		case "spk":
			viewType = "spk-view"
		case "sop":
			viewType = "sop-view"
		default:
			viewType = "sop-view"
		}

		url := fmt.Sprintf("%s/dashboard/master/%s/%d", *input.Url, viewType, *input.ReferenceID)
		linkHTML := fmt.Sprintf(`<p><strong>Link SOP/SPK : </strong><a href="%s" target="_blank" rel="noopener noreferrer nofollow">%s</a></p>`, url, url)

		if input.Description != nil && len(*input.Description) > 0 {
			desc := *input.Description
			if strings.HasPrefix(desc, "<p><strong>Link SOP/SPK : </strong><a href=") {
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

	payload, err := mapper.UpdateSopJobDtoToModel(input)
	if err != nil {
		return nil, err
	}

	updatedData, err := sopJobService.UpdateSopJob(id, payload)
	if err != nil {
		return nil, err
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedData, nil
}

func (h *SopJobHandler) UpdateSopJobGraphHandler(id int64, input *dto.UpdateSopJobDto) (*graphdb.SopJobNode, error) {
	graphNode := &graphdb.SopJobNode{
		ID:        id,
		UpdatedAt: time.Now().Format(time.RFC3339Nano),
	}

	if input.Name != nil {
		graphNode.Name = *input.Name
	}
	if input.Alias != nil {
		graphNode.Alias = *input.Alias
	}
	if input.Type != nil {
		graphNode.Type = *input.Type
	}
	if input.Description != nil {
		graphNode.Description = fmt.Sprintf("%v", *input.Description)
	}

	if err := h.Service.UpdateGraphSopJob(graphNode); err != nil {
		return nil, fmt.Errorf("failed to update graph SOP Job: %w", err)
	}

	return graphNode, nil
}

func (h *SopJobHandler) DeleteSopJobHandler(id int64, isPermanent bool) error {
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

	sopJobService := h.Service.WithTx(db)

	if err := sopJobService.DeleteSopJob(id, isPermanent); err != nil {
		return err
	}

	if err := sopJobService.DeleteGraphSopJob(id); err != nil {
		return fmt.Errorf("failed to delete graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}

func (h *SopJobHandler) DeleteSopJobSqlHandler(id int64, isPermanent bool) error {
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

	sopJobService := h.Service.WithTx(db)

	if err := sopJobService.DeleteSopJob(id, isPermanent); err != nil {
		return err
	}

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}

func (h *SopJobHandler) DeleteSopJobGraphHandler(id int64) error {
	if err := h.Service.DeleteGraphSopJob(id); err != nil {
		return fmt.Errorf("failed to delete graph SOP Job: %w", err)
	}

	return nil
}

func (h *SopJobHandler) GetSopJobByIDHandler(id int64, filter dto.SopJobFilterDto) (*models.SopJob, error) {
	return h.Service.GetSopJobByID(id, filter)
}

func (h *SopJobHandler) GetSopJobByIdGraphHandler(id int64) (*graphdb.SopJobNode, error) {
	return h.Service.GetGraphSopJobByID(id)
}

func (h *SopJobHandler) GetAllSopJobsHandler(filter dto.SopJobFilterDto) ([]models.SopJob, int64, error) {
	data, err := h.Service.GetAllSopJobs(filter)
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
	if err := db.Model(&models.SopJob{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (h *SopJobHandler) GetAllSopJobsGraphHandler(filter dto.SopJobFilterDto) ([]*graphdb.SopJobNode, int64, error) {
	data, err := h.Service.GetAllGraphSopJobs(filter)
	if err != nil {
		return nil, 0, err
	}

	total, err := h.Service.CountGraphSopJobs(filter)
	if err != nil {
		return nil, 0, err
	}

	return data, total, nil
}

func (h *SopJobHandler) BulkCreateSopJobsHandler(input *dto.BulkCreateSopJobs) ([]*models.SopJob, error) {
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

	sopJobService := h.Service.WithTx(db)
	var sopJobs []*models.SopJob

	for _, createDto := range input.Data {
		sopJob, err := mapper.CreateSopJobDtoToModel(createDto)
		if err != nil {
			return nil, err
		}
		if sopJob != nil {
			sopJobs = append(sopJobs, sopJob)
		}
	}

	createdSopJobs, err := sopJobService.BulkCreateSopJobs(sopJobs)
	if err != nil {
		return nil, err
	}

	var graphNodes []*graphdb.SopJobNode
	for _, sqlData := range createdSopJobs {
		graphNodes = append(graphNodes, &graphdb.SopJobNode{
			ID:          sqlData.ID,
			Name:        sqlData.Name,
			Alias:       sqlData.Alias,
			Code:        sqlData.Code,
			Description: fmt.Sprintf("%v", sqlData.Description),
		})
	}
	if err := sopJobService.BulkInsertGraphSopJobs(graphNodes); err != nil {
		return nil, fmt.Errorf("failed to bulk insert graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return createdSopJobs, nil
}

func (h *SopJobHandler) BulkUpdateSopJobsHandler(input *dto.BulkUpdateSopJobDto) ([]*models.SopJob, error) {
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

	sopJobService := h.Service.WithTx(db)

	updates, err := mapper.UpdateSopJobDtoToModel(input.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to map update data: %w", err)
	}

	if err := sopJobService.BulkUpdateSopJobs(input.IDs, updates); err != nil {
		return nil, fmt.Errorf("failed to bulk update SopJobs: %w", err)
	}

	updatedSopJobs, err := sopJobService.GetSopJobByIDs(input.IDs)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated SopJobs: %w", err)
	}

	var graphNodes []*graphdb.SopJobNode
	for _, sqlData := range updatedSopJobs {
		graphNodes = append(graphNodes, &graphdb.SopJobNode{
			ID:          sqlData.ID,
			Name:        sqlData.Name,
			Alias:       sqlData.Alias,
			Code:        sqlData.Code,
			Description: fmt.Sprintf("%v", sqlData.Description),
		})
	}
	if err := sopJobService.BulkUpdateGraphSopJobs(graphNodes); err != nil {
		return nil, fmt.Errorf("failed to bulk update graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return nil, err
	}
	committed = true

	return updatedSopJobs, nil
}

func (h *SopJobHandler) BulkDeleteSopJobsHandler(input *dto.BulkDeleteSopJobDto, isPermanent bool) error {
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

	sopJobService := h.Service.WithTx(db)

	if err := sopJobService.BulkDeleteSopJobs(input.IDs, isPermanent); err != nil {
		return err
	}

	if err := sopJobService.BulkDeleteGraphSopJobs(input.IDs); err != nil {
		return fmt.Errorf("failed to bulk delete graph: %w", err)
	}

	if err := db.Commit().Error; err != nil {
		return err
	}
	committed = true

	return nil
}

func (h *SopJobHandler) ReorderSopJobHandler(id int64, input *dto.ReorderSopJobDto) error {
	return h.Service.ReorderSopJob(id, input.NewIndex, input.SopID)
}

func (h *SopJobHandler) CountSopJobsHandler(filter dto.SopJobFilterDto) (int64, error) {
	return h.Service.CountSopJobs(filter)
}
