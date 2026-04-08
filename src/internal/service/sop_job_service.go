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

type SopJobService interface {
	WithTx(tx *gorm.DB) SopJobService

	CreateSopJob(input *models.SopJob) (*models.SopJob, error)
	UpdateSopJob(id int64, updates map[string]any) (*models.SopJob, error)
	DeleteSopJob(id int64, isPermanent bool) error
	GetAllSopJobs(filter dto.SopJobFilterDto) ([]models.SopJob, error)
	GetDB() *gorm.DB

	GetSopJobByID(id int64, filter dto.SopJobFilterDto) (*models.SopJob, error)
	GetSopJobByIDs(ids []int64) ([]*models.SopJob, error)
	BulkCreateSopJobs(input []*models.SopJob) ([]*models.SopJob, error)
	BulkUpdateSopJobs(ids []int64, updates map[string]any) error
	BulkDeleteSopJobs(ids []int64, isPermanent bool) error
	ReorderSopJob(sopJobID int64, newIndex int, sopID int64) error
	CountSopJobs(filter dto.SopJobFilterDto) (int64, error)

	GetAllGraphSopJobs(filter dto.SopJobFilterDto) ([]*graphdb.SopJobNode, error)
	GetGraphSopJobByID(id int64) (*graphdb.SopJobNode, error)
	InsertGraphSopJob(data *graphdb.SopJobNode) error
	UpdateGraphSopJob(data *graphdb.SopJobNode) error
	DeleteGraphSopJob(sopJobId int64) error
	BulkInsertGraphSopJobs(data []*graphdb.SopJobNode) error
	BulkUpdateGraphSopJobs(data []*graphdb.SopJobNode) error
	BulkDeleteGraphSopJobs(ids []int64) error
	CountGraphSopJobs(filter dto.SopJobFilterDto) (int64, error)
}

type sopJobService struct {
	repo      sql.SopJobRepository
	graphRepo graphdb.SopJobRepository
	tx        *gorm.DB
}

func NewSopJobService(repo sql.SopJobRepository, graphRepo graphdb.SopJobRepository) SopJobService {
	return &sopJobService{repo: repo, graphRepo: graphRepo}
}

func (s *sopJobService) WithTx(tx *gorm.DB) SopJobService {
	return &sopJobService{
		repo:      s.repo.WithTx(tx),
		graphRepo: s.graphRepo,
		tx:        tx,
	}
}

func (s *sopJobService) GetDB() *gorm.DB {
	if s.tx != nil {
		return s.tx
	}
	return config.DB
}

func (s *sopJobService) CreateSopJob(input *models.SopJob) (*models.SopJob, error) {
	data, err := s.repo.InsertSopJob(input)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	if err := s.GetDB().Preload("HasSop").First(&data, data.ID).Error; err != nil {
		return nil, err
	}

	if input.ReferenceID != nil && *input.Type == "sop" {
		if err := s.GetDB().Model(&models.Sop{}).
			Where("id = ? AND parent_job_id IS NULL", input.ReferenceID).
			Update("parent_job_id", data.ID).Error; err != nil {
			return nil, gorm_err.TranslateGormError(err)
		}
	}

	if err := s.InsertGraphSopJob(toSopJobNode(data)); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}

	return data, nil
}

func (s *sopJobService) UpdateSopJob(id int64, updates map[string]any) (*models.SopJob, error) {
	if _, err := s.repo.FindSopJobByID(id); err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	updates["updated_at"] = time.Now()

	data, err := s.repo.UpdateSopJob(id, updates)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	if err := s.UpdateGraphSopJob(toSopJobNode(data)); err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	if updates["reference_id"] != nil && updates["type"] == "sop" {
		if err := s.GetDB().Model(&models.Sop{}).
			Where("id = ? AND parent_job_id IS NULL", updates["reference_id"]).
			Update("parent_job_id", id).Error; err != nil {
			return nil, gorm_err.TranslateGormError(err)
		}
	}

	return data, nil
}

func (s *sopJobService) DeleteSopJob(id int64, isPermanent bool) (err error) {
	repo := s.repo
	if isPermanent {
		repo = repo.WithUnscoped()
		err = repo.RemoveSopJob(id)
		return gorm_err.TranslateGormError(err)
	}

	_, err = s.GetSopJobByID(id, dto.SopJobFilterDto{Preload: true})
	if err != nil {
		return err
	}
	if err := s.GetDB().Model(&models.Sop{}).
		Where("parent_job_id = ?", id).
		Update("parent_job_id", nil).Error; err != nil {
		return gorm_err.TranslateGormError(err)
	}
	err = repo.RemoveSopJob(id)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	if err := s.DeleteGraphSopJob(id); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	return nil
}

func (s *sopJobService) GetAllSopJobs(filter dto.SopJobFilterDto) ([]models.SopJob, error) {
	repo := s.repo

	if filter.SopID != 0 {
		repo = repo.WithWhere("sop_id = ?", filter.SopID)
	}
	if filter.TitleID != 0 {
		repo = repo.WithWhere("title_id = ?", filter.TitleID)
	}
	if filter.Preload {
		repo = repo.WithPreloads("HasSop", "HasTitle", "HasFlowchart")
	}
	if filter.Name != "" {
		repo = repo.WithWhere("name ILIKE ?", "%"+filter.Name+"%")
	}
	if filter.ShowDeleted {
		repo = repo.WithUnscoped().WithWhere("deleted_at IS NOT NULL")
	}

	data, err := repo.FindSopJob()
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	if filter.Preload {
		for i := range data {
			if err := s.loadDynamicReference(&data[i]); err != nil {
				return nil, err
			}
		}
	}

	return data, nil
}

func (s *sopJobService) GetSopJobByID(id int64, filter dto.SopJobFilterDto) (*models.SopJob, error) {
	repo := s.repo
	if filter.Preload {
		repo = repo.WithPreloads("HasSop", "HasTitle")
	}
	data, err := repo.FindSopJobByID(id)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	if filter.Preload {
		if err := s.loadDynamicReference(data); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (s *sopJobService) GetSopJobByIDs(ids []int64) ([]*models.SopJob, error) {
	data, err := s.repo.FindSopJobByIDs(ids)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *sopJobService) BulkCreateSopJobs(input []*models.SopJob) ([]*models.SopJob, error) {
	datas, err := s.repo.InsertManySopJobs(input)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	if err := s.BulkInsertGraphSopJobs(toSopJobNodeSlice(datas)); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}
	return datas, nil
}

func (s *sopJobService) BulkUpdateSopJobs(ids []int64, updates map[string]any) error {
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
			if err := s.sopJobRestore(ids); err != nil {
				return err
			}
			err := repo.UpdateManySopJobs(ids, updates)
			return gorm_err.TranslateGormError(err)
		}
	}

	updatedJobs, err := s.repo.FindSopJobByIDs(ids)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	if err := s.BulkUpdateGraphSopJobs(toSopJobNodeSlice(updatedJobs)); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	err = repo.UpdateManySopJobs(ids, updates)
	return gorm_err.TranslateGormError(err)
}

func (s *sopJobService) BulkDeleteSopJobs(ids []int64, isPermanent bool) (err error) {
	repo := s.repo

	if isPermanent {
		repo = repo.WithUnscoped()
		err = repo.RemoveManySopJobs(ids)
		return gorm_err.TranslateGormError(err)
	}

	if len(ids) > 0 {
		if err := s.GetDB().Model(&models.Sop{}).Where("parent_job_id IN ?", ids).Update("parent_job_id", nil).Error; err != nil {
			return gorm_err.TranslateGormError(err)
		}
	}

	if err := s.BulkDeleteGraphSopJobs(ids); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	err = repo.RemoveManySopJobs(ids)
	return gorm_err.TranslateGormError(err)
}

