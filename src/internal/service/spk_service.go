package service

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/pkg/errors/gorm_err"
	"jk-api/internal/shared/helper"
	"jk-api/internal/repository/sql"
	"jk-api/pkg/neo4j/builder"
	"time"

	"gorm.io/gorm"
)

type SpkService interface {
	WithTx(tx *gorm.DB) SpkService

	CreateSpk(input *models.Spk) (*models.Spk, error)
	UpdateSpk(id int64, updates map[string]interface{}, associations map[string]interface{}) (*models.Spk, error)
	DeleteSpk(id int64) error
	GetAllSpks(filter dto.SpkFilterDto) ([]models.Spk, error)
	GetSpkByID(id int64, filter dto.SpkFilterDto) (*models.Spk, error)
	GetSpksByIDs(ids []int64) ([]*models.Spk, error)
	GetDB() *gorm.DB
	BulkCreateSpks(data []*models.Spk) ([]*models.Spk, error)
	BulkUpdateSpks(ids []int64, updates map[string]interface{}) error
	BulkDeleteSpks(ids []int64) error
}

type spkService struct {
	repo sql.SpkRepository
	tx   *gorm.DB
}

func NewSpkService(repo sql.SpkRepository) SpkService {
	return &spkService{repo: repo}
}

func (s *spkService) WithTx(tx *gorm.DB) SpkService {
	return &spkService{
		repo: s.repo.WithTx(tx),
		tx:   tx,
	}
}

func (s *spkService) GetDB() *gorm.DB {
	if s.tx != nil {
		return s.tx
	}
	return config.DB
}

func (s *spkService) CreateSpk(input *models.Spk) (*models.Spk, error) {
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	data, err := s.repo.InsertSpk(input)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	// Sync to Neo4j - create SPK node only
	if err := s.insertGraphSpk(data); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}

	return data, nil
}

func (s *spkService) UpdateSpk(id int64, updates map[string]interface{}, associations map[string]interface{}) (*models.Spk, error) {
	repo := s.repo

	if len(associations) > 0 {
		assocNames := make([]string, 0, len(associations))
		for name := range associations {
			assocNames = append(assocNames, name)
			delete(updates, name)
		}
		repo = repo.WithAssociations(assocNames...).WithReplacements(associations)
	}

	data, err := repo.UpdateSpk(id, updates)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	// Sync to Neo4j after SQL update
	if err := s.updateGraphSpk(data); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}

	return data, nil
}

func (s *spkService) DeleteSpk(id int64) error {
	// Get SPK data before deletion for Neo4j cleanup
	spk, err := s.repo.FindSpkByID(id)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	// Delete from Neo4j graph first
	if err := s.deleteGraphSpk(spk); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	// Then delete from SQL
	err = s.repo.RemoveSpk(id)
	return gorm_err.TranslateGormError(err)
}

func (s *spkService) GetAllSpks(filter dto.SpkFilterDto) ([]models.Spk, error) {
	repo := s.repo

	if filter.Preload {
		repo = repo.WithPreloads("HasTitles", "HasJobs.HasSop")
	}
	if filter.Limit != 0 {
		repo = repo.WithLimit(int(filter.Limit))
	}
	if filter.Cursor != 0 {
		repo = repo.WithCursor(int(filter.Cursor))
	}
	if filter.Name != "" {
		repo = repo.WithWhere("name ILIKE ?", "%"+filter.Name+"%")
	}
	if filter.TitleIDs != 0 {
		repo = repo.WithJoins("JOIN spk_titles ON spk_titles.spk_id = spks.id").WithWhere("spk_titles.title_id = ?", filter.TitleIDs)
	}
	if filter.ShowDeleted {
		repo = repo.WithUnscoped().WithWhere("spks.deleted_at IS NOT NULL")
	}

	data, err := repo.FindSpk()
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *spkService) GetSpkByID(id int64, filter dto.SpkFilterDto) (*models.Spk, error) {
	repo := s.repo
	if filter.Preload {
		repo = repo.WithPreloads("HasTitles", "HasJobs.HasSop")
	}
	data, err := repo.FindSpkByID(id)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *spkService) BulkUpdateSpks(ids []int64, updates map[string]interface{}) error {
	repo := s.repo

	if _, ok := updates["deleted_at"]; ok {
		repo = repo.WithUnscoped()
	}

	err := repo.UpdateManySpks(ids, updates)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	// Sync to Neo4j using batch - reload SPKs and update graph
	spks, err := s.repo.FindSpksByIDs(ids)
	if err != nil {
		fmt.Printf("Failed to load SPKs for graph sync: %v\n", err)
		return nil
	}

	if err := s.batchUpdateGraphSpks(spks); err != nil {
		fmt.Printf("Failed to batch update SPKs in graph: %v\n", err)
	}

	return nil
}

func (s *spkService) BulkCreateSpks(data []*models.Spk) ([]*models.Spk, error) {
	datas, err := s.repo.InsertManySpks(data)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	// Sync to Neo4j using batch
	if err := s.batchInsertGraphSpks(datas); err != nil {
		fmt.Printf("Failed to batch sync SPKs to graph: %v\n", err)
	}

	return datas, nil
}

