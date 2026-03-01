// sop_job_mapper.go - updated

package mapper

import (
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/database/models"
)

func CreateSopJobDtoToModel(dto *dto.CreateSopJobDto) (*models.SopJob, error) {
	if dto == nil {
		return nil, nil
	}

	data := &models.SopJob{
		Name:        dto.Name,
		Alias:       dto.Alias,
		Description: dto.Description,
		Type:        dto.Type,
		TitleID:     dto.TitleID,
		SopID:       dto.SopID,
		ReferenceID: dto.ReferenceID,
		IsPublished: dto.IsPublished,
	}

	return data, nil
}

func UpdateSopJobDtoToModel(dto *dto.UpdateSopJobDto) (map[string]any, error) {
	if dto == nil {
		return nil, nil
	}

	updates := map[string]any{}

	if dto.Name != nil {
		updates["name"] = *dto.Name
	}
	if dto.Alias != nil {
		updates["alias"] = *dto.Alias
	}
	if dto.Type != nil {
		updates["type"] = *dto.Type
	}
	if dto.TitleID != nil {
		updates["title_id"] = *dto.TitleID
	}
	if dto.SopID != nil {
		updates["sop_id"] = *dto.SopID
	}
	if dto.ReferenceID != nil {
		updates["reference_id"] = *dto.ReferenceID
	}
	if dto.Description != nil {
		updates["description"] = *dto.Description
	}
	if dto.NextIndex != nil {
		updates["next_index"] = *dto.NextIndex
	}
	if dto.PrevIndex != nil {
		updates["prev_index"] = *dto.PrevIndex
	}
	if dto.FlowchartID != nil {
		updates["flowchart_id"] = *dto.FlowchartID
	}
	if dto.IsPublished != nil {
		updates["is_published"] = *dto.IsPublished
	}
	if dto.IsHide != nil {
		updates["is_hide"] = *dto.IsHide
	}

	return updates, nil
}

func SopJobModelToResponseDto(data *models.SopJob) (*dto.SopJobResponseDto, error) {
	if data == nil {
		return nil, nil
	}

	responseDto := &dto.SopJobResponseDto{
		SopJob:       *data,
		HasSop:       data.HasSop,
		HasTitle:     data.HasTitle,
		HasReference: data.HasReference,
	}

	return responseDto, nil
}
