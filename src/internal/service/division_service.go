package service

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/internal/repository/sql"
	"jk-api/pkg/errors/gorm_err"
	"jk-api/pkg/neo4j/builder"
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

	if err := s.insertGraphDivision(data); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
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

	if err := s.updateGraphDivision(data); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}

	return data, nil
}

func (s *divisionService) DeleteDivision(id int64) error {
	err := s.repo.RemoveDivision(id)

	if err := s.deleteGraphDivision(id); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

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

	if err := s.bulkInsertGraphDivisions(datas); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}

	return datas, nil
}

func (s *divisionService) BulkUpdateDivisions(ids []int64, updates map[string]interface{}) error {
	repo := s.repo

	if _, ok := updates["deleted_at"]; ok {
		repo = s.repo.WithUnscoped()
	}

	err := repo.UpdateManyDivisions(ids, updates)
	
	// Fetch updated data untuk sync ke Neo4j
	updatedJobs, err := s.repo.FindDivisionsByIDs(ids)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	// ✅ BULK UPDATE - SEKALI JALAN!
	if err := s.bulkUpdateGraphDivisions(updatedJobs); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

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

	if err := s.bulkDeleteGraphDivisions(ids); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	return gorm_err.TranslateGormError(err)
}

func (s *divisionService) insertGraphDivision(data *models.Division) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"name":        data.Name,
		"code":        data.Code,
		"id":          data.ID,
		"createdAt":   data.CreatedAt.Format(time.RFC3339Nano),
		"updatedAt":   data.UpdatedAt.Format(time.RFC3339Nano),
	}

	if err := graph.
		WithMerge("(s:Division {id: $id, name: $name, code: $code, id: $id})").
		WithSet("s.created_at = datetime($createdAt), s.updated_at = datetime($updatedAt)", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to merge Division node: %w", err)
	}

	return nil
}

func (s *divisionService) updateGraphDivision(data *models.Division) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"name":        data.Name,
		"code":        data.Code,
		"id":          data.ID,
	}

	if err := graph.
		WithMatch("(s:Division {id: $id})").
		WithSet("s.name = $name, s.code = $code", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update Division graph with id %d: %w", data.ID, err)
	}

	return nil
}

func (s *divisionService) deleteGraphDivision(divisionId int64) error {
    graph := builder.NewGraphRepository()
    
    params := map[string]interface{}{
        "docId":     divisionId,
        "deletedAt": time.Now().Format(time.RFC3339Nano),
    }

    // Karena Division adalah titik akhir (leaf node) dari relasi,
    // kita cukup melakukan MATCH langsung pada node tersebut tanpa perlu APOC traversal.
    if err := graph.
        WithMatch("(d:Division {id: $docId})").
        WithSet("d.deleted_at = $deletedAt", nil).
        WithParams(params).
        RunWrite(); err != nil {
        return fmt.Errorf("failed to soft delete Division graph with id %d: %w", divisionId, err)
    }
    
    return nil
}

func (s *divisionService) bulkInsertGraphDivisions(data []*models.Division) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	// 1. Siapkan batch data
	divisionNodes := make([]map[string]any, 0, len(data))
	for _, div := range data {
		divisionNodes = append(divisionNodes, map[string]any{
			"id":   div.ID,
			"code": div.Code,
			"name": div.Name,
		})
	}

	params := map[string]any{
		"divisions": divisionNodes,
	}

	// 2. Eksekusi UNWIND untuk Bulk Merge
	if err := graph.
		WithUnwind("$divisions", "div").
		WithMerge("(d:Division {id: div.id})").
		WithSet(`d.code = div.code, d.name = div.name`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk insert Division nodes: %w", err)
	}

	return nil
}

func (s *divisionService) bulkUpdateGraphDivisions(data []*models.Division) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	// 1. Siapkan batch data
	divisionNodes := make([]map[string]any, 0, len(data))
	for _, div := range data {
		divisionNodes = append(divisionNodes, map[string]any{
			"id":   div.ID,
			"code": div.Code,
			"name": div.Name,
		})
	}

	params := map[string]any{
		"divisions": divisionNodes,
	}

	// 2. Eksekusi UNWIND untuk Bulk Update.
	// Kita gunakan MATCH karena kita hanya ingin update node yang sudah ada.
	if err := graph.
		WithUnwind("$divisions", "div").
		WithMatch("(d:Division {id: div.id})").
		WithSet("d.code = div.code, d.name = div.name", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk update Division nodes: %w", err)
	}

	return nil
}

func (s *divisionService) bulkDeleteGraphDivisions(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	params := map[string]any{
		"divisionIds": ids,
		// Gunakan format RFC3339Nano agar konsisten dengan created_at (seperti kesepakatan sebelumnya)
		"deletedAt":   time.Now().Format(time.RFC3339Nano), 
	}

	// Bulk Soft Delete sederhana tanpa APOC
	if err := graph.
		WithMatch("(d:Division)").
		WithWhere("d.id IN $divisionIds", nil).
		WithSet("d.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk soft delete Division nodes: %w", err)
	}

	return nil
}