package service

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/pkg/errors/gorm_err"
	"jk-api/internal/repository/sql"
	"jk-api/pkg/neo4j/builder"
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

	if err := s.insertGraphTitle(data); err != nil {
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

	if err := s.updateGraphTitle(data); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}

	return data, nil
}

func (s *titleService) DeleteTitle(id int64) error {
	err := s.repo.RemoveTitle(id)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}
	if err := s.deleteGraphTitle(id); err != nil {
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

	if err := s.bulkInsertGraphTitles(datas); err != nil {
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

		// Fetch updated data untuk sync ke Neo4j
	updatedJobs, err := s.repo.FindTitlesByIDs(ids)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	// ✅ BULK UPDATE - SEKALI JALAN!
	if err := s.bulkUpdateGraphTitles(updatedJobs); err != nil {
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

	if err := s.bulkDeleteGraphTitles(ids); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	return gorm_err.TranslateGormError(err)
}

func (s *titleService) insertGraphTitle(data *models.Title) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"name":      data.Name,
		"code":      data.Code,
		"id":        data.ID,
		"createdAt": data.CreatedAt.Format(time.RFC3339Nano),
		"updatedAt": data.UpdatedAt.Format(time.RFC3339Nano),
	}

	// Aku hapus duplikasi id: $id yang ada di versi division sebelumnya
	if err := graph.
		WithMerge("(t:Title {id: $id, name: $name, code: $code})").
		WithSet("t.created_at = datetime($createdAt), t.updated_at = datetime($updatedAt)", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to merge Title node: %w", err)
	}

	return nil
}

func (s *titleService) updateGraphTitle(data *models.Title) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"name": data.Name,
		"code": data.Code,
		"id":   data.ID,
	}

	if err := graph.
		WithMatch("(t:Title {id: $id})").
		WithSet("t.name = $name, t.code = $code", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update Title graph with id %d: %w", data.ID, err)
	}

	return nil
}

func (s *titleService) deleteGraphTitle(titleId int64) error {
	graph := builder.NewGraphRepository()

	params := map[string]interface{}{
		"docId":     titleId,
		"deletedAt": time.Now().Format(time.RFC3339Nano),
	}

	// Title juga merupakan node ujung, jadi kita langsung MATCH dan SET soft delete
	if err := graph.
		WithMatch("(t:Title {id: $docId})").
		WithSet("t.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to soft delete Title graph with id %d: %w", titleId, err)
	}

	return nil
}

func (s *titleService) bulkInsertGraphTitles(data []*models.Title) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	// 1. Siapkan batch data
	titleNodes := make([]map[string]any, 0, len(data))
	for _, title := range data {
		titleNodes = append(titleNodes, map[string]any{
			"id":   title.ID,
			"code": title.Code,
			"name": title.Name,
		})
	}

	params := map[string]any{
		"titles": titleNodes,
	}

	// 2. Eksekusi UNWIND untuk Bulk Merge
	if err := graph.
		WithUnwind("$titles", "title").
		WithMerge("(t:Title {id: title.id})").
		WithSet(`t.code = title.code, t.name = title.name`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk insert Title nodes: %w", err)
	}

	return nil
}

func (s *titleService) bulkUpdateGraphTitles(data []*models.Title) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	// 1. Siapkan batch data
	titleNodes := make([]map[string]any, 0, len(data))
	for _, title := range data {
		titleNodes = append(titleNodes, map[string]any{
			"id":   title.ID,
			"code": title.Code,
			"name": title.Name,
		})
	}

	params := map[string]any{
		"titles": titleNodes,
	}

	// 2. Eksekusi UNWIND untuk Bulk Update
	if err := graph.
		WithUnwind("$titles", "title").
		WithMatch("(t:Title {id: title.id})").
		WithSet("t.code = title.code, t.name = title.name", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk update Title nodes: %w", err)
	}

	return nil
}

func (s *titleService) bulkDeleteGraphTitles(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	params := map[string]any{
		"titleIds":  ids,
		"deletedAt": time.Now().Format(time.RFC3339Nano),
	}

	// Bulk Soft Delete sederhana
	if err := graph.
		WithMatch("(t:Title)").
		WithWhere("t.id IN $titleIds", nil).
		WithSet("t.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk soft delete Title nodes: %w", err)
	}

	return nil
}