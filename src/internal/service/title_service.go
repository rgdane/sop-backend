package service

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/internal/repository/graphdb"
	"jk-api/internal/repository/sql"
	"jk-api/pkg/errors/gorm_err"
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

	GetAllGraphTitles(filter dto.TitleFilterDto) ([]*graphdb.TitleNode, error)
	GetGraphTitleByID(id int64) (*graphdb.TitleNode, error)
	InsertGraphTitle(data *graphdb.TitleNode) error
	UpdateGraphTitle(data *graphdb.TitleNode) error
	DeleteGraphTitle(titleId int64) error
	BulkInsertGraphTitles(data []*graphdb.TitleNode) error
	BulkUpdateGraphTitles(data []*graphdb.TitleNode) error
	BulkDeleteGraphTitles(ids []int64) error

	CountGraphTitles(filter dto.TitleFilterDto) (int64, error)
}

type titleService struct {
	repo      sql.TitleRepository
	graphRepo graphdb.TitleRepository
	tx        *gorm.DB
}

func NewTitleService(repo sql.TitleRepository, graphRepo graphdb.TitleRepository) TitleService {
	return &titleService{repo: repo, graphRepo: graphRepo}
}

func (s *titleService) WithTx(tx *gorm.DB) TitleService {
	return &titleService{
		repo:      s.repo.WithTx(tx),
		graphRepo: s.graphRepo,
		tx:        tx,
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

	if err := s.InsertGraphTitle(toTitleNode(data)); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
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

	if err := s.UpdateGraphTitle(toTitleNode(data)); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}

	return data, nil
}

func (s *titleService) DeleteTitle(id int64) error {
	err := s.repo.RemoveTitle(id)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}
	if err := s.DeleteGraphTitle(id); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}
	return nil
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

	if err := s.BulkInsertGraphTitles(toTitleNodeSlice(datas)); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}

	return datas, nil
}

func (s *titleService) BulkUpdateTitles(ids []int64, updates map[string]interface{}) error {
	repo := s.repo
	if _, ok := updates["deleted_at"]; ok {
		repo = repo.WithUnscoped()
	}
	err := repo.UpdateManyTitles(ids, updates)

	updatedTitles, err := s.repo.FindTitlesByIDs(ids)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	if err := s.BulkUpdateGraphTitles(toTitleNodeSlice(updatedTitles)); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

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

	if err := s.BulkDeleteGraphTitles(ids); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	return gorm_err.TranslateGormError(err)
}

func (s *titleService) GetAllGraphTitles(filter dto.TitleFilterDto) ([]*graphdb.TitleNode, error) {
	return s.graphRepo.GetAllGraphTitles(filter)
}

func (s *titleService) GetGraphTitleByID(id int64) (*graphdb.TitleNode, error) {
	return s.graphRepo.GetGraphTitleByID(id)
}

func (s *titleService) InsertGraphTitle(data *graphdb.TitleNode) error {
	return s.graphRepo.InsertGraphTitle(data)
}

func (s *titleService) UpdateGraphTitle(data *graphdb.TitleNode) error {
	return s.graphRepo.UpdateGraphTitle(data)
}

func (s *titleService) DeleteGraphTitle(titleId int64) error {
	return s.graphRepo.DeleteGraphTitle(titleId)
}

func (s *titleService) BulkInsertGraphTitles(data []*graphdb.TitleNode) error {
	return s.graphRepo.BulkInsertGraphTitles(data)
}

func (s *titleService) BulkUpdateGraphTitles(data []*graphdb.TitleNode) error {
	return s.graphRepo.BulkUpdateGraphTitles(data)
}

func (s *titleService) BulkDeleteGraphTitles(ids []int64) error {
	return s.graphRepo.BulkDeleteGraphTitles(ids)
}

func (s *titleService) CountGraphTitles(filter dto.TitleFilterDto) (int64, error) {
	return s.graphRepo.CountGraphTitles(filter)
}

func toTitleNode(m *models.Title) *graphdb.TitleNode {
	return &graphdb.TitleNode{
		ID:        m.ID,
		Name:      m.Name,
		Code:      m.Code,
		Color:     m.Color,
		CreatedAt: m.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt: m.UpdatedAt.Format(time.RFC3339Nano),
	}
}

func toTitleNodeSlice(m []*models.Title) []*graphdb.TitleNode {
	result := make([]*graphdb.TitleNode, 0, len(m))
	for _, title := range m {
		result = append(result, toTitleNode(title))
	}
	return result
}
