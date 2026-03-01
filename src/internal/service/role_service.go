package service

import (
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/pkg/errors/gorm_err"
	"jk-api/internal/repository/sql"

	"gorm.io/gorm"
)

type RoleService interface {
	WithTx(tx *gorm.DB) RoleService

	CreateRole(input *models.Role) (*models.Role, error)
	UpdateRole(id int64, updates map[string]interface{}, associations map[string]interface{}) (*models.Role, error)
	DeleteRole(id int64) error
	GetAllRoles(filter dto.RoleFilterDto) ([]models.Role, error)
	GetRoleByID(id int64, filter dto.RoleFilterDto) (*models.Role, error)
	GetDB() *gorm.DB
}

type roleService struct {
	repo sql.RoleRepository
	tx   *gorm.DB
}

func NewRoleService(repo sql.RoleRepository) RoleService {
	return &roleService{repo: repo}
}

func (s *roleService) WithTx(tx *gorm.DB) RoleService {
	return &roleService{
		repo: s.repo.WithTx(tx),
		tx:   tx,
	}
}

func (s *roleService) GetDB() *gorm.DB {
	if s.tx != nil {
		return s.tx
	}
	return config.DB
}

func (s *roleService) CreateRole(input *models.Role) (*models.Role, error) {
	data, err := s.repo.InsertRole(input)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *roleService) UpdateRole(
	id int64,
	payload map[string]interface{},
	associations map[string]interface{},
) (*models.Role, error) {

	repo := s.repo

	if len(associations) > 0 {
		var assocNames []string
		for name := range associations {
			assocNames = append(assocNames, name)
		}
		repo = repo.WithAssociations(assocNames...).WithReplacements(associations)
	}

	for key := range associations {
		delete(payload, key)
	}

	updated, err := repo.UpdateRole(id, payload)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	return updated, nil
}

func (s *roleService) DeleteRole(id int64) error {
	err := s.repo.WithAssociations("HasUsers", "HasPermissions").RemoveRole(id)
	return gorm_err.TranslateGormError(err)
}

func (s *roleService) GetAllRoles(filter dto.RoleFilterDto) ([]models.Role, error) {
	repo := s.repo
	if filter.Preload {
		repo = repo.WithPreloads("HasUsers", "HasPermissions")
	}
	data, err := repo.FindRole()
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *roleService) GetRoleByID(id int64, filter dto.RoleFilterDto) (*models.Role, error) {
	repo := s.repo
	if filter.Preload {
		repo = repo.WithPreloads("HasUsers", "HasPermissions")
	}
	data, err := repo.FindRoleByID(id)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}
