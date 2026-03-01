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

type LevelService interface {
	WithTx(tx *gorm.DB) LevelService

	CreateLevel(input *models.Level) (*models.Level, error)
	UpdateLevel(id int64, updates map[string]interface{}) (*models.Level, error)
	DeleteLevel(id int64) error
	GetAllLevels(filter dto.LevelFilterDto) ([]models.Level, error)
	GetLevelByID(id int64, filter dto.LevelFilterDto) (*models.Level, error)
	GetLevelsByIDs(ids []int64) ([]*models.Level, error)
	GetDB() *gorm.DB
	BulkCreateLevels(data []*models.Level) ([]*models.Level, error)
	BulkUpdateLevels(ids []int64, updates map[string]interface{}) error
	BulkDeleteLevels(ids []int64) error
}

type levelService struct {
	repo sql.LevelRepository
	tx   *gorm.DB
}

func NewLevelService(repo sql.LevelRepository) LevelService {
	return &levelService{repo: repo}
}

func (s *levelService) WithTx(tx *gorm.DB) LevelService {
	return &levelService{
		repo: s.repo.WithTx(tx),
		tx:   tx,
	}
}

func (s *levelService) GetDB() *gorm.DB {
	if s.tx != nil {
		return s.tx
	}
	return config.DB
}

func (s *levelService) CreateLevel(input *models.Level) (*models.Level, error) {
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	data, err := s.repo.InsertLevel(input)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *levelService) UpdateLevel(id int64, updates map[string]interface{}) (*models.Level, error) {
	repo := s.repo

	if _, ok := updates["deleted_at"]; ok {
		repo = repo.WithUnscoped()
	}

	if _, err := repo.FindLevelByID(id); err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	updates["updated_at"] = time.Now()
	fmt.Println(updates)

	data, err := repo.UpdateLevel(id, updates)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *levelService) DeleteLevel(id int64) error {
	_, err := s.GetLevelByID(id, dto.LevelFilterDto{Preload: true})
	if err != nil {
		return err
	}

	err = s.repo.RemoveLevel(id)
	return gorm_err.TranslateGormError(err)
}

func (s *levelService) GetAllLevels(filter dto.LevelFilterDto) ([]models.Level, error) {
	repo := s.repo
	if filter.Preload {
		repo = repo.WithPreloads()
	}
	if filter.Sort != "" && filter.Order != "" {
		orderClause := filter.Sort + " " + filter.Order
		repo = repo.WithOrder(orderClause)
	}
	if filter.ShowDeleted {
		repo = repo.WithUnscoped().WithWhere("levels.deleted_at IS NOT NULL")
	}
	data, err := repo.FindLevel()
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *levelService) GetLevelByID(id int64, filter dto.LevelFilterDto) (*models.Level, error) {
	repo := s.repo
	if filter.Preload {
		repo = repo.WithPreloads()
	}
	data, err := repo.FindLevelByID(id)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *levelService) BulkCreateLevels(data []*models.Level) ([]*models.Level, error) {
	datas, err := s.repo.InsertManyLevels(data)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return datas, nil
}

func (s *levelService) BulkUpdateLevels(ids []int64, updates map[string]interface{}) error {
	repo := s.repo

	// Check if we're trying to restore (deleted_at is being updated)
	if _, ok := updates["deleted_at"]; ok {
		repo = repo.WithUnscoped()
		// Don't force deleted_at to nil, let the actual value from request be used
	}

	err := repo.UpdateManyLevels(ids, updates)
	return gorm_err.TranslateGormError(err)
}

func (s *levelService) GetLevelsByIDs(ids []int64) ([]*models.Level, error) {
	data, err := s.repo.FindLevelsByIDs(ids)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *levelService) BulkDeleteLevels(ids []int64) error {
	err := s.repo.RemoveManyLevels(ids)
	return gorm_err.TranslateGormError(err)
}