func (s *spkService) GetSpksByIDs(ids []int64) ([]*models.Spk, error) {
	data, err := s.repo.FindSpksByIDs(ids)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *spkService) BulkDeleteSpks(ids []int64) error {
	// Delete from Neo4j first using batch
	if err := s.batchDeleteGraphSpks(ids); err != nil {
		fmt.Printf("Failed to batch delete SPKs from graph: %v\n", err)
	}

	// Then delete from SQL
	err := s.repo.RemoveManySpks(ids)
	return gorm_err.TranslateGormError(err)
}

// insertGraphSpk creates SPK Document node in Neo4j
func (s *spkService) insertGraphSpk(data *models.Spk) error {
	graph := builder.NewGraphRepository()

	// Create SPK node (not Document) with properties from SQL
	// Convert description JSON map to string for Neo4j
	docParam := map[string]interface{}{
		"id":          data.ID,
		"name":        data.Name,
		"code":        data.Code,
		"description": helper.ToJSONString(data.Description),
	}

	if err := graph.
		WithMerge("(s:SPK {id: $id})").
		WithSet("s.name = $name, s.code = $code, s.description = $description", docParam).
		WithParams(docParam).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to create SPK node: %w", err)
	}

	return nil
}

// updateGraphSpk updates SPK node in Neo4j
func (s *spkService) updateGraphSpk(data *models.Spk) error {
	graph := builder.NewGraphRepository()

	// Update SPK node properties in Neo4j
	docParam := map[string]interface{}{
		"id":          data.ID,
		"name":        data.Name,
		"code":        data.Code,
		"description": helper.ToJSONString(data.Description),
	}

	if err := graph.
		WithMatch("(s:SPK {id: $id})").
		WithSet("s.name = $name, s.code = $code, s.description = $description", docParam).
		WithParams(docParam).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update SPK node: %w", err)
	}

	return nil
}

// deleteGraphSpk removes SPK node and all children from Neo4j
func (s *spkService) deleteGraphSpk(data *models.Spk) error {
	return s.deleteGraphSpkByID(data.ID)
}

func (s *spkService) deleteGraphSpkByID(spkId int64) error {
	graph := builder.NewGraphRepository()

	// Delete SPK node and all children (Jobs) recursively
	params := map[string]interface{}{
		"spkId": spkId,
	}

	if err := graph.
		WithMatch("(s:SPK {id: $spkId})").
		WithOptionalMatch("(s)-[*]->(child)").
		WithDetachDelete("s, child").
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to delete SPK graph with id %d: %w", spkId, err)
	}

	return nil
}

// batchInsertGraphSpks creates multiple SPK nodes in Neo4j using UNWIND batch operation
func (s *spkService) batchInsertGraphSpks(data []*models.Spk) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	// Prepare batch data
	var spkList []map[string]interface{}
	for _, spk := range data {
		spkList = append(spkList, map[string]interface{}{
			"id":          spk.ID,
			"name":        spk.Name,
			"code":        spk.Code,
			"description": helper.ToJSONString(spk.Description),
		})
	}

	params := map[string]interface{}{
		"spks": spkList,
	}

	// Use UNWIND to batch create SPK nodes
	if err := graph.
		WithUnwind("$spks", "spkData").
		WithMerge("(s:SPK {id: spkData.id})").
		WithSet("s.name = spkData.name, s.code = spkData.code, s.description = spkData.description", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to batch insert SPK nodes: %w", err)
	}

	return nil
}

// batchUpdateGraphSpks updates multiple SPK nodes in Neo4j using UNWIND batch operation
func (s *spkService) batchUpdateGraphSpks(data []*models.Spk) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	// Prepare batch data
	var spkList []map[string]interface{}
	for _, spk := range data {
		spkList = append(spkList, map[string]interface{}{
			"id":          spk.ID,
			"name":        spk.Name,
			"code":        spk.Code,
			"description": helper.ToJSONString(spk.Description),
		})
	}

	params := map[string]interface{}{
		"spks": spkList,
	}

	// Use UNWIND to batch update SPK nodes
	if err := graph.
		WithUnwind("$spks", "spkData").
		WithMatch("(s:SPK {id: spkData.id})").
		WithSet("s.name = spkData.name, s.code = spkData.code, s.description = spkData.description", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to batch update SPK nodes: %w", err)
	}

	return nil
}

// batchDeleteGraphSpks deletes multiple SPK nodes from Neo4j using UNWIND batch operation
func (s *spkService) batchDeleteGraphSpks(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	params := map[string]interface{}{
		"spkIds": ids,
	}

	// Use UNWIND to batch delete SPK nodes and their children
	if err := graph.
		WithUnwind("$spkIds", "spkId").
		WithMatch("(s:SPK {id: spkId})").
		WithOptionalMatch("(s)-[*]->(child)").
		WithDetachDelete("s, child").
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to batch delete SPK nodes: %w", err)
	}

	return nil
}