func (s *sopJobService) ReorderSopJob(sopJobID int64, newIndex int, sopID int64) error {
	db := s.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
		}
	}()

	repoWithTx := s.repo.WithTx(db)

	if err := repoWithTx.ReorderSopJob(sopJobID, newIndex, sopID); err != nil {
		db.Rollback()
		return gorm_err.TranslateGormError(err)
	}

	return db.Commit().Error
}

func (s *sopJobService) CountSopJobs(filter dto.SopJobFilterDto) (int64, error) {
	repo := s.repo
	if filter.SopID != 0 {
		repo = repo.WithWhere("sop_id = ?", filter.SopID)
	}

	data, err := repo.CountSopJobs()
	if err != nil {
		return 0, gorm_err.TranslateGormError(err)
	}

	return data, nil
}

func (s *sopJobService) GetAllGraphSopJobs(filter dto.SopJobFilterDto) ([]*graphdb.SopJobNode, error) {
	return s.graphRepo.GetAllGraphSopJobs(filter)
}

func (s *sopJobService) GetGraphSopJobByID(id int64) (*graphdb.SopJobNode, error) {
	return s.graphRepo.GetGraphSopJobByID(id)
}

func (s *sopJobService) InsertGraphSopJob(data *graphdb.SopJobNode) error {
	return s.graphRepo.InsertGraphSopJob(data)
}

func (s *sopJobService) UpdateGraphSopJob(data *graphdb.SopJobNode) error {
	return s.graphRepo.UpdateGraphSopJob(data)
}

func (s *sopJobService) DeleteGraphSopJob(sopJobId int64) error {
	return s.graphRepo.DeleteGraphSopJob(sopJobId)
}

func (s *sopJobService) BulkInsertGraphSopJobs(data []*graphdb.SopJobNode) error {
	return s.graphRepo.BulkInsertGraphSopJobs(data)
}

func (s *sopJobService) BulkUpdateGraphSopJobs(data []*graphdb.SopJobNode) error {
	return s.graphRepo.BulkUpdateGraphSopJobs(data)
}

func (s *sopJobService) BulkDeleteGraphSopJobs(ids []int64) error {
	return s.graphRepo.BulkDeleteGraphSopJobs(ids)
}

func (s *sopJobService) CountGraphSopJobs(filter dto.SopJobFilterDto) (int64, error) {
	return s.graphRepo.CountGraphSopJobs(filter)
}

func toSopJobNode(m *models.SopJob) *graphdb.SopJobNode {
	var typeStr string
	if m.Type != nil {
		typeStr = *m.Type
	}
	return &graphdb.SopJobNode{
		ID:          m.ID,
		Name:        m.Name,
		Alias:       m.Alias,
		Type:        typeStr,
		Code:        m.Code,
		Description: helper.ToJSONString(m.Description),
		TitleID:     m.TitleID,
		SopID:       m.SopID,
		ReferenceID: m.ReferenceID,
		Index:       m.Index,
		FlowchartID: m.FlowchartID,
		NextIndex:   m.NextIndex,
		PrevIndex:   m.PrevIndex,
		IsPublished: m.IsPublished,
		IsHide:      m.IsHide,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   m.UpdatedAt.Format(time.RFC3339Nano),
	}
}

func toSopJobNodeSlice(m []*models.SopJob) []*graphdb.SopJobNode {
	result := make([]*graphdb.SopJobNode, 0, len(m))
	for _, sopJob := range m {
		result = append(result, toSopJobNode(sopJob))
	}
	return result
}

