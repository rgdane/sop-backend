package handlers

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/api/http/controllers/v1/mapper"
	"jk-api/internal/database/models"
	"jk-api/internal/service"
	"strings"
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

	// Generate description if sop_id is present
	if input.SopID != nil && input.Url != nil {
		url := fmt.Sprintf("%s/dashboard/master/sop-view/%d", *input.Url, *input.SopID)

		// Create HTML link
		linkHTML := fmt.Sprintf(`<p><strong>Link SOP : </strong><a href="%s" target="_blank" rel="noopener noreferrer nofollow">%s</a></p>`, url, url)

		// Prepend link to existing description
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

	// Generate description if sop_id is present
	if input.SopID != nil && input.Url != nil {
		url := fmt.Sprintf("%s/dashboard/master/sop-view/%d", *input.Url, *input.SopID)

		// Create HTML link
		linkHTML := fmt.Sprintf(`<p><strong>Link SOP : </strong><a href="%s" target="_blank" rel="noopener noreferrer nofollow">%s</a></p>`, url, url)

		if input.Description != nil && len(*input.Description) > 0 {
			desc := *input.Description
			// Check if link already exists at the beginning
			if strings.HasPrefix(desc, "<p><strong>Link SOP : </strong><a href=") {
				// Find the end of the first <p> tag and replace it
				endIdx := strings.Index(desc, "</p>")
				if endIdx != -1 {
					// Replace the first link paragraph with updated URL
					newDesc := linkHTML + desc[endIdx+4:]
					input.Description = &newDesc
				}
			} else {
				// Prepend link
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

func (h *SpkJobHandler) DeleteSpkJobHandler(id int64) error {
	return h.Service.DeleteSpkJob(id)
}

func (h *SpkJobHandler) GetSpkJobByIDHandler(id int64, filter dto.SpkJobFilterDto) (*models.SpkJob, error) {
	return h.Service.GetSpkJobByID(id, filter)
}

func (h *SpkJobHandler) GetAllSpkJobsHandler(filter dto.SpkJobFilterDto) ([]models.SpkJob, error) {
	return h.Service.GetAllSpkJobs(filter)
}

func (h *SpkJobHandler) BulkCreateSpkJobsHandler(input *dto.BulkCreateSpkJobsDto) ([]*models.SpkJob, error) {
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
	return h.Service.BulkCreateSpkJobs(spkJobs)
}

func (h *SpkJobHandler) BulkDeleteHandler(input *dto.BulkDeleteSpkJobDto) error {
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

	if err := spkJobService.BulkDeleteSpkJobs(input.IDs); err != nil {
		return err
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
