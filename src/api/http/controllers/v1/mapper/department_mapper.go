package mapper

import (
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/database/models"
)

func CreateDepartmentDtoToModel(dto *dto.CreateDepartmentDto) (*models.Department, error) {
	if dto == nil {
		return nil, nil
	}

	data := &models.Department{
		Name: dto.Name,
		Code: dto.Code,
	}

	return data, nil
}

func UpdateDepartmentDtoToModel(dto *dto.UpdateDepartmentDto) (map[string]interface{}, error) {
	if dto == nil {
		return nil, nil
	}

	updates := map[string]interface{}{}

	if dto.Name != nil {
		updates["name"] = *dto.Name
	}
	if dto.Code != nil {
		updates["code"] = *dto.Code
	}
	updates["deleted_at"] = dto.DeletedAt

	return updates, nil
}

func DepartmentModelToResponseDto(data *models.Department) (*dto.DepartmentResponseDto, error) {
	if data == nil {
		return nil, nil
	}

	responseDto := &dto.DepartmentResponseDto{
		Department: *data,
	}

	return responseDto, nil
}
