package service

import (
	"fmt"
	"time"

	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/internal/repository/sql"
	"jk-api/pkg/errors/gorm_err"
	"jk-api/pkg/neo4j/builder"

	"gorm.io/gorm"
)

type SopJobService interface {
	WithTx(tx *gorm.DB) SopJobService

	CreateSopJob(input *models.SopJob) (*models.SopJob, error)
	UpdateSopJob(id int64, updates map[string]any) (*models.SopJob, error)
	DeleteSopJob(id int64) error
	GetAllSopJobs(filter dto.SopJobFilterDto) ([]models.SopJob, error)
	GetDB() *gorm.DB

	GetSopJobByID(id int64, filter dto.SopJobFilterDto) (*models.SopJob, error)
	GetSopJobByIDs(ids []int64) ([]*models.SopJob, error)
	BulkCreateSopJobs(input []*models.SopJob) ([]*models.SopJob, error)
	BulkUpdateSopJobs(ids []int64, updates map[string]any) error
	BulkDeleteSopJobs(ids []int64) error
	ReorderSopJob(sopJobID int64, newIndex int, sopID int64) error
}

type sopJobService struct {
	repo sql.SopJobRepository
	tx   *gorm.DB
}

func NewSopJobService(repo sql.SopJobRepository) SopJobService {
	return &sopJobService{repo: repo}
}

func (s *sopJobService) WithTx(tx *gorm.DB) SopJobService {
	return &sopJobService{
		repo: s.repo.WithTx(tx),
		tx:   tx,
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

	if err := s.insertGraphSopJob(data, input); err != nil {
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

	err = s.updateGraphSopJob(data)
	if err != nil {
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

func (s *sopJobService) DeleteSopJob(id int64) error {
	_, err := s.GetSopJobByID(id, dto.SopJobFilterDto{Preload: true})
	if err != nil {
		return err
	}
	if err := s.GetDB().Model(&models.Sop{}).
		Where("parent_job_id = ?", id).
		Update("parent_job_id", nil).Error; err != nil {
		return gorm_err.TranslateGormError(err)
	}
	err = s.repo.RemoveSopJob(id)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}
	err = s.deleteGraphSopJob(id)

	return gorm_err.TranslateGormError(err)
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

	data, err := repo.FindSopJob()
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	// Load dynamic references if preload is enabled
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

	// Load dynamic reference if preload is enabled
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
	if err := s.bulkInsertGraphSopJobs(datas); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}
	return datas, nil
}

func (s *sopJobService) BulkUpdateSopJobs(ids []int64, updates map[string]any) error {
	err := s.repo.UpdateManySopJobs(ids, updates)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	// Fetch updated data untuk sync ke Neo4j
	updatedJobs, err := s.repo.FindSopJobByIDs(ids)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	// ✅ BULK UPDATE - SEKALI JALAN!
	if err := s.bulkUpdateGraphSopJobs(updatedJobs); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	return nil
}

func (s *sopJobService) BulkDeleteSopJobs(ids []int64) error {
	db := s.GetDB()
	if len(ids) > 0 {
		if err := db.Model(&models.Sop{}).Where("parent_job_id IN ?", ids).Update("parent_job_id", nil).Error; err != nil {
			return gorm_err.TranslateGormError(err)
		}
	}
	err := s.repo.RemoveManySopJobs(ids)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	// ✅ BULK DELETE - SEKALI JALAN!
	if err := s.bulkDeleteGraphSopJobs(ids); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	return nil
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

func (s *sopJobService) insertGraphSopJob(data *models.SopJob, input *models.SopJob) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"jobId":       data.ID,
		"jobName":     data.Name,
		"alias":       data.Alias,
		"type":        data.Type,
		"code":        data.Code,
		"description": data.Description,
		"titleId":     data.TitleID,
		"sopId":       data.SopID,
		"referenceId": data.ReferenceID,
		"index":       data.Index,
		"flowchartId": data.FlowchartID,
		"nextIndex":   data.NextIndex,
		"prevIndex":   data.PrevIndex,
		"sopName":     data.HasSop.Name,
		"isPublished": data.IsPublished,
		"isHide":      data.IsHide,
	}

	// 🔹 Merge Job node
	if err := graph.
		WithMerge("(j:Job {id: $jobId})").
		WithSet(`j.name = $jobName, j.alias = $alias, j.type = $type, 
		j.code = $code, j.description = $description, 
		j.title_id = $titleId, j.sop_id = $sopId, 
		j.reference_id = $referenceId, j.index = $index, 
		j.flowchart_id = $flowchartId, j.is_hide = $isHide,
		j.next_index = $nextIndex, j.prev_index = $prevIndex, j.is_published = $isPublished`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to merge Job node: %w", err)
	}

	// 🔹 Merge SOP node
	if err := graph.
		WithMerge("(s:SOP {id: $sopId})").
		WithSet("s.name = $sopName", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to merge SOP node: %w", err)
	}

	// 🔹 Relationship (SOP)-[:HAS_JOB]->(Job)
	if err := graph.
		WithMatch("(j:Job {id: $jobId})").
		WithMatch("(s:SOP {id: $sopId})").
		WithRelate("s", "HAS_JOB", "j", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to create HAS_JOB relationship: %w", err)
	}

	// 🔹 Optional Relationship (Job)-[:HAS_SOP]->(Parent SOP)
	if input.ReferenceID != nil && *input.Type == "sop" {
		var parentSop models.Sop
		_ = s.GetDB().Select("name").First(&parentSop, *input.ReferenceID).Error

		parentParams := map[string]any{
			"jobId":      data.ID,
			"parentId":   *input.ReferenceID,
			"parentName": parentSop.Name,
		}

		if err := graph.
			WithMatch("(j:Job {id: $jobId})").
			WithMerge("(p:SOP {id: $parentId})").
			WithSet("p.name = $parentName", nil).
			WithMerge("(j)-[:HAS_REFERENCE]->(p)").
			WithParams(parentParams).
			RunWrite(); err != nil {
			return fmt.Errorf("failed to create HAS_REFERENCE relationship: %w", err)
		}
	}

	return nil
}

func (s *sopJobService) updateGraphSopJob(data *models.SopJob) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"id":          data.ID,
		"name":        data.Name,
		"alias":       data.Alias,
		"type":        data.Type,
		"code":        data.Code,
		"description": data.Description,
		"titleId":     data.TitleID,
		"sopId":       data.SopID,
		"referenceId": data.ReferenceID,
		"index":       data.Index,
		"flowchartId": data.FlowchartID,
		"nextIndex":   data.NextIndex,
		"prevIndex":   data.PrevIndex,
		"isPublished": data.IsPublished,
		"isHide":      data.IsHide,
	}

	if err := graph.
		WithMatch("(j:Job {id: $id})").
		WithSet(`j.name = $name, j.alias = $alias, j.type = $type, 
		j.code = $code, j.description = $description, 
		j.title_id = $titleId, j.sop_id = $sopId, 
		j.reference_id = $referenceId, j.index = $index, 
		j.flowchart_id = $flowchartId, j.is_hide = $isHide,
		j.next_index = $nextIndex, j.prev_index = $prevIndex, j.is_published = $isPublished`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update Job graph with id %d: %w", data.ID, err)
	}

	deleteRefParam := map[string]any{
		"jobId": data.ID,
	}

	if err := graph.
		WithMatch("()-[r:HAS_JOB]->(j:Job {id: $jobId})").
		WithDelete("r").
		WithParams(deleteRefParam).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}
	if err := graph.
		WithMatch("(j:Job {id: $jobId})-[r:HAS_REFERENCE]->()").
		WithDelete("r").
		WithParams(deleteRefParam).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	if data.SopID != 0 {
		refParam := map[string]any{
			"jobId": data.ID,
			"sopId": data.SopID,
		}

		if err := graph.
			WithMatch("(j:Job {id: $jobId})").
			WithMatch("(sop:SOP {id: $sopId})").
			WithMerge("(sop)-[:HAS_JOB]->(j)").
			WithParams(refParam).
			RunWrite(); err != nil {
			return fmt.Errorf("failed to create HAS_JOB relation to SOP: %w", err)
		}
	}

	if data.ReferenceID != nil && data.Type != nil {
		refParam := map[string]any{
			"jobId": data.ID,
			"refId": data.ReferenceID,
		}

		if *data.Type == "sop" {
			if err := graph.
				WithMatch("(j:Job {id: $jobId})").
				WithMatch("(sop:SOP {id: $refId})").
				WithMerge("(j)-[:HAS_REFERENCE]->(sop)").
				WithParams(refParam).
				RunWrite(); err != nil {
				return fmt.Errorf("failed to create HAS_REFERENCE relation to SOP: %w", err)
			}
		} else if *data.Type == "spk" {
			if err := graph.
				WithMatch("(j:Job {id: $jobId})").
				WithMatch("(spk:SPK {id: $refId})").
				WithMerge("(j)-[:HAS_REFERENCE]->(spk)").
				WithParams(refParam).
				RunWrite(); err != nil {
				return fmt.Errorf("failed to create HAS_REFERENCE relation to SPK: %w", err)
			}
		}
	}

	return nil
}

