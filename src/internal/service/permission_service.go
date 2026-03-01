package service

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/pkg/errors/gorm_err"
	"jk-api/internal/repository/sql"
	"time"

	"gorm.io/gorm"
)

type PermissionService interface {
	WithTx(tx *gorm.DB) PermissionService

	CreatePermission(input *models.Permission) (*models.Permission, error)
	UpdatePermission(id int64, updates map[string]interface{}) (*models.Permission, error)
	DeletePermission(id int64) error
	GetAllPermissions(filter dto.PermissionFilterDto) ([]models.Permission, error)
	GetPermissionByID(id int64, filter dto.PermissionFilterDto) (*models.Permission, error)
	GetDB() *gorm.DB
}

type permissionService struct {
	repo sql.PermissionRepository
	tx   *gorm.DB
}

func NewPermissionService(repo sql.PermissionRepository) PermissionService {
	return &permissionService{repo: repo}
}

func (s *permissionService) WithTx(tx *gorm.DB) PermissionService {
	return &permissionService{
		repo: s.repo.WithTx(tx),
		tx:   tx,
	}
}

func (s *permissionService) GetDB() *gorm.DB {
	if s.tx != nil {
		return s.tx
	}
	return config.DB
}

func (s *permissionService) CreatePermission(input *models.Permission) (*models.Permission, error) {
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	data, err := s.repo.InsertPermission(input)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *permissionService) UpdatePermission(id int64, updates map[string]interface{}) (*models.Permission, error) {
	if _, err := s.repo.FindPermissionByID(id); err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	updates["updated_at"] = time.Now()
	fmt.Println(updates)

	data, err := s.repo.UpdatePermission(id, updates)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *permissionService) DeletePermission(id int64) error {
	err := s.repo.RemovePermission(id)
	return gorm_err.TranslateGormError(err)
}

func (s *permissionService) GetAllPermissions(filter dto.PermissionFilterDto) ([]models.Permission, error) {
	repo := s.repo
	if filter.RoleID != 0 {
		repo = repo.
			WithJoins("JOIN role_has_permissions ON role_has_permissions.permission_id = permissions.id").
			WithWhere("role_has_permissions.role_id = ?", filter.RoleID)
	}
	if filter.Preload {
		repo = repo.WithPreloads("HasRole")
	}
	data, err := repo.FindPermission()
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *permissionService) GetPermissionByID(id int64, filter dto.PermissionFilterDto) (*models.Permission, error) {
	repo := s.repo
	if filter.RoleID != 0 {
		repo = repo.WithWhere("role_id = ?", filter.RoleID)
	}
	if filter.Preload {
		repo = repo.WithPreloads("HasRole")
	}
	data, err := repo.FindPermissionByID(id)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}
