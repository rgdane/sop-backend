package service

import (
	"fmt"
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

type SpkJobService interface {
	WithTx(tx *gorm.DB) SpkJobService

	CreateSpkJob(input *models.SpkJob) (*models.SpkJob, error)
	UpdateSpkJob(id int64, updates map[string]interface{}) (*models.SpkJob, error)
	DeleteSpkJob(id int64, isPermanent bool) error
	GetAllSpkJobs(filter dto.SpkJobFilterDto) ([]models.SpkJob, error)
	GetSpkJobByID(id int64, filter dto.SpkJobFilterDto) (*models.SpkJob, error)
	GetSpkJobsByIDs(ids []int64) ([]*models.SpkJob, error)
	GetDB() *gorm.DB
	BulkCreateSpkJobs(data []*models.SpkJob) ([]*models.SpkJob, error)
	BulkUpdateSpkJobs(ids []int64, updates map[string]interface{}) error
	BulkDeleteSpkJobs(ids []int64, isPermanent bool) error
	ReorderSpkJob(spkJobID int64, newIndex int, spkID int64) error
	CountSpkJobs(filter dto.SpkJobFilterDto) (int64, error)

	GetAllGraphSpkJobs(filter dto.SpkJobFilterDto) ([]*graphdb.SpkJobNode, error)
	GetGraphSpkJobByID(id int64) (*graphdb.SpkJobNode, error)
	InsertGraphSpkJob(data *graphdb.SpkJobNode) error
	UpdateGraphSpkJob(data *graphdb.SpkJobNode) error
	DeleteGraphSpkJob(spkJobId int64) error
	BulkInsertGraphSpkJobs(data []*graphdb.SpkJobNode) error
	BulkUpdateGraphSpkJobs(data []*graphdb.SpkJobNode) error
	BulkDeleteGraphSpkJobs(ids []int64) error
	CountGraphSpkJobs(filter dto.SpkJobFilterDto) (int64, error)
}

type spkJobService struct {
	repo      sql.SpkJobRepository
	graphRepo graphdb.SpkJobRepository
	tx        *gorm.DB
}

func NewSpkJobService(repo sql.SpkJobRepository, graphRepo graphdb.SpkJobRepository) SpkJobService {
	return &spkJobService{repo: repo, graphRepo: graphRepo}
}

func (s *spkJobService) WithTx(tx *gorm.DB) SpkJobService {
	return &spkJobService{
		repo:      s.repo.WithTx(tx),
		graphRepo: s.graphRepo,
		tx:        tx,
	}
}

func (s *spkJobService) GetDB() *gorm.DB {
	if s.tx != nil {
		return s.tx
	}
	return config.DB
}

func (s *spkJobService) CreateSpkJob(input *models.SpkJob) (*models.SpkJob, error) {
	data, err := s.repo.InsertSpkJob(input)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	if err := s.InsertGraphSpkJob(toSpkJobNode(data)); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}

	return data, nil
}

func (s *spkJobService) UpdateSpkJob(id int64, updates map[string]interface{}) (*models.SpkJob, error) {
	if _, err := s.repo.FindSpkJobByID(id); err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	updates["updated_at"] = time.Now()

	data, err := s.repo.UpdateSpkJob(id, updates)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	if err := s.UpdateGraphSpkJob(toSpkJobNode(data)); err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	return data, nil
}

func (s *spkJobService) DeleteSpkJob(id int64, isPermanent bool) (err error) {
	repo := s.repo
	if isPermanent {
		repo = repo.WithUnscoped()
		err = repo.RemoveSpkJob(id)
		return gorm_err.TranslateGormError(err)
	}

	_, err = s.repo.FindSpkJobByID(id)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	if err := s.DeleteGraphSpkJob(id); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	err = repo.RemoveSpkJob(id)
	return gorm_err.TranslateGormError(err)
}