func (s *sopJobService) deleteGraphSopJob(id int64) error {
	graph := builder.NewGraphRepository()
	params := map[string]any{
		"docId":     id,
		"deletedAt": time.Now().Unix(), // atau gunakan format timestamp yang sesuai
	}
	data, err := graph.
		WithMatch("(j:Job {id: $docId})").
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
		WithSet("n.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWriteWithReturn()

	if err != nil {
		return fmt.Errorf("failed to delete SOP graph with id %d: %w", id, err)
	}

	fmt.Println("deleted data:", data)
	return nil
}

func (s *sopJobService) bulkInsertGraphSopJobs(data []*models.SopJob) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	// Prepare batch data
	jobNodes := make([]map[string]any, 0, len(data))
	sopMap := make(map[int64]string) // sopId -> sopName

	// Load SOP names in batch
	sopIDs := make([]int64, 0, len(data))
	for _, job := range data {
		sopIDs = append(sopIDs, job.SopID)
	}

	var sops []models.Sop
	if err := s.GetDB().Select("id, name").Where("id IN ?", sopIDs).Find(&sops).Error; err != nil {
		return fmt.Errorf("failed to load SOP names: %w", err)
	}

	for _, sop := range sops {
		sopMap[sop.ID] = sop.Name
	}

	// Prepare job nodes data
	for _, job := range data {
		jobNode := map[string]any{
			"jobId":       job.ID,
			"jobName":     job.Name,
			"alias":       job.Alias,
			"type":        job.Type,
			"code":        job.Code,
			"description": job.Description,
			"titleId":     job.TitleID,
			"sopId":       job.SopID,
			"referenceId": job.ReferenceID,
			"index":       job.Index,
			"flowchartId": job.FlowchartID,
			"nextIndex":   job.NextIndex,
			"prevIndex":   job.PrevIndex,
			"isPublished": job.IsPublished,
			"isHide":      job.IsHide,
		}
		jobNodes = append(jobNodes, jobNode)
	}

	// 🔹 Step 1: Bulk merge Job nodes
	jobParams := map[string]any{
		"jobs": jobNodes,
	}

	if err := graph.
		WithUnwind("$jobs", "job").
		WithMerge("(j:Job {id: job.jobId})").
		WithSet(`j.name = job.jobName, j.alias = job.alias, j.type = job.type, 
			j.code = job.code, j.description = job.description, 
			j.title_id = job.titleId, j.sop_id = job.sopId, 
			j.reference_id = job.referenceId, j.index = job.index, 
			j.flowchart_id = job.flowchartId, j.is_hide = job.isHide,
			j.next_index = job.nextIndex, j.prev_index = job.prevIndex, j.is_published = job.isPublished`, nil).
		WithParams(jobParams).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to merge Job nodes: %w", err)
	}

	// 🔹 Step 2: Bulk merge SOP nodes
	sopNodes := make([]map[string]any, 0, len(sopMap))
	for sopID, sopName := range sopMap {
		sopNodes = append(sopNodes, map[string]any{
			"sopId":   sopID,
			"sopName": sopName,
		})
	}

	sopParams := map[string]any{
		"sops": sopNodes,
	}

	if err := graph.
		WithUnwind("$sops", "sop").
		WithMerge("(s:SOP {id: sop.sopId})").
		WithSet("s.name = sop.sopName", nil).
		WithParams(sopParams).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to merge SOP nodes: %w", err)
	}

	// 🔹 Step 3: Bulk create HAS_JOB relationships
	hasJobRels := make([]map[string]any, 0, len(data))
	for _, job := range data {
		hasJobRels = append(hasJobRels, map[string]any{
			"jobId": job.ID,
			"sopId": job.SopID,
		})
	}

	hasJobParams := map[string]any{
		"rels": hasJobRels,
	}

	if err := graph.
		WithUnwind("$rels", "rel").
		WithMatch("(j:Job {id: rel.jobId})").
		WithMatch("(s:SOP {id: rel.sopId})").
		WithMerge("(s)-[:HAS_JOB]->(j)").
		WithParams(hasJobParams).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to create HAS_JOB relationships: %w", err)
	}

	// 🔹 Step 4: Bulk create HAS_SOP relationships (for jobs with sop reference)
	hasSopRels := make([]map[string]any, 0)
	parentIDs := make([]int64, 0)

	for _, job := range data {
		if job.ReferenceID != nil && job.Type != nil && *job.Type == "sop" {
			hasSopRels = append(hasSopRels, map[string]any{
				"jobId":    job.ID,
				"parentId": *job.ReferenceID,
			})
			parentIDs = append(parentIDs, *job.ReferenceID)
		}
	}

	if len(hasSopRels) > 0 {
		// Load parent SOP names
		var parentSops []models.Sop
		if err := s.GetDB().Select("id, name").Where("id IN ?", parentIDs).Find(&parentSops).Error; err != nil {
			return fmt.Errorf("failed to load parent SOP names: %w", err)
		}

		parentSopMap := make(map[int64]string)
		for _, sop := range parentSops {
			parentSopMap[sop.ID] = sop.Name
		}

		// Add parent names to relationships
		for i := range hasSopRels {
			parentID := hasSopRels[i]["parentId"].(int64)
			hasSopRels[i]["parentName"] = parentSopMap[parentID]
		}

		hasSopParams := map[string]any{
			"rels": hasSopRels,
		}

		if err := graph.
			WithUnwind("$rels", "rel").
			WithMatch("(j:Job {id: rel.jobId})").
			WithMerge("(p:SOP {id: rel.parentId})").
			WithSet("p.name = rel.parentName", nil).
			WithMerge("(j)-[:HAS_REFERENCE]->(p)").
			WithParams(hasSopParams).
			RunWrite(); err != nil {
			return fmt.Errorf("failed to create HAS_REFERENCE relationships: %w", err)
		}
	}

	return nil
}

