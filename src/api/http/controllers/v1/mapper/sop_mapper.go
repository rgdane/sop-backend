package mapper

import (
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/database/models"
)

func CreateSopDtoToModel(dto *dto.CreateSopDto) (*models.Sop, error) {
	if dto == nil {
		return nil, nil
	}

	var titles []models.Title
	for _, id := range dto.HasTitles {
		titles = append(titles, models.Title{ID: id})
	}

	var divisions []models.Division
	for _, id := range dto.HasDivisions {
		divisions = append(divisions, models.Division{ID: id})
	}

	data := &models.Sop{
		Name:         dto.Name,
		Description:  dto.Description,
		Code:         dto.Code,
		HasTitles:    titles,
		HasDivisions: divisions,
		ParentJobID:  dto.ParentJobID,
	}

	return data, nil
}

func UpdateSopDtoToModel(dto *dto.UpdateSopDto) (
	payload map[string]interface{},
	associations map[string]interface{},
) {
	payload = make(map[string]interface{})
	associations = make(map[string]interface{})

	if dto.Name != nil {
		payload["name"] = *dto.Name
	}
	if dto.Description != nil {
		payload["description"] = *dto.Description
	}
	if dto.Code != nil {
		payload["code"] = *dto.Code
	}
	if dto.HasDivisions != nil {
		var divisions []models.Division
		for _, id := range *dto.HasDivisions {
			divisions = append(divisions, models.Division{ID: id})
		}
		associations["HasDivisions"] = divisions
	}
	if dto.HasTitles != nil {
		var titles []models.Title
		for _, id := range *dto.HasTitles {
			titles = append(titles, models.Title{ID: id})
		}
		associations["HasTitles"] = titles
	}
	payload["deleted_at"] = dto.DeletedAt
	if dto.ParentJobID != nil {
		payload["parent_job_id"] = dto.ParentJobID
	}

	return payload, associations
}

func SopModelToResponseDto(data *models.Sop) (*dto.SopResponseDto, error) {
	if data == nil {
		return nil, nil
	}

	responseDto := &dto.SopResponseDto{
		Sop: *data,
	}

	return responseDto, nil
}
