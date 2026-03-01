package mapper

import (
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/database/models"
)

func CreateSpkJobDtoToModel(dto *dto.CreateSpkJobDto) (*models.SpkJob, error) {
	if dto == nil {
		return nil, nil
	}

	data := &models.SpkJob{
		Name:        dto.Name,
		Description: dto.Description,
		SpkID:       dto.SpkID,
		SopID:       dto.SopID,
		TitleID:     dto.TitleID,
		Index:       dto.Index,
	}

	return data, nil
}

func UpdateSpkJobDtoToModel(dto *dto.UpdateSpkJobDto) (map[string]interface{}, error) {
	if dto == nil {
		return nil, nil
	}

	updates := map[string]interface{}{}

	if dto.Name != nil {
		updates["name"] = *dto.Name
	}
	if dto.Description != nil {
		updates["description"] = *dto.Description
	}
	if dto.SopID != nil {
		updates["sop_id"] = *dto.SopID
	}
	if dto.TitleID != nil {
		updates["title_id"] = *dto.TitleID
	}
	if dto.Index != nil {
		updates["index"] = *dto.Index
	}
	if dto.FlowchartID != nil {
		updates["flowchart_id"] = *dto.FlowchartID
	}
	if dto.NextIndex != nil {
		updates["next_index"] = *dto.NextIndex
	}
	if dto.PrevIndex != nil {
		updates["prev_index"] = *dto.PrevIndex
	}

	return updates, nil
}

func SpkJobModelToResponseDto(data *models.SpkJob) (*dto.SpkJobResponseDto, error) {
	if data == nil {
		return nil, nil
	}

	responseDto := &dto.SpkJobResponseDto{
		SpkJob: *data,
	}

	return responseDto, nil
}
