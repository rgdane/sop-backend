package dto

import "jk-api/internal/database/models"

type LoginRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"password123"`
}

type LoginResponse struct {
	ID       int64         `json:"id"`
	Name     string        `json:"name"`
	Email    string        `json:"email"`
	Token    string        `json:"token"`
	HasRoles []models.Role `json:"has_roles"`
}

type ProfileResponse struct {
	ID                int64          `json:"id"`
	Name              string         `json:"name"`
	Email             string         `json:"email"`
	IsPasswordDefault bool           `json:"is_password_default"`
	HasRoles          []models.Role  `json:"has_roles"`
	HasTitle          models.Title   `json:"has_title"`
	Workspace         string         `json:"workspace"`
}
