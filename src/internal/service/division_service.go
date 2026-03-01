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

type DivisionService interface {
	WithTx(tx *gorm.DB) DivisionService

	CreateDivision(input *models.Division) (*models.Division, error)
	UpdateDivision(id int64, updates map[string]interface{}) (*models.Division, error)
	DeleteDivision(id int64) error
	GetAllDivisions(filter dto.DivisionFilterDto) ([]models.Division, error)
	GetDivisionByID(id int64, filter dto.DivisionFilterDto) (*models.Division, error)
	GetDivisionsByIDs(ids []int64) ([]*models.Division, error)
	GetDB() *gorm.DB
	BulkCreateDivisions(data []*models.Division) ([]*models.Division, error)
	BulkUpdateDivisions(ids []int64, updates map[string]interface{}) error
	BulkDeleteDivisions(ids []int64) error
}

type divisionService struct {
	repo sql.DivisionRepository
	tx   *gorm.DB
}

func NewDivisionService(repo sql.DivisionRepository) DivisionService {
	return &divisionService{repo: repo}
}

func (s *divisionService) WithTx(tx *gorm.DB) DivisionService {
	return &divisionService{
		repo: s.repo.WithTx(tx),
		tx:   tx,
	}
}

func (s *divisionService) GetDB() *gorm.DB {
	if s.tx != nil {
		return s.tx
	}
	return config.DB
}

func (s *divisionService) CreateDivision(input *models.Division) (*models.Division, error) {
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	data, err := s.repo.InsertDivision(input)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *divisionService) UpdateDivision(id int64, updates map[string]interface{}) (*models.Division, error) {
	repo := s.repo

	if _, ok := updates["deleted_at"]; ok {
		repo = repo.WithUnscoped()
	}

	if _, err := repo.FindDivisionByID(id); err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	updates["updated_at"] = time.Now()
	fmt.Println(updates)

	data, err := s.repo.UpdateDivision(id, updates)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *divisionService) DeleteDivision(id int64) error {
	err := s.repo.RemoveDivision(id)
	return gorm_err.TranslateGormError(err)
}

func (s *divisionService) GetAllDivisions(filter dto.DivisionFilterDto) ([]models.Division, error) {
	repo := s.repo

	if filter.Preload {
		repo = repo.WithPreloads("HasDepartment")
	}

	if filter.Sort != "" && filter.Order != "" {
		orderClause := filter.Sort + " " + filter.Order
		repo = repo.WithOrder(orderClause)
	}

	if filter.ShowDeleted {
		repo = repo.WithUnscoped().WithWhere("divisions.deleted_at IS NOT NULL")
	}
	fmt.Println("filter", filter)

	data, err := repo.FindDivision()
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *divisionService) GetDivisionByID(id int64, filter dto.DivisionFilterDto) (*models.Division, error) {
	repo := s.repo
	if filter.Preload {
		repo = repo.WithPreloads("HasDepartment")
	}
	data, err := repo.FindDivisionByID(id)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *divisionService) BulkCreateDivisions(data []*models.Division) ([]*models.Division, error) {
	datas, err := s.repo.InsertManyDivisions(data)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return datas, nil
}

func (s *divisionService) BulkUpdateDivisions(ids []int64, updates map[string]interface{}) error {
	repo := s.repo

	if _, ok := updates["deleted_at"]; ok {
		repo = s.repo.WithUnscoped()
	}

	err := repo.UpdateManyDivisions(ids, updates)
	return gorm_err.TranslateGormError(err)
}

func (s *divisionService) GetDivisionsByIDs(ids []int64) ([]*models.Division, error) {
	data, err := s.repo.FindDivisionsByIDs(ids)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *divisionService) BulkDeleteDivisions(ids []int64) error {
	err := s.repo.RemoveManyDivisions(ids)
	return gorm_err.TranslateGormError(err)
}