func (s *spkJobService) GetAllSpkJobs(filter dto.SpkJobFilterDto) ([]models.SpkJob, error) {
	repo := s.repo

	if filter.Preload {
		repo = repo.WithPreloads("HasTitle", "HasSop", "HasFlowchart", "HasSpk")
	}
	if filter.TitleID != 0 {
		repo = repo.WithWhere("title_id = ?", filter.TitleID)
	}
	if filter.SpkID != 0 {
		repo = repo.WithWhere("spk_id = ?", filter.SpkID)
	}
	if filter.SopID != 0 {
		repo = repo.WithWhere("sop_id = ?", filter.SopID)
	}
	if filter.Name != "" {
		repo = repo.WithWhere("name ILIKE ?", "%"+filter.Name+"%")
	}
	if filter.ShowDeleted {
		repo = repo.WithUnscoped().WithWhere("deleted_at IS NOT NULL")
	}

	data, err := repo.FindSpkJobs()
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *spkJobService) GetSpkJobByID(id int64, filter dto.SpkJobFilterDto) (*models.SpkJob, error) {
	repo := s.repo
	if filter.Preload {
		repo = repo.WithPreloads("HasTitle", "HasSop")
	}
	data, err := repo.FindSpkJobByID(id)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *spkJobService) GetSpkJobsByIDs(ids []int64) ([]*models.SpkJob, error) {
	data, err := s.repo.FindSpkJobsByIDs(ids)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *spkJobService) BulkCreateSpkJobs(input []*models.SpkJob) ([]*models.SpkJob, error) {
	datas, err := s.repo.InsertManySpkJobs(input)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	if err := s.BulkInsertGraphSpkJobs(toSpkJobNodeSlice(datas)); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}
	return datas, nil
}

func (s *spkJobService) BulkUpdateSpkJobs(ids []int64, updates map[string]interface{}) error {
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
			if err := s.spkJobRestore(ids); err != nil {
				return err
			}
			err := repo.UpdateManySpkJobs(ids, updates)
			return gorm_err.TranslateGormError(err)
		}
	}

	updatedJobs, err := s.repo.FindSpkJobsByIDs(ids)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	if err := s.BulkUpdateGraphSpkJobs(toSpkJobNodeSlice(updatedJobs)); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	err = repo.UpdateManySpkJobs(ids, updates)
	return gorm_err.TranslateGormError(err)
}

func (s *spkJobService) BulkDeleteSpkJobs(ids []int64, isPermanent bool) (err error) {
	repo := s.repo

	if isPermanent {
		repo = repo.WithUnscoped()
		err = repo.RemoveManySpkJobs(ids)
		return gorm_err.TranslateGormError(err)
	}

	if err := s.BulkDeleteGraphSpkJobs(ids); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	err = repo.RemoveManySpkJobs(ids)
	return gorm_err.TranslateGormError(err)
}

func (s *spkJobService) ReorderSpkJob(spkJobID int64, newIndex int, spkID int64) error {
	db := s.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
		}
	}()
	repoWithTx := s.repo.WithTx(db)

	if err := repoWithTx.ReorderSpkJob(spkJobID, newIndex, spkID); err != nil {
		db.Rollback()
		return gorm_err.TranslateGormError(err)
	}

	return db.Commit().Error
}

func (s *spkJobService) CountSpkJobs(filter dto.SpkJobFilterDto) (int64, error) {
	repo := s.repo
	if filter.SpkID != 0 {
		repo = repo.WithWhere("spk_id = ?", filter.SpkID)
	}

	data, err := repo.CountSpkJobs()
	if err != nil {
		return 0, gorm_err.TranslateGormError(err)
	}

	return data, nil
}

func (s *spkJobService) GetAllGraphSpkJobs(filter dto.SpkJobFilterDto) ([]*graphdb.SpkJobNode, error) {
	return s.graphRepo.GetAllGraphSpkJobs(filter)
}

func (s *spkJobService) GetGraphSpkJobByID(id int64) (*graphdb.SpkJobNode, error) {
	return s.graphRepo.GetGraphSpkJobByID(id)
}

func (s *spkJobService) InsertGraphSpkJob(data *graphdb.SpkJobNode) error {
	return s.graphRepo.InsertGraphSpkJob(data)
}

func (s *spkJobService) UpdateGraphSpkJob(data *graphdb.SpkJobNode) error {
	return s.graphRepo.UpdateGraphSpkJob(data)
}

func (s *spkJobService) DeleteGraphSpkJob(spkJobId int64) error {
	return s.graphRepo.DeleteGraphSpkJob(spkJobId)
}

func (s *spkJobService) BulkInsertGraphSpkJobs(data []*graphdb.SpkJobNode) error {
	return s.graphRepo.BulkInsertGraphSpkJobs(data)
}

func (s *spkJobService) BulkUpdateGraphSpkJobs(data []*graphdb.SpkJobNode) error {
	return s.graphRepo.BulkUpdateGraphSpkJobs(data)
}

func (s *spkJobService) BulkDeleteGraphSpkJobs(ids []int64) error {
	return s.graphRepo.BulkDeleteGraphSpkJobs(ids)
}

func (s *spkJobService) CountGraphSpkJobs(filter dto.SpkJobFilterDto) (int64, error) {
	return s.graphRepo.CountGraphSpkJobs(filter)
}