func (s *sopJobService) bulkUpdateGraphSopJobs(data []*models.SopJob) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	// Prepare job update data
	jobNodes := make([]map[string]any, 0, len(data))
	jobIDs := make([]int64, 0, len(data))

	for _, job := range data {
		jobIDs = append(jobIDs, job.ID)
		jobNode := map[string]any{
			"id":          job.ID,
			"name":        job.Name,
			"alias":       job.Alias,
			"type":        job.Type,
			"code":        job.Code,
			"description": job.Description,
			"titleId":     job.TitleID,
			"sopId":       job.SopID,
			"referenceId": job.ReferenceID,
			"index":       job.Index,
			"flowchartId": job.FlowchartID,
			"nextIndex":   job.NextIndex,
			"prevIndex":   job.PrevIndex,
			"isPublished": job.IsPublished,
			"isHide":      job.IsHide,
		}
		jobNodes = append(jobNodes, jobNode)
	}

	// 🔹 Step 1: Bulk update Job nodes properties
	jobParams := map[string]any{
		"jobs": jobNodes,
	}

	if err := graph.
		WithUnwind("$jobs", "job").
		WithMatch("(j:Job {id: job.id})").
		WithSet(`j.name = job.name, j.alias = job.alias, j.type = job.type, 
			j.code = job.code, j.description = job.description, 
			j.title_id = job.titleId, j.sop_id = job.sopId, 
			j.reference_id = job.referenceId, j.index = job.index, 
			j.flowchart_id = job.flowchartId, j.is_hide = job.isHide,
			j.next_index = job.nextIndex, j.prev_index = job.prevIndex, j.is_published = job.isPublished`, nil).
		WithParams(jobParams).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update Job nodes: %w", err)
	}

	// 🔹 Step 2: Delete old relationships - SIMPLE dengan WHERE IN
	deleteParams := map[string]any{
		"jobIds": jobIDs,
	}

	// Delete HAS_JOB relationships
	if err := graph.
		WithMatch("()-[r:HAS_JOB]->(j:Job)").
		WithWhere("j.id IN $jobIds", nil).
		WithDelete("r").
		WithParams(deleteParams).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to delete HAS_JOB relationships: %w", err)
	}

	// Delete HAS_REFERENCE relationships
	if err := graph.
		WithMatch("(j:Job)-[r:HAS_REFERENCE]->()").
		WithWhere("j.id IN $jobIds", nil).
		WithDelete("r").
		WithParams(deleteParams).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to delete HAS_REFERENCE relationships: %w", err)
	}

	// 🔹 Step 3: Recreate HAS_JOB relationships
	hasJobRels := make([]map[string]any, 0, len(data))
	sopIDs := make([]int64, 0)

	for _, job := range data {
		if job.SopID != 0 {
			hasJobRels = append(hasJobRels, map[string]any{
				"jobId": job.ID,
				"sopId": job.SopID,
			})
			sopIDs = append(sopIDs, job.SopID)
		}
	}

	if len(hasJobRels) > 0 {
		// Load SOP names
		var sops []models.Sop
		if err := s.GetDB().Select("id, name").Where("id IN ?", sopIDs).Find(&sops).Error; err != nil {
			return fmt.Errorf("failed to load SOP names: %w", err)
		}

		sopMap := make(map[int64]string)
		for _, sop := range sops {
			sopMap[sop.ID] = sop.Name
		}

		// Merge SOP nodes first
		sopNodes := make([]map[string]any, 0, len(sopMap))
		for sopID, sopName := range sopMap {
			sopNodes = append(sopNodes, map[string]any{
				"sopId":   sopID,
				"sopName": sopName,
			})
		}

		sopParams := map[string]any{
			"sops": sopNodes,
		}

		if err := graph.
			WithUnwind("$sops", "sop").
			WithMerge("(s:SOP {id: sop.sopId})").
			WithSet("s.name = sop.sopName", nil).
			WithParams(sopParams).
			RunWrite(); err != nil {
			return fmt.Errorf("failed to merge SOP nodes: %w", err)
		}

		// Create HAS_JOB relationships
		hasJobParams := map[string]any{
			"rels": hasJobRels,
		}

		if err := graph.
			WithUnwind("$rels", "rel").
			WithMatch("(j:Job {id: rel.jobId})").
			WithMatch("(s:SOP {id: rel.sopId})").
			WithMerge("(s)-[:HAS_JOB]->(j)").
			WithParams(hasJobParams).
			RunWrite(); err != nil {
			return fmt.Errorf("failed to create HAS_JOB relationships: %w", err)
		}
	}

	// 🔹 Step 4: Recreate HAS_REFERENCE relationships
	hasSopRefs := make([]map[string]any, 0)
	hasSpkRefs := make([]map[string]any, 0)
	parentSopIDs := make([]int64, 0)
	parentSpkIDs := make([]int64, 0)

	for _, job := range data {
		if job.ReferenceID != nil && job.Type != nil {
			if *job.Type == "sop" {
				hasSopRefs = append(hasSopRefs, map[string]any{
					"jobId": job.ID,
					"refId": *job.ReferenceID,
				})
				parentSopIDs = append(parentSopIDs, *job.ReferenceID)
			} else if *job.Type == "spk" {
				hasSpkRefs = append(hasSpkRefs, map[string]any{
					"jobId": job.ID,
					"refId": *job.ReferenceID,
				})
				parentSpkIDs = append(parentSpkIDs, *job.ReferenceID)
			}
		}
	}

	// Handle SOP references
	if len(hasSopRefs) > 0 {
		var parentSops []models.Sop
		if err := s.GetDB().Select("id, name").Where("id IN ?", parentSopIDs).Find(&parentSops).Error; err != nil {
			return fmt.Errorf("failed to load parent SOP names: %w", err)
		}

		parentSopMap := make(map[int64]string)
		for _, sop := range parentSops {
			parentSopMap[sop.ID] = sop.Name
		}

		// Merge parent SOP nodes
		parentSopNodes := make([]map[string]any, 0, len(parentSopMap))
		for sopID, sopName := range parentSopMap {
			parentSopNodes = append(parentSopNodes, map[string]any{
				"sopId":   sopID,
				"sopName": sopName,
			})
		}

		parentSopParams := map[string]any{
			"sops": parentSopNodes,
		}

		if err := graph.
			WithUnwind("$sops", "sop").
			WithMerge("(s:SOP {id: sop.sopId})").
			WithSet("s.name = sop.sopName", nil).
			WithParams(parentSopParams).
			RunWrite(); err != nil {
			return fmt.Errorf("failed to merge parent SOP nodes: %w", err)
		}

		// Create HAS_REFERENCE to SOP
		sopRefParams := map[string]any{
			"rels": hasSopRefs,
		}

		if err := graph.
			WithUnwind("$rels", "rel").
			WithMatch("(j:Job {id: rel.jobId})").
			WithMatch("(s:SOP {id: rel.refId})").
			WithMerge("(j)-[:HAS_REFERENCE]->(s)").
			WithParams(sopRefParams).
			RunWrite(); err != nil {
			return fmt.Errorf("failed to create HAS_REFERENCE to SOP: %w", err)
		}
	}

	// Handle SPK references
	if len(hasSpkRefs) > 0 {
		spkRefParams := map[string]any{
			"rels": hasSpkRefs,
		}

		if err := graph.
			WithUnwind("$rels", "rel").
			WithMatch("(j:Job {id: rel.jobId})").
			WithMatch("(spk:SPK {id: rel.refId})").
			WithMerge("(j)-[:HAS_REFERENCE]->(spk)").
			WithParams(spkRefParams).
			RunWrite(); err != nil {
			return fmt.Errorf("failed to create HAS_REFERENCE to SPK: %w", err)
		}
	}

	return nil
}

func (s *sopJobService) bulkDeleteGraphSopJobs(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	graph := builder.NewGraphRepository()
	params := map[string]any{
		"jobIds":    ids,
		"deletedAt": time.Now().Unix(),
	}
	if err := graph.
		WithMatch("(j:Job)").
		WithWhere("j.id IN $jobIds", nil).
		WithCall(`
			apoc.path.expandConfig(j, {
				relationshipFilter: ">",
				minLevel: 0,
				maxLevel: -1
			})
		`).
		WithYield("path").
		WithWith("collect(DISTINCT path) AS paths").
		WithUnwind("paths", "p").
		WithUnwind("nodes(p)", "n").
		WithWith("DISTINCT n").
		WithSet("n.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to delete Job nodes: %w", err)
	}
	return nil
}
