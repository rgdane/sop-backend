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

type SpkJobService interface {
	WithTx(tx *gorm.DB) SpkJobService

	CreateSpkJob(input *models.SpkJob) (*models.SpkJob, error)
	UpdateSpkJob(id int64, updates map[string]interface{}) (*models.SpkJob, error)
	DeleteSpkJob(id int64) error
	GetAllSpkJobs(filter dto.SpkJobFilterDto) ([]models.SpkJob, error)
	GetSpkJobByID(id int64, filter dto.SpkJobFilterDto) (*models.SpkJob, error)
	GetSpkJobsByIDs(ids []int64) ([]*models.SpkJob, error)
	GetDB() *gorm.DB
	BulkCreateSpkJobs(data []*models.SpkJob) ([]*models.SpkJob, error)
	BulkUpdateSpkJobs(ids []int64, updates map[string]interface{}) error
	BulkDeleteSpkJobs(ids []int64) error
	ReorderSpkJob(spkJobID int64, newIndex int, spkID int64) error
}

type spkJobService struct {
	repo sql.SpkJobRepository
	tx   *gorm.DB
}

func NewSpkJobService(repo sql.SpkJobRepository) SpkJobService {
	return &spkJobService{repo: repo}
}

func (s *spkJobService) WithTx(tx *gorm.DB) SpkJobService {
	return &spkJobService{
		repo: s.repo.WithTx(tx),
		tx:   tx,
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

	// Sync to Neo4j - create Job node and relate to SPK Document
	if err := s.insertGraphSpkJob(data); err != nil {
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

	// Sync to Neo4j after SQL update
	if err := s.updateGraphSpkJob(data); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}

	return data, nil
}

func (s *spkJobService) DeleteSpkJob(id int64) error {
	// Get SPK Job data before deletion for Neo4j cleanup
	spkJob, err := s.repo.FindSpkJobByID(id)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	// Delete from Neo4j graph first
	if err := s.deleteGraphSpkJob(spkJob); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	// Then delete from SQL
	err = s.repo.RemoveSpkJob(id)
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

func (s *spkJobService) BulkUpdateSpkJobs(ids []int64, updates map[string]interface{}) error {
	err := s.repo.UpdateManySpkJobs(ids, updates)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	// Sync to Neo4j using batch - reload SPK Jobs and update graph
	spkJobs, err := s.repo.FindSpkJobsByIDs(ids)
	if err != nil {
		fmt.Printf("Failed to load SPK Jobs for graph sync: %v\n", err)
		return nil
	}

	if err := s.batchUpdateGraphSpkJobs(spkJobs); err != nil {
		fmt.Printf("Failed to batch update SPK Jobs in graph: %v\n", err)
	}

	return nil
}

func (s *spkJobService) BulkCreateSpkJobs(input []*models.SpkJob) ([]*models.SpkJob, error) {
	datas, err := s.repo.InsertManySpkJobs(input)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	// Sync to Neo4j using batch
	if err := s.batchInsertGraphSpkJobs(datas); err != nil {
		fmt.Printf("Failed to batch sync SPK Jobs to graph: %v\n", err)
	}

	return datas, nil
}

func (s *spkJobService) GetSpkJobsByIDs(ids []int64) ([]*models.SpkJob, error) {
	data, err := s.repo.FindSpkJobsByIDs(ids)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *spkJobService) BulkDeleteSpkJobs(ids []int64) error {
	// Delete from Neo4j first using batch
	if err := s.batchDeleteGraphSpkJobs(ids); err != nil {
		fmt.Printf("Failed to batch delete SPK Jobs from graph: %v\n", err)
	}

	// Then delete from SQL
	err := s.repo.RemoveManySpkJobs(ids)
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

// insertGraphSpkJob creates Job node in Neo4j and relates to SPK node
func (s *spkJobService) insertGraphSpkJob(data *models.SpkJob) error {
	graph := builder.NewGraphRepository()

	// Create Job node and relate to parent SPK node with properties from SQL
	// Convert description JSON to string for Neo4j
	jobParam := map[string]interface{}{
		"spkId":       data.SpkID,
		"id":          data.ID,
		"name":        data.Name,
		"description": helper.ToJSONString(data.Description),
		"index":       data.Index,
		"sopId":       data.SopID,
		"titleId":     data.TitleID,
		"flowchartId": data.FlowchartID,
		"nextIndex":   data.NextIndex,
		"prevIndex":   data.PrevIndex,
	}

	// Match parent SPK node and create Job node with HAS_JOB relation
	if err := graph.
		WithMatch("(s:SPK {id: $spkId})").
		WithMerge("(s)-[:HAS_JOB]->(j:Job {id: $id})").
		WithSet(`j.name = $name, 
			j.description = $description, 
			j.index = $index,
			j.sopId = $sopId,
			j.titleId = $titleId,
			j.flowchartId = $flowchartId,
			j.nextIndex = $nextIndex,
			j.prevIndex = $prevIndex`, jobParam).
		WithParams(jobParam).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to create SPK Job node: %w", err)
	}

	// If sopId is not null, create HAS_REFERENCE relation to SOP
	if data.SopID != nil && *data.SopID != 0 {
		refParam := map[string]interface{}{
			"jobId": data.ID,
			"sopId": *data.SopID,
		}

		if err := graph.
			WithMatch("(j:Job {id: $jobId})").
			WithMatch("(sop:SOP {id: $sopId})").
			WithMerge("(j)-[:HAS_REFERENCE]->(sop)").
			WithParams(refParam).
			RunWrite(); err != nil {
			return fmt.Errorf("failed to create HAS_REFERENCE relation to SOP: %w", err)
		}
	}

	return nil
}

// updateGraphSpkJob updates Job node in Neo4j
func (s *spkJobService) updateGraphSpkJob(data *models.SpkJob) error {
	graph := builder.NewGraphRepository()

	// Update Job node properties in Neo4j
	jobParam := map[string]interface{}{
		"id":          data.ID,
		"name":        data.Name,
		"description": helper.ToJSONString(data.Description),
		"index":       data.Index,
		"sopId":       data.SopID,
		"titleId":     data.TitleID,
		"flowchartId": data.FlowchartID,
		"nextIndex":   data.NextIndex,
		"prevIndex":   data.PrevIndex,
	}

	if err := graph.
		WithMatch("(j:Job {id: $id})").
		WithSet(`j.name = $name, 
			j.description = $description, 
			j.index = $index,
			j.sopId = $sopId,
			j.titleId = $titleId,
			j.flowchartId = $flowchartId,
			j.nextIndex = $nextIndex,
			j.prevIndex = $prevIndex`, jobParam).
		WithParams(jobParam).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update SPK Job node: %w", err)
	}

	// Delete existing HAS_REFERENCE relation first
	deleteRefParam := map[string]interface{}{
		"jobId": data.ID,
	}

	if err := graph.
		WithMatch("(j:Job {id: $jobId})-[r:HAS_REFERENCE]->()").
		WithDelete("r").
		WithParams(deleteRefParam).
		RunWrite(); err != nil {
		// Ignore error if no relation exists
	}

	// If sopId is not null, create HAS_REFERENCE relation to SOP
	if data.SopID != nil && *data.SopID != 0 {
		refParam := map[string]interface{}{
			"jobId": data.ID,
			"sopId": *data.SopID,
		}

		if err := graph.
			WithMatch("(j:Job {id: $jobId})").
			WithMatch("(sop:SOP {id: $sopId})").
			WithMerge("(j)-[:HAS_REFERENCE]->(sop)").
			WithParams(refParam).
			RunWrite(); err != nil {
			return fmt.Errorf("failed to create HAS_REFERENCE relation to SOP: %w", err)
		}
	}

	return nil
}

// deleteGraphSpkJob removes Job node from Neo4j
func (s *spkJobService) deleteGraphSpkJob(data *models.SpkJob) error {
	return s.deleteGraphSpkJobByID(data.ID)
}

func (s *spkJobService) deleteGraphSpkJobByID(jobId int64) error {
	graph := builder.NewGraphRepository()

	// Delete Job node and all its children recursively
	params := map[string]interface{}{
		"jobId": jobId,
	}

	if err := graph.
		WithMatch("(j:Job {id: $jobId})").
		WithOptionalMatch("(j)-[*]->(child)").
		WithDetachDelete("j, child").
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to delete SPK Job graph with id %d: %w", jobId, err)
	}

	return nil
}

// batchInsertGraphSpkJobs creates multiple SPK Job nodes in Neo4j using UNWIND batch operation
func (s *spkJobService) batchInsertGraphSpkJobs(data []*models.SpkJob) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	// Prepare batch data for Job nodes with all properties
	var jobList []map[string]interface{}
	for _, job := range data {
		jobList = append(jobList, map[string]interface{}{
			"id":          job.ID,
			"name":        job.Name,
			"description": helper.ToJSONString(job.Description),
			"spkId":       job.SpkID,
			"index":       job.Index,
			"sopId":       job.SopID,
			"titleId":     job.TitleID,
			"flowchartId": job.FlowchartID,
			"nextIndex":   job.NextIndex,
			"prevIndex":   job.PrevIndex,
		})
	}

	params := map[string]interface{}{
		"jobs": jobList,
	}

	// Use UNWIND to batch match SPK and create Job nodes with HAS_JOB relationship
	// This matches the pattern: (s:SPK)-[:HAS_JOB]->(j:Job)
	if err := graph.
		WithUnwind("$jobs", "jobData").
		WithMatch("(s:SPK {id: jobData.spkId})").
		WithMerge("(s)-[:HAS_JOB]->(j:Job {id: jobData.id})").
		WithSet(`j.name = jobData.name, 
			j.description = jobData.description, 
			j.index = jobData.index,
			j.sopId = jobData.sopId,
			j.titleId = jobData.titleId,
			j.flowchartId = jobData.flowchartId,
			j.nextIndex = jobData.nextIndex,
			j.prevIndex = jobData.prevIndex`, params).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to batch insert SPK Job nodes with HAS_JOB relationship: %w", err)
	}

	// Create HAS_REFERENCE relationships for jobs with SOP reference (sop_id)
	var sopRefList []map[string]interface{}
	for _, job := range data {
		if job.SopID != nil && *job.SopID != 0 {
			sopRefList = append(sopRefList, map[string]interface{}{
				"jobId": job.ID,
				"sopId": *job.SopID,
			})
		}
	}

	if len(sopRefList) > 0 {
		sopRefParams := map[string]interface{}{
			"refs": sopRefList,
		}

		// Create HAS_REFERENCE to SOP
		if err := graph.
			WithUnwind("$refs", "refData").
			WithMatch("(j:Job {id: refData.jobId}), (sop:SOP {id: refData.sopId})").
			WithMerge("(j)-[:HAS_REFERENCE]->(sop)").
			WithParams(sopRefParams).
			RunWrite(); err != nil {
			fmt.Printf("Warning: failed to batch create HAS_REFERENCE to SOP: %v\n", err)
		}
	}

	return nil
}

