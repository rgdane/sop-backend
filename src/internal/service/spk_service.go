package service

import (
	"fmt"
	"strings"
	"time"

	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/internal/repository/graphdb"
	"jk-api/internal/repository/sql"
	"jk-api/internal/shared/helper"
	"jk-api/pkg/errors/gorm_err"
	"jk-api/pkg/neo4j/builder"

	"gorm.io/gorm"
)

type SpkService interface {
	WithTx(tx *gorm.DB) SpkService

	CreateSpk(input *models.Spk) (*models.Spk, error)
	UpdateSpk(id int64, updates map[string]interface{}, associations map[string]interface{}) (*models.Spk, error)
	DeleteSpk(id int64, isPermanent bool) error
	GetAllSpks(filter dto.SpkFilterDto) ([]models.Spk, error)
	GetSpkByID(id int64, filter dto.SpkFilterDto) (*models.Spk, error)
	GetSpksByIDs(ids []int64) ([]*models.Spk, error)
	GetDB() *gorm.DB
	BulkCreateSpks(data []*models.Spk) ([]*models.Spk, error)
	BulkUpdateSpks(ids []int64, updates map[string]interface{}) error
	BulkDeleteSpks(ids []int64, isPermanent bool) error
	CountSpks(filter dto.SpkFilterDto) (int64, error)

	GetAllGraphSpks(filter dto.SpkFilterDto) ([]*graphdb.SpkNode, error)
	GetGraphSpkByID(id int64) (*graphdb.SpkNode, error)
	InsertGraphSpk(data *graphdb.SpkNode) error
	UpdateGraphSpk(data *graphdb.SpkNode) error
	DeleteGraphSpk(spkId int64) error
	BulkInsertGraphSpks(data []*graphdb.SpkNode) error
	BulkUpdateGraphSpks(data []*graphdb.SpkNode) error
	BulkDeleteGraphSpks(ids []int64) error
	CountGraphSpks(filter dto.SpkFilterDto) (int64, error)
}

type spkService struct {
	repo      sql.SpkRepository
	graphRepo graphdb.SpkRepository
	tx        *gorm.DB
}

func NewSpkService(repo sql.SpkRepository, graphRepo graphdb.SpkRepository) SpkService {
	return &spkService{repo: repo, graphRepo: graphRepo}
}

