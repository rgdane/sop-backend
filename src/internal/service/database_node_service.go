package service

import (
	"fmt"
	"time"

	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/pkg/errors/gorm_err"
	"jk-api/internal/repository/sql"
	"jk-api/pkg/neo4j/builder"

	"gorm.io/gorm"
)

type DatabaseNodeService interface {
	WithTx(tx *gorm.DB) DatabaseNodeService
	GetDB() *gorm.DB

	CreateDatabaseNode(input *models.DatabaseNode) (*models.DatabaseNode, error)
	UpdateDatabaseNode(id int64, updates map[string]interface{}) (*models.DatabaseNode, error)
	DeleteDatabaseNode(id int64) error

	GetAllDatabaseNodes(filter dto.DatabaseNodeFilter) ([]models.DatabaseNode, error)
	GetDatabaseNodeByID(id int64, filter dto.DatabaseNodeFilter) (*models.DatabaseNode, error)
	GetDatabaseNodesByIDs(ids []int64) ([]*models.DatabaseNode, error)
	CountDatabaseNodes(filter dto.DatabaseNodeFilter) (int64, error)

	BulkCreateDatabaseNodes(data []*models.DatabaseNode) ([]*models.DatabaseNode, error)
	BulkUpdateDatabaseNodes(ids []int64, updates map[string]interface{}) error
	BulkDeleteDatabaseNodes(ids []int64) error
}

type databaseNodeService struct {
	repo sql.DatabaseNodeRepository
	tx   *gorm.DB
}

func NewDatabaseNodeService(repo sql.DatabaseNodeRepository) DatabaseNodeService {
	return &databaseNodeService{
		repo: repo,
	}
}

func (s *databaseNodeService) WithTx(tx *gorm.DB) DatabaseNodeService {
	return &databaseNodeService{
		repo: s.repo.WithTx(tx),
		tx:   tx,
	}
}

func (s *databaseNodeService) GetDB() *gorm.DB {
	if s.tx != nil {
		return s.tx
	}
	return config.DB
}

func (s *databaseNodeService) CreateDatabaseNode(input *models.DatabaseNode) (*models.DatabaseNode, error) {
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	data, err := s.repo.InsertDatabaseNode(input)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	if input.GraphRef != "" {
		graph := builder.NewGraphRepository()
		params := map[string]interface{}{
			"elementId": input.GraphRef,
			"editable":  false,
		}
		err = graph.
			WithMatch("(n:Job)").
			WithWhere("elementId(n) = $elementId", nil).
			WithRemove("n:Job", nil).
			WithSet("n:Table", nil).
			WithSet("n.editable = $editable", nil).
			WithReturn("labels(n)").
			WithParams(params).
			RunWrite()
		if err != nil {
			return nil, fmt.Errorf("neo4j sync failed: %w", err)
		}
	}

	return data, nil
}

func (s *databaseNodeService) UpdateDatabaseNode(id int64, updates map[string]interface{}) (*models.DatabaseNode, error) {
	repo := s.repo

	data, err := repo.UpdateDatabaseNode(id, updates)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	return data, nil
}

func (s *databaseNodeService) DeleteDatabaseNode(id int64) error {
	err := s.repo.RemoveDatabaseNode(id)
	return gorm_err.TranslateGormError(err)
}

func (s *databaseNodeService) buildFilterQuery(repo sql.DatabaseNodeRepository, filter dto.DatabaseNodeFilter) sql.DatabaseNodeRepository {
	if filter.Search != "" {
		searchQuery := fmt.Sprintf("%%%s%%", filter.Search)
		repo = repo.WithWhere("name ILIKE ? OR table_ref ILIKE ? OR graph_ref ILIKE ?", searchQuery, searchQuery, searchQuery)
	}
	if filter.Cursor != 0 {
		repo = repo.WithCursor(int(filter.Cursor))
	}
	if filter.Limit != 0 {
		repo = repo.WithLimit(int(filter.Limit))
	}
	// if filter.Preload {
	// }
	if filter.ShowDeleted {
		repo = repo.WithUnscoped().WithWhere("deleted_at IS NOT NULL")
	}
	return repo
}

func (s *databaseNodeService) GetAllDatabaseNodes(filter dto.DatabaseNodeFilter) ([]models.DatabaseNode, error) {
	repo := s.repo
	repo = s.buildFilterQuery(repo, filter)

	data, err := repo.FindDatabaseNodes()
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *databaseNodeService) CountDatabaseNodes(filter dto.DatabaseNodeFilter) (int64, error) {
	repo := s.repo

	filter.Limit = 0
	filter.Cursor = 0

	repo = s.buildFilterQuery(repo, filter)

	data, err := repo.CountDatabaseNodes()
	if err != nil {
		return 0, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *databaseNodeService) GetDatabaseNodeByID(id int64, filter dto.DatabaseNodeFilter) (*models.DatabaseNode, error) {
	repo := s.repo

	if filter.Preload {
	}

	data, err := repo.FindDatabaseNodeByID(id)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *databaseNodeService) BulkCreateDatabaseNodes(data []*models.DatabaseNode) ([]*models.DatabaseNode, error) {
	datas, err := s.repo.InsertManyDatabaseNodes(data)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return datas, nil
}

func (s *databaseNodeService) BulkUpdateDatabaseNodes(ids []int64, updates map[string]interface{}) error {
	repo := s.repo

	if _, ok := updates["deleted_at"]; ok {
		repo = repo.WithUnscoped()
	}

	err := repo.UpdateManyDatabaseNodes(ids, updates)
	return gorm_err.TranslateGormError(err)
}

func (s *databaseNodeService) GetDatabaseNodesByIDs(ids []int64) ([]*models.DatabaseNode, error) {
	data, err := s.repo.FindDatabaseNodesByIDs(ids)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *databaseNodeService) BulkDeleteDatabaseNodes(ids []int64) error {
	err := s.repo.RemoveManyDatabaseNodes(ids)
	return gorm_err.TranslateGormError(err)
}
