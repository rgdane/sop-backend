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

type TitleService interface {
	WithTx(tx *gorm.DB) TitleService

	CreateTitle(input *models.Title) (*models.Title, error)
	UpdateTitle(id int64, updates map[string]interface{}) (*models.Title, error)
	DeleteTitle(id int64) error
	GetAllTitles(filter dto.TitleFilterDto) ([]models.Title, error)
	GetTitleByID(id int64, filter dto.TitleFilterDto) (*models.Title, error)
	GetTitlesByIDs(ids []int64) ([]*models.Title, error)
	GetDB() *gorm.DB
	BulkCreateTitles(data []*models.Title) ([]*models.Title, error)
	BulkUpdateTitles(ids []int64, updates map[string]interface{}) error
	BulkDeleteTitles(ids []int64) error
}

type titleService struct {
	repo sql.TitleRepository
	tx   *gorm.DB
}

func NewTitleService(repo sql.TitleRepository) TitleService {
	return &titleService{repo: repo}
}

func (s *titleService) WithTx(tx *gorm.DB) TitleService {
	return &titleService{
		repo: s.repo.WithTx(tx),
		tx:   tx,
	}
}

func (s *titleService) GetDB() *gorm.DB {
	if s.tx != nil {
		return s.tx
	}
	return config.DB
}

func (s *titleService) CreateTitle(input *models.Title) (*models.Title, error) {
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	data, err := s.repo.InsertTitle(input)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *titleService) UpdateTitle(id int64, updates map[string]interface{}) (*models.Title, error) {
	if _, err := s.repo.FindTitleByID(id); err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	updates["updated_at"] = time.Now()
	fmt.Println(updates)

	data, err := s.repo.UpdateTitle(id, updates)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *titleService) DeleteTitle(id int64) error {
	err := s.repo.RemoveTitle(id)
	return gorm_err.TranslateGormError(err)
}

func (s *titleService) GetAllTitles(filter dto.TitleFilterDto) ([]models.Title, error) {
	repo := s.repo
	if filter.Preload {
		repo = repo.WithPreloads()
	}
	if filter.Sort != "" && filter.Order != "" {
		orderClause := filter.Sort + " " + filter.Order
		repo = repo.WithOrder(orderClause)
	}
	if filter.ShowDeleted {
		repo = repo.WithUnscoped().WithWhere("titles.deleted_at IS NOT NULL")
	}
	data, err := repo.FindTitle()
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	return data, nil
}

func (s *titleService) GetTitleByID(id int64, filter dto.TitleFilterDto) (*models.Title, error) {
	repo := s.repo
	if filter.Preload {
		repo = repo.WithPreloads()
	}
	data, err := repo.FindTitleByID(id)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *titleService) BulkCreateTitles(data []*models.Title) ([]*models.Title, error) {
	datas, err := s.repo.InsertManyTitles(data)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return datas, nil
}

func (s *titleService) BulkUpdateTitles(ids []int64, updates map[string]interface{}) error {
	repo := s.repo
	if _, ok := updates["deleted_at"]; ok {
		repo = repo.WithUnscoped()
	}
	err := repo.UpdateManyTitles(ids, updates)
	return gorm_err.TranslateGormError(err)
}

func (s *titleService) GetTitlesByIDs(ids []int64) ([]*models.Title, error) {
	data, err := s.repo.FindTitlesByIDs(ids)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *titleService) BulkDeleteTitles(ids []int64) error {
	err := s.repo.RemoveManyTitles(ids)
	return gorm_err.TranslateGormError(err)
}
