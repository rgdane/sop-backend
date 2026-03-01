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

type DepartmentService interface {
	WithTx(tx *gorm.DB) DepartmentService

	CreateDepartment(input *models.Department) (*models.Department, error)
	UpdateDepartment(id int64, updates map[string]interface{}) (*models.Department, error)
	DeleteDepartment(id int64) error
	GetAllDepartments(filter dto.DepartmentFilterDto) ([]models.Department, error)
	GetDepartmentByID(id int64, filter dto.DepartmentFilterDto) (*models.Department, error)
	GetDepartmentsByIDs(ids []int64) ([]*models.Department, error)
	GetDB() *gorm.DB
	BulkCreateDepartments(data []*models.Department) ([]*models.Department, error)
	BulkUpdateDepartments(ids []int64, updates map[string]interface{}) error
	BulkDeleteDepartments(ids []int64) error
}

type departmentService struct {
	repo sql.DepartmentRepository
	tx   *gorm.DB
}

func NewDepartmentService(repo sql.DepartmentRepository) DepartmentService {
	return &departmentService{repo: repo}
}

func (s *departmentService) WithTx(tx *gorm.DB) DepartmentService {
	return &departmentService{
		repo: s.repo.WithTx(tx),
		tx:   tx,
	}
}

func (s *departmentService) GetDB() *gorm.DB {
	if s.tx != nil {
		return s.tx
	}
	return config.DB
}

func (s *departmentService) CreateDepartment(input *models.Department) (*models.Department, error) {
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	data, err := s.repo.InsertDepartment(input)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *departmentService) UpdateDepartment(id int64, updates map[string]interface{}) (*models.Department, error) {
	if _, err := s.repo.FindDepartmentByID(id); err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	updates["updated_at"] = time.Now()
	fmt.Println(updates)

	data, err := s.repo.UpdateDepartment(id, updates)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *departmentService) DeleteDepartment(id int64) error {
	err := s.repo.RemoveDepartment(id)
	return gorm_err.TranslateGormError(err)
}

func (s *departmentService) GetAllDepartments(filter dto.DepartmentFilterDto) ([]models.Department, error) {
	repo := s.repo
	if filter.Preload {
		repo = repo.WithPreloads()
	}
	if filter.Sort != "" && filter.Order != "" {
		orderClause := filter.Sort + " " + filter.Order
		repo = repo.WithOrder(orderClause)
	}
	if filter.ShowDeleted {
		repo = repo.WithUnscoped().WithWhere("departments.deleted_at IS NOT NULL")
	}
	data, err := repo.FindDepartment()
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *departmentService) GetDepartmentByID(id int64, filter dto.DepartmentFilterDto) (*models.Department, error) {
	repo := s.repo
	if filter.Preload {
		repo = repo.WithPreloads()
	}
	data, err := repo.FindDepartmentByID(id)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *departmentService) BulkCreateDepartments(data []*models.Department) ([]*models.Department, error) {
	datas, err := s.repo.InsertManyDepartments(data)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return datas, nil
}

func (s *departmentService) BulkUpdateDepartments(ids []int64, updates map[string]interface{}) error {
	repo := s.repo
	if _, ok := updates["deleted_at"]; ok {
		repo = repo.WithUnscoped()
	}
	err := repo.UpdateManyDepartments(ids, updates)
	return gorm_err.TranslateGormError(err)
}

func (s *departmentService) GetDepartmentsByIDs(ids []int64) ([]*models.Department, error) {
	data, err := s.repo.FindDepartmentsByIDs(ids)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *departmentService) BulkDeleteDepartments(ids []int64) error {
	err := s.repo.RemoveManyDepartments(ids)
	return gorm_err.TranslateGormError(err)
}