func (s *spkService) WithTx(tx *gorm.DB) SpkService {
	return &spkService{
		repo:      s.repo.WithTx(tx),
		graphRepo: s.graphRepo,
		tx:        tx,
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

	if err := s.InsertGraphSpk(toSpkNode(data)); err != nil {
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

	if err := s.UpdateGraphSpk(toSpkNode(data)); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}
	return data, nil
}

func (s *spkService) DeleteSpk(id int64, isPermanent bool) (err error) {
	repo := s.repo
	if isPermanent {
		repo = repo.WithUnscoped()
		err = repo.RemoveSpk(id)
		return gorm_err.TranslateGormError(err)
	}

	data, err := repo.FindSpkByID(id)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	payload := map[string]any{
		"name": fmt.Sprintf("DELETED-%s", data.Name),
	}

	if _, err = s.UpdateSpk(id, payload, nil); err != nil {
		return gorm_err.TranslateGormError(err)
	}

	if err := s.DeleteGraphSpk(id); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	err = repo.RemoveSpk(id)
	return gorm_err.TranslateGormError(err)
}

func (s *spkService) GetAllSpks(filter dto.SpkFilterDto) ([]models.Spk, error) {
	repo := s.repo

	if filter.Limit != 0 {
		repo = repo.WithLimit(int(filter.Limit))
	}
	if filter.Cursor != 0 {
		repo = repo.WithCursor(int(filter.Cursor))
	}
	if filter.Preload {
		repo = repo.WithPreloads("HasTitles", "HasJobs.HasSop")
	}
	if filter.Name != "" {
		repo = repo.WithWhere("name ILIKE ?", "%"+filter.Name+"%")
	}
	if filter.TitleIDs != 0 {
		repo = repo.WithJoins("JOIN spk_titles ON spk_titles.spk_id = spks.id").WithWhere("spk_titles.title_id = ?", filter.TitleIDs)
	}
	if filter.ShowDeleted {
		repo = repo.WithUnscoped().WithWhere("deleted_at IS NOT NULL")
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

func (s *spkService) BulkCreateSpks(data []*models.Spk) ([]*models.Spk, error) {
	datas, err := s.repo.InsertManySpks(data)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	if err := s.BulkInsertGraphSpks(toSpkNodeSlice(datas)); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}
	return datas, nil
}

func (s *spkService) BulkUpdateSpks(ids []int64, updates map[string]interface{}) error {
	repo := s.repo

	if _, ok := updates["deleted_at"]; ok {
		repo = repo.WithUnscoped()
		deletedAtValue := updates["deleted_at"]
		shouldRestore := false

		switch v := deletedAtValue.(type) {
		case nil:
			shouldRestore = true
		case *time.Time:
			shouldRestore = (v == nil)
		case time.Time:
			shouldRestore = v.IsZero()
		default:
			shouldRestore = false
		}

		if shouldRestore {
			if err := s.spkRestore(ids); err != nil {
				return err
			}
			err := repo.UpdateManySpks(ids, updates)
			return gorm_err.TranslateGormError(err)
		}
	}

	updatedSpks, err := s.repo.FindSpksByIDs(ids)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	if err := s.BulkUpdateGraphSpks(toSpkNodeSlice(updatedSpks)); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	err = repo.UpdateManySpks(ids, updates)
	return gorm_err.TranslateGormError(err)
}

func (s *spkService) GetSpksByIDs(ids []int64) ([]*models.Spk, error) {
	data, err := s.repo.FindSpksByIDs(ids)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *spkService) BulkDeleteSpks(ids []int64, isPermanent bool) (err error) {
	repo := s.repo

	if isPermanent {
		repo = repo.WithUnscoped()
		err = repo.RemoveManySpks(ids)
		return gorm_err.TranslateGormError(err)
	}

	if err := s.BulkDeleteGraphSpks(ids); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	err = repo.RemoveManySpks(ids)
	return gorm_err.TranslateGormError(err)
}

func (s *spkService) CountSpks(filter dto.SpkFilterDto) (int64, error) {
	repo := s.repo
	if filter.TitleIDs != 0 {
		repo = repo.WithJoins("JOIN spk_titles ON spks.id = spk_titles.spk_id").
			WithWhere("spk_titles.title_id = ?", filter.TitleIDs)
	}

	data, err := repo.CountSpks()
	if err != nil {
		return 0, gorm_err.TranslateGormError(err)
	}

	return data, nil
}

func (s *spkService) GetAllGraphSpks(filter dto.SpkFilterDto) ([]*graphdb.SpkNode, error) {
	return s.graphRepo.GetAllGraphSpks(filter)
}

func (s *spkService) GetGraphSpkByID(id int64) (*graphdb.SpkNode, error) {
	return s.graphRepo.GetGraphSpkByID(id)
}

func (s *spkService) InsertGraphSpk(data *graphdb.SpkNode) error {
	return s.graphRepo.InsertGraphSpk(data)
}

func (s *spkService) UpdateGraphSpk(data *graphdb.SpkNode) error {
	return s.graphRepo.UpdateGraphSpk(data)
}

func (s *spkService) DeleteGraphSpk(spkId int64) error {
	return s.graphRepo.DeleteGraphSpk(spkId)
}

func (s *spkService) BulkInsertGraphSpks(data []*graphdb.SpkNode) error {
	return s.graphRepo.BulkInsertGraphSpks(data)
}

func (s *spkService) BulkUpdateGraphSpks(data []*graphdb.SpkNode) error {
	return s.graphRepo.BulkUpdateGraphSpks(data)
}

func (s *spkService) BulkDeleteGraphSpks(ids []int64) error {
	return s.graphRepo.BulkDeleteGraphSpks(ids)
}

func (s *spkService) CountGraphSpks(filter dto.SpkFilterDto) (int64, error) {
	return s.graphRepo.CountGraphSpks(filter)
}

func toSpkNode(m *models.Spk) *graphdb.SpkNode {
	return &graphdb.SpkNode{
		ID:          m.ID,
		Name:        m.Name,
		Code:        m.Code,
		Description: helper.ToJSONString(m.Description),
		CreatedAt:   m.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   m.UpdatedAt.Format(time.RFC3339Nano),
	}
}

func toSpkNodeSlice(m []*models.Spk) []*graphdb.SpkNode {
	result := make([]*graphdb.SpkNode, 0, len(m))
	for _, spk := range m {
		result = append(result, toSpkNode(spk))
	}
	return result
}

func (s *spkService) spkRestore(ids []int64) error {
	var spks []models.Spk
	if err := s.GetDB().Unscoped().Preload("HasTitles").Where("id IN ? AND deleted_at IS NOT NULL", ids).Find(&spks).Error; err != nil {
		return err
	}

	for _, spk := range spks {
		spk.Name = strings.TrimPrefix(spk.Name, "DELETED-")

		var updatedSpk models.Spk
		if err := s.GetDB().Unscoped().Where("id = ?", spk.ID).First(&updatedSpk).Error; err != nil {
			return err
		}

		if updatedSpk.CreatedAt.IsZero() {
			updatedSpk.CreatedAt = time.Now()
		}
		updatedSpk.UpdatedAt = time.Now()

		if err := s.removeDeletedAtFromGraph(updatedSpk.ID); err != nil {
			return fmt.Errorf("failed to remove deleted_at from graph: %w", err)
		}

		if err := s.insertGraphSpkWithRelations(&updatedSpk); err != nil {
			return fmt.Errorf("failed to restore SPK graph: %w", err)
		}
	}
	return nil
}

func (s *spkService) insertGraphSpkWithRelations(data *models.Spk) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"name":        data.Name,
		"code":        data.Code,
		"description": helper.ToJSONString(data.Description),
		"id":          data.ID,
		"createdAt":   data.CreatedAt.Format(time.RFC3339Nano),
		"updatedAt":   data.UpdatedAt.Format(time.RFC3339Nano),
	}

	if err := graph.
		WithMerge("(s:SPK {id: $id})").
		WithSet("s.name = $name, s.code = $code, s.description = $description, s.created_at = datetime($createdAt), s.updated_at = datetime($updatedAt)", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to merge SPK node: %w", err)
	}

	if len(data.HasTitles) > 0 {
		fmt.Printf("Creating relationships for %d titles\n", len(data.HasTitles))
		titleIDs := make([]int64, 0, len(data.HasTitles))
		for _, title := range data.HasTitles {
			titleIDs = append(titleIDs, title.ID)
		}

		relParams := map[string]any{
			"spkId":    data.ID,
			"titleIds": titleIDs,
		}

		if err := graph.
			WithMatch("(s:SPK {id: $spkId})").
			WithUnwind("$titleIds", "titleId").
			WithMatch("(t:Title {id: titleId})").
			WithMerge("(s)-[:HAS_TITLE]->(t)").
			WithParams(relParams).
			RunWrite(); err != nil {
			fmt.Printf("Failed to create title relationship: %v\n", err)
			return fmt.Errorf("failed to create HAS_TITLE relationship for SPK %d: %w", data.ID, err)
		}
	}

	return nil
}

