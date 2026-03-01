package mapper

import (
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/database/models"
	"jk-api/internal/shared/helper"
)

func CreateDtoToDatabaseNode(input *dto.CreateDatabaseNodeDto) (*models.DatabaseNode, error) {
	return &models.DatabaseNode{
		Name:     input.Name,
		TableRef: input.TableRef,
		GraphRef: input.GraphRef,
	}, nil
}

func UpdateDtoToDatabaseNode(input *dto.UpdateDatabaseNodeDto) (map[string]interface{}, error) {
	result := helper.StructToMap(input)

	return result, nil
}

func ToDatabaseNodeResponseDto(model *models.DatabaseNode) (*dto.DatabaseNodeResponseDto, error) {
	if model == nil {
		return nil, nil // Handle nil model
	}
	return &dto.DatabaseNodeResponseDto{
		DatabaseNode: *model,
	}, nil
}

func ToDatabaseNodeResponseDtoList(models []models.DatabaseNode) []*dto.DatabaseNodeResponseDto {
	responses := make([]*dto.DatabaseNodeResponseDto, len(models))
	for i, model := range models {
		responses[i], _ = ToDatabaseNodeResponseDto(&model)
	}
	return responses
}
