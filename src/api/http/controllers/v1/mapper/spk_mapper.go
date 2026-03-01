package mapper

import (
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/database/models"
)

func CreateSpkDtoToModel(dto *dto.CreateSpkDto) (*models.Spk, error) {
	if dto == nil {
		return nil, nil
	}

	var titles []models.Title
	for _, id := range dto.HasTitles {
		titles = append(titles, models.Title{ID: id})
	}

	data := &models.Spk{
		Name:        dto.Name,
		Description: dto.Description,
		Code:        dto.Code,
		HasTitles:   titles,
	}

	return data, nil
}

func UpdateSpkDtoToModel(dto *dto.UpdateSpkDto) (
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
	if dto.HasTitles != nil {
		var titles []models.Title
		for _, id := range *dto.HasTitles {
			titles = append(titles, models.Title{ID: id})
		}
		associations["HasTitles"] = titles
	}
	payload["deleted_at"] = dto.DeletedAt

	return payload, associations
}

func SpkModelToResponseDto(data *models.Spk) (*dto.SpkResponseDto, error) {
	if data == nil {
		return nil, nil
	}

	responseDto := &dto.SpkResponseDto{
		Spk: *data,
	}

	return responseDto, nil
}
