package handlers

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/api/http/controllers/v1/mapper"
	"jk-api/internal/service"
)

type AuthHandler struct {
	AuthService service.AuthService
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{AuthService: service}
}

func (h *AuthHandler) GetProfileHandler(token string) (*dto.ProfileResponse, error) {
	user, workspace, err := h.AuthService.GetProfile(token)
	fmt.Println(user)
	if err != nil {
		return nil, err
	}
	data, _ := mapper.AuthModelToProfile(user, workspace)
	return data, nil
}

func (h *AuthHandler) Login(req *dto.LoginRequest) (*dto.LoginResponse, string, error) {
	user, err := h.AuthService.Login(req.Email, req.Password)
	if err != nil {
		return nil, "", err
	}

	token, err := h.AuthService.GenerateToken(user.ID, user.Name)
	if err != nil {
		return nil, "", err
	}

	data, err := mapper.AuthModelToDto(user, token)
	if err != nil {
		return nil, "", err
	}

	return data, token, nil
}