func (s *spkService) removeDeletedAtFromGraph(spkID int64) error {
	graph := builder.NewGraphRepository()
	params := map[string]interface{}{
		"spkId": spkID,
	}

	if err := graph.
		WithMatch("(s:SPK {id: $spkId})").
		WithCall(`
			apoc.path.expandConfig(s, {
				relationshipFilter: ">",
				minLevel: 0,
				maxLevel: -1
			})
		`).
		WithYield("path").
		WithWith("s, collect(DISTINCT path) AS paths").
		WithUnwind("paths", "p").
		WithUnwind("nodes(p)", "n").
		WithWith("DISTINCT n").
		WithSet("n.deleted_at = NULL", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to remove deleted_at from SPK graph with id %d: %w", spkID, err)
	}
	return nil
}

func GetSPKGraphs() (any, error) {
	graph := builder.NewGraphRepository()

	result, err := graph.
		WithMatch("(s:SPK)").
		WithWhere("s.deleted_at IS NULL", nil).
		WithOptionalMatch("(s)-[:HAS_TITLE]->(t:Title)").
		WithWhere("t.deleted_at IS NULL", nil).
		WithWith("s, collect(apoc.map.removeKey(apoc.convert.toMap(t), 'deleted_at')) AS titles").
		WithWith("apoc.map.removeKey(apoc.convert.toMap(s), 'deleted_at') AS sMap").
		WithWith("apoc.map.setKey(sMap, 'has_title', titles) AS spk").
		WithReturn("spk").
		RunWriteWithReturn()
	if err != nil {
		return nil, err
	}

	return helper.Neo4jFormatter(result), nil
}
