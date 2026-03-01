package mapper

import (
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/database/models"
)

func AuthModelToDto(data *models.User, token string) (*dto.LoginResponse, error) {
	if data == nil {
		return nil, nil
	}

	response := &dto.LoginResponse{
		ID:       data.ID,
		Name:     data.Name,
		Email:    data.Email,
		Token:    token,
		HasRoles: data.HasRoles,
	}
	return response, nil
}

func AuthModelToProfile(data *models.User, workspace string) (*dto.ProfileResponse, error) {
	if data == nil {
		return nil, nil
	}

	response := &dto.ProfileResponse{
		ID:                data.ID,
		Name:              data.Name,
		Email:             data.Email,
		HasRoles:          data.HasRoles,
		HasTitle:          data.HasTitle,
		IsPasswordDefault: data.IsPasswordDefault,
		Workspace:         workspace,
	}
	return response, nil
}
