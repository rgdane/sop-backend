package mapper

import (
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/database/models"
)

func CreateSopMenuDtoToModel(dto *dto.CreateSopMenuDto) (*models.SopMenu, error) {
	if dto == nil {
		return nil, nil
	}

	data := &models.SopMenu{
		Name:       dto.Name,
		SopID:      dto.SopID,
		DivisionID: dto.DivisionID,
		Type:       dto.Type,
		Multiple:   dto.Multiple,
		ParentID:   dto.ParentID,
		MasterID:   dto.MasterID,
		IsMaster:   dto.IsMaster,
	}

	return data, nil
}

func UpdateSopMenuDtoToModel(dto *dto.UpdateSopMenuDto) (map[string]interface{}, error) {
	if dto == nil {
		return nil, nil
	}

	updates := map[string]interface{}{}

	if dto.Name != nil {
		updates["name"] = *dto.Name
	}
	if dto.SopID != nil {
		updates["sop_id"] = *dto.SopID
	}
	if dto.DivisionID != nil {
		updates["division_id"] = *dto.DivisionID
	}
	if dto.Type != nil {
		updates["type"] = *dto.Type
	}
	if dto.Multiple != nil {
		updates["multiple"] = *dto.Multiple
	}
	if dto.ParentID != nil {
		updates["parent_id"] = *dto.ParentID
	}
	if dto.MasterID != nil {
		updates["master_id"] = *dto.MasterID
	}
	if dto.IsMaster != nil {
		updates["is_master"] = *dto.IsMaster
	}

	return updates, nil
}

func SopMenuModelToResponseDto(data *models.SopMenu) (*dto.SopMenuResponseDto, error) {
	if data == nil {
		return nil, nil
	}

	responseDto := &dto.SopMenuResponseDto{
		SopMenu: *data,
	}

	return responseDto, nil
}
