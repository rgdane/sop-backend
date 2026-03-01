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

type PositionService interface {
	WithTx(tx *gorm.DB) PositionService

	CreatePosition(input *models.Position) (*models.Position, error)
	UpdatePosition(id int64, updates map[string]interface{}) (*models.Position, error)
	DeletePosition(id int64) error
	GetAllPositions(filter dto.PositionFilterDto) ([]models.Position, error)
	GetPositionByID(id int64, filter dto.PositionFilterDto) (*models.Position, error)
	GetPositionsByIDs(ids []int64) ([]*models.Position, error)
	GetDB() *gorm.DB
	BulkCreatePositions(data []*models.Position) ([]*models.Position, error)
	BulkUpdatePositions(ids []int64, updates map[string]interface{}) error
	BulkDeletePositions(ids []int64) error
}

type positionService struct {
	repo sql.PositionRepository
	tx   *gorm.DB
}

func NewPositionService(repo sql.PositionRepository) PositionService {
	return &positionService{repo: repo}
}

func (s *positionService) WithTx(tx *gorm.DB) PositionService {
	return &positionService{
		repo: s.repo.WithTx(tx),
		tx:   tx,
	}
}

func (s *positionService) GetDB() *gorm.DB {
	if s.tx != nil {
		return s.tx
	}
	return config.DB
}

func (s *positionService) CreatePosition(input *models.Position) (*models.Position, error) {
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	data, err := s.repo.InsertPosition(input)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *positionService) UpdatePosition(id int64, updates map[string]interface{}) (*models.Position, error) {
	if _, err := s.repo.FindPositionByID(id); err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	updates["updated_at"] = time.Now()
	fmt.Println(updates)

	data, err := s.repo.UpdatePosition(id, updates)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *positionService) DeletePosition(id int64) error {
	err := s.repo.RemovePosition(id)
	return gorm_err.TranslateGormError(err)
}

func (s *positionService) GetAllPositions(filter dto.PositionFilterDto) ([]models.Position, error) {
	repo := s.repo

	if filter.Preload {
		repo = repo.WithPreloads("HasDivision")
	}

	if filter.Sort != "" && filter.Order != "" {
		orderClause := filter.Sort + " " + filter.Order
		repo = repo.WithOrder(orderClause)
	}

	if filter.ShowDeleted {
		repo = repo.WithUnscoped().WithWhere("positions.deleted_at IS NOT NULL")
	}
	fmt.Println("filter", filter)

	data, err := repo.FindPosition()
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *positionService) GetPositionByID(id int64, filter dto.PositionFilterDto) (*models.Position, error) {
	repo := s.repo
	if filter.Preload {
		repo = repo.WithPreloads("HasDivision")
	}
	data, err := repo.FindPositionByID(id)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *positionService) BulkCreatePositions(data []*models.Position) ([]*models.Position, error) {
	datas, err := s.repo.InsertManyPositions(data)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return datas, nil
}

func (s *positionService) BulkUpdatePositions(ids []int64, updates map[string]interface{}) error {
	repo := s.repo
	if _, ok := updates["deleted_at"]; ok {
		repo = repo.WithUnscoped()
	}
	err := repo.UpdateManyPositions(ids, updates)
	return gorm_err.TranslateGormError(err)
}

func (s *positionService) GetPositionsByIDs(ids []int64) ([]*models.Position, error) {
	data, err := s.repo.FindPositionsByIDs(ids)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *positionService) BulkDeletePositions(ids []int64) error {
	err := s.repo.RemoveManyPositions(ids)
	return gorm_err.TranslateGormError(err)
}