func (s *sopJobService) loadDynamicReference(sopJob *models.SopJob) error {
	if sopJob.Type == nil || sopJob.ReferenceID == nil {
		return nil
	}

	db := s.GetDB()

	switch *sopJob.Type {
	case "sop":
		var sop models.Sop
		if err := db.Where("id = ?", *sopJob.ReferenceID).First(&sop).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return gorm_err.TranslateGormError(err)
			}
			return nil
		}
		sopJob.HasReference = &sop

	case "spk":
		var spk models.Spk
		if err := db.Where("id = ?", *sopJob.ReferenceID).First(&spk).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return gorm_err.TranslateGormError(err)
			}
			return nil
		}
		sopJob.HasReference = &spk
	}

	return nil
}

func (s *sopJobService) sopJobRestore(ids []int64) error {
	var sopJobs []models.SopJob
	if err := s.GetDB().Unscoped().Preload("HasSop").Where("id IN ? AND deleted_at IS NOT NULL", ids).Find(&sopJobs).Error; err != nil {
		return err
	}

	for _, sopJob := range sopJobs {
		var updatedSopJob models.SopJob
		if err := s.GetDB().Unscoped().Where("id = ?", sopJob.ID).First(&updatedSopJob).Error; err != nil {
			return err
		}

		if updatedSopJob.CreatedAt.IsZero() {
			updatedSopJob.CreatedAt = time.Now()
		}
		updatedSopJob.UpdatedAt = time.Now()

		if err := s.removeDeletedAtFromGraph(updatedSopJob.ID); err != nil {
			return fmt.Errorf("failed to remove deleted_at from graph: %w", err)
		}

		if err := s.insertGraphSopJobWithRelations(&updatedSopJob); err != nil {
			return fmt.Errorf("failed to restore SOP Job graph: %w", err)
		}
	}
	return nil
}

func (s *sopJobService) insertGraphSopJobWithRelations(data *models.SopJob) error {
	graph := builder.NewGraphRepository()

	sopName := ""
	if data.HasSop != nil {
		sopName = data.HasSop.Name
	}

	params := map[string]any{
		"id":          data.ID,
		"name":        data.Name,
		"alias":       data.Alias,
		"type":        data.Type,
		"code":        data.Code,
		"description": helper.ToJSONString(data.Description),
		"titleId":     data.TitleID,
		"sopId":       data.SopID,
		"sopName":     sopName,
		"createdAt":   data.CreatedAt.Format(time.RFC3339Nano),
		"updatedAt":   data.UpdatedAt.Format(time.RFC3339Nano),
	}

	if err := graph.
		WithMerge("(j:Job {id: $id})").
		WithSet("j.name = $name, j.alias = $alias, j.type = $type, j.code = $code, j.description = $description, j.created_at = datetime($createdAt), j.updated_at = datetime($updatedAt)", nil).
		WithMerge("(s:SOP {id: $sopId})").
		WithSet("s.name = $sopName", nil).
		WithMerge("(s)-[:HAS_JOB]->(j)").
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to merge Job node: %w", err)
	}

	return nil
}

func (s *sopJobService) removeDeletedAtFromGraph(sopJobID int64) error {
	graph := builder.NewGraphRepository()
	params := map[string]interface{}{
		"jobId": sopJobID,
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
		return fmt.Errorf("failed to remove deleted_at from Job graph with id %d: %w", sopJobID, err)
	}
	return nil
}

func GetSOPJobGraphs() (any, error) {
	graph := builder.NewGraphRepository()

	result, err := graph.
		WithMatch("(j:Job)").
		WithWhere("j.deleted_at IS NULL", nil).
		WithOptionalMatch("(s:SOP)-[:HAS_JOB]->(j)").
		WithWhere("s.deleted_at IS NULL", nil).
		WithWith("j, collect(apoc.map.removeKey(apoc.convert.toMap(s), 'deleted_at')) AS sops").
		WithWith("apoc.map.removeKey(apoc.convert.toMap(j), 'deleted_at') AS jMap").
		WithWith("apoc.map.setKey(jMap, 'has_sop', sops) AS job").
		WithReturn("job").
		RunWriteWithReturn()
	if err != nil {
		return nil, err
	}

	return helper.Neo4jFormatter(result), nil
}