// batchUpdateGraphSpkJobs updates multiple SPK Job nodes in Neo4j using UNWIND batch operation
func (s *spkJobService) batchUpdateGraphSpkJobs(data []*models.SpkJob) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	// Prepare batch data with all properties
	var jobList []map[string]interface{}
	var jobIds []int64
	for _, job := range data {
		jobList = append(jobList, map[string]interface{}{
			"id":          job.ID,
			"name":        job.Name,
			"description": helper.ToJSONString(job.Description),
			"index":       job.Index,
			"sopId":       job.SopID,
			"titleId":     job.TitleID,
			"flowchartId": job.FlowchartID,
			"nextIndex":   job.NextIndex,
			"prevIndex":   job.PrevIndex,
		})
		jobIds = append(jobIds, job.ID)
	}

	params := map[string]interface{}{
		"jobs": jobList,
	}

	// Use UNWIND to batch update Job nodes with all properties
	if err := graph.
		WithUnwind("$jobs", "jobData").
		WithMatch("(j:Job {id: jobData.id})").
		WithSet(`j.name = jobData.name, 
			j.description = jobData.description, 
			j.index = jobData.index,
			j.sopId = jobData.sopId,
			j.titleId = jobData.titleId,
			j.flowchartId = jobData.flowchartId,
			j.nextIndex = jobData.nextIndex,
			j.prevIndex = jobData.prevIndex`, params).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to batch update SPK Job nodes: %w", err)
	}

	// Delete existing HAS_REFERENCE relations for all jobs in batch
	deleteParams := map[string]interface{}{
		"jobIds": jobIds,
	}

	if err := graph.
		WithUnwind("$jobIds", "jobId").
		WithMatch("(j:Job {id: jobId})-[r:HAS_REFERENCE]->()").
		WithDelete("r").
		WithParams(deleteParams).
		RunWrite(); err != nil {
		// Ignore error if no relations exist
	}

	// Recreate HAS_REFERENCE relationships for jobs with SOP reference
	var sopRefList []map[string]interface{}
	for _, job := range data {
		if job.SopID != nil && *job.SopID != 0 {
			sopRefList = append(sopRefList, map[string]interface{}{
				"jobId": job.ID,
				"sopId": *job.SopID,
			})
		}
	}

	if len(sopRefList) > 0 {
		sopRefParams := map[string]interface{}{
			"refs": sopRefList,
		}

		// Create HAS_REFERENCE to SOP in batch
		if err := graph.
			WithUnwind("$refs", "refData").
			WithMatch("(j:Job {id: refData.jobId}), (sop:SOP {id: refData.sopId})").
			WithMerge("(j)-[:HAS_REFERENCE]->(sop)").
			WithParams(sopRefParams).
			RunWrite(); err != nil {
			fmt.Printf("Warning: failed to batch create HAS_REFERENCE to SOP: %v\n", err)
		}
	}

	return nil
}

// batchDeleteGraphSpkJobs deletes multiple SPK Job nodes from Neo4j using UNWIND batch operation
func (s *spkJobService) batchDeleteGraphSpkJobs(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	params := map[string]interface{}{
		"jobIds": ids,
	}

	// Use UNWIND to batch delete Job nodes and their children
	if err := graph.
		WithUnwind("$jobIds", "jobId").
		WithMatch("(j:Job {id: jobId})").
		WithOptionalMatch("(j)-[*]->(child)").
		WithDetachDelete("j, child").
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to batch delete SPK Job nodes: %w", err)
	}

	return nil
}