func toSpkJobNode(m *models.SpkJob) *graphdb.SpkJobNode {
	return &graphdb.SpkJobNode{
		ID:          m.ID,
		Name:        m.Name,
		Description: helper.ToJSONString(m.Description),
		SpkID:       m.SpkID,
		SopID:       m.SopID,
		TitleID:     m.TitleID,
		Index:       m.Index,
		FlowchartID: m.FlowchartID,
		NextIndex:   m.NextIndex,
		PrevIndex:   m.PrevIndex,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   m.UpdatedAt.Format(time.RFC3339Nano),
	}
}

func toSpkJobNodeSlice(m []*models.SpkJob) []*graphdb.SpkJobNode {
	result := make([]*graphdb.SpkJobNode, 0, len(m))
	for _, spkJob := range m {
		result = append(result, toSpkJobNode(spkJob))
	}
	return result
}

func (s *spkJobService) spkJobRestore(ids []int64) error {
	var spkJobs []models.SpkJob
	if err := s.GetDB().Unscoped().Where("id IN ? AND deleted_at IS NOT NULL", ids).Find(&spkJobs).Error; err != nil {
		return err
	}

	for _, spkJob := range spkJobs {
		var updatedSpkJob models.SpkJob
		if err := s.GetDB().Unscoped().Where("id = ?", spkJob.ID).First(&updatedSpkJob).Error; err != nil {
			return err
		}

		if updatedSpkJob.CreatedAt.IsZero() {
			updatedSpkJob.CreatedAt = time.Now()
		}
		updatedSpkJob.UpdatedAt = time.Now()

		if err := s.removeDeletedAtFromGraph(updatedSpkJob.ID); err != nil {
			return fmt.Errorf("failed to remove deleted_at from graph: %w", err)
		}

		if err := s.insertGraphSpkJobWithRelations(&updatedSpkJob); err != nil {
			return fmt.Errorf("failed to restore SPK Job graph: %w", err)
		}
	}
	return nil
}

func (s *spkJobService) insertGraphSpkJobWithRelations(data *models.SpkJob) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"id":          data.ID,
		"name":        data.Name,
		"description": helper.ToJSONString(data.Description),
		"spkId":       data.SpkID,
		"sopId":       data.SopID,
		"titleId":     data.TitleID,
		"index":       data.Index,
		"createdAt":   data.CreatedAt.Format(time.RFC3339Nano),
		"updatedAt":   data.UpdatedAt.Format(time.RFC3339Nano),
	}

	if err := graph.
		WithMerge("(j:Job {id: $id})").
		WithSet("j.name = $name, j.description = $description, j.created_at = datetime($createdAt), j.updated_at = datetime($updatedAt)", nil).
		WithMerge("(s:SPK {id: $spkId})").
		WithMerge("(s)-[:HAS_JOB]->(j)").
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to merge Job node: %w", err)
	}

	return nil
}

func (s *spkJobService) removeDeletedAtFromGraph(spkJobID int64) error {
	graph := builder.NewGraphRepository()
	params := map[string]interface{}{
		"jobId": spkJobID,
	}

	if err := graph.
		WithMatch("(j:Job {id: $jobId})").
		WithCall(`
			apoc.path.expandConfig(j, {
				relationshipFilter: ">",
				minLevel: 0,
				maxLevel: -1
			})
		`).
		WithYield("path").
		WithWith("j, collect(DISTINCT path) AS paths").
		WithUnwind("paths", "p").
		WithUnwind("nodes(p)", "n").
		WithWith("DISTINCT n").
		WithSet("n.deleted_at = NULL", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to remove deleted_at from Job graph with id %d: %w", spkJobID, err)
	}
	return nil
}

func GetSPKJobGraphs() (any, error) {
	graph := builder.NewGraphRepository()

	result, err := graph.
		WithMatch("(j:Job)").
		WithWhere("j.deleted_at IS NULL", nil).
		WithOptionalMatch("(s:SPK)-[:HAS_JOB]->(j)").
		WithWhere("s.deleted_at IS NULL", nil).
		WithWith("j, collect(apoc.map.removeKey(apoc.convert.toMap(s), 'deleted_at')) AS spks").
		WithWith("apoc.map.removeKey(apoc.convert.toMap(j), 'deleted_at') AS jMap").
		WithWith("apoc.map.setKey(jMap, 'has_spk', spks) AS job").
		WithReturn("job").
		RunWriteWithReturn()
	if err != nil {
		return nil, err
	}

	return helper.Neo4jFormatter(result), nil
}
