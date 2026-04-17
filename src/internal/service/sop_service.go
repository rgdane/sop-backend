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

type SopService interface {
	WithTx(tx *gorm.DB) SopService

	CreateSop(input *models.Sop) (*models.Sop, error)
	UpdateSop(id int64, updates map[string]interface{}, associations map[string]interface{}) (*models.Sop, error)
	DeleteSop(id int64, isPermanent bool) error
	GetAllSops(filter dto.SopFilterDto) ([]models.Sop, error)
	GetSopByID(id int64, filter dto.SopFilterDto) (*models.Sop, error)
	GetSopsByIDs(ids []int64) ([]*models.Sop, error)
	GetDB() *gorm.DB
	BulkCreateSops(data []*models.Sop) ([]*models.Sop, error)
	BulkUpdateSops(ids []int64, updates map[string]interface{}) error
	BulkDeleteSops(ids []int64, isPermanent bool) error
	CountSops(filter dto.SopFilterDto) (int64, error)

	GetAllGraphSops(filter dto.SopFilterDto) ([]*graphdb.SopNode, error)
	GetGraphSopByID(id int64) (*graphdb.SopNode, error)
	InsertGraphSop(data *graphdb.SopNode) error
	UpdateGraphSop(data *graphdb.SopNode) error
	DeleteGraphSop(sopId int64) error
	BulkInsertGraphSops(data []*graphdb.SopNode) error
	BulkUpdateGraphSops(data []*graphdb.SopNode) error
	BulkDeleteGraphSops(ids []int64) error
	CountGraphSops(filter dto.SopFilterDto) (int64, error)
}

type sopService struct {
	repo      sql.SopRepository
	graphRepo graphdb.SopRepository
	tx        *gorm.DB
}

func NewSopService(repo sql.SopRepository, graphRepo graphdb.SopRepository) SopService {
	return &sopService{repo: repo, graphRepo: graphRepo}
}

func (s *sopService) WithTx(tx *gorm.DB) SopService {
	return &sopService{
		repo:      s.repo.WithTx(tx),
		graphRepo: s.graphRepo,
		tx:        tx,
	}
}

func (s *sopService) GetDB() *gorm.DB {
	if s.tx != nil {
		return s.tx
	}
	return config.DB
}

func (s *sopService) CreateSop(input *models.Sop) (*models.Sop, error) {
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	data, err := s.repo.InsertSop(input)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	if err := s.InsertGraphSop(toSopNode(data)); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}
	return data, nil
}

func (s *sopService) UpdateSop(id int64, updates map[string]interface{}, associations map[string]interface{}) (*models.Sop, error) {
	repo := s.repo

	if len(associations) > 0 {
		assocNames := make([]string, 0, len(associations))
		for name := range associations {
			assocNames = append(assocNames, name)
			delete(updates, name)
		}
		repo = repo.WithAssociations(assocNames...).WithReplacements(associations)
	}

	data, err := repo.UpdateSop(id, updates)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	if err := s.UpdateGraphSop(toSopNode(data)); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}
	return data, nil
}

func (s *sopService) DeleteSop(id int64, isPermanent bool) (err error) {
	repo := s.repo
	if isPermanent {
		repo = repo.WithUnscoped()
		err = repo.RemoveSop(id)
		return gorm_err.TranslateGormError(err)
	}

	data, err := repo.FindSopByID(id)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	payload := map[string]any{
		"name": fmt.Sprintf("DELETED-%s", data.Name),
	}

	if _, err = s.UpdateSop(id, payload, nil); err != nil {
		return gorm_err.TranslateGormError(err)
	}

	if err := s.DeleteGraphSop(id); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	err = repo.RemoveSop(id)
	return gorm_err.TranslateGormError(err)
}

func (s *sopService) GetAllSops(filter dto.SopFilterDto) ([]models.Sop, error) {
	repo := s.repo

	if filter.Limit != 0 {
		repo = repo.WithLimit(int(filter.Limit))
	}
	if filter.Cursor != 0 {
		repo = repo.WithCursor(int(filter.Cursor))
	}
	if filter.Preload {
		repo = repo.WithPreloads("HasJobs", "HasDivisions")
	}
	if filter.Name != "" {
		repo = repo.WithWhere("name ILIKE ?", "%"+filter.Name+"%")
	}
	if filter.TitleID != 0 {
		repo = repo.WithJoins("JOIN sop_titles ON sop_titles.sop_id = sops.id").WithWhere("sop_titles.title_id = ?", filter.TitleID)
	}
	if filter.Code != nil {
		repo = repo.WithWhere("code ILIKE ?", "%"+*filter.Code+"%")
	}
	if filter.DivisionID != 0 {
		repo = repo.WithJoins("JOIN sop_divisions ON sop_divisions.sop_id = sops.id").WithWhere("sop_divisions.division_id = ?", filter.DivisionID)
	}
	if len(filter.DivisionIDs) > 0 {
		repo = repo.WithJoins("JOIN sop_divisions ON sop_divisions.sop_id = sops.id").WithWhere("sop_divisions.division_id IN ?", filter.DivisionIDs)
	}
	if filter.ShowDeleted {
		repo = repo.WithUnscoped().WithWhere("deleted_at IS NOT NULL")
	}
	if filter.ExcludeID != 0 {
		repo = repo.WithWhere("parent_job_id IS NULL AND sops.id != ?", filter.ExcludeID)
	}

	data, err := repo.FindSops()
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	return data, nil
}

func (s *sopService) GetSopByID(id int64, filter dto.SopFilterDto) (*models.Sop, error) {
	repo := s.repo
	if filter.Preload {
		repo = repo.WithPreloads("HasJobs", "HasDivisions")
	}
	data, err := repo.FindSopByID(id)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *sopService) BulkCreateSops(data []*models.Sop) ([]*models.Sop, error) {
	datas, err := s.repo.InsertManySops(data)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	if err := s.BulkInsertGraphSops(toSopNodeSlice(datas)); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}
	return datas, nil
}

func (s *sopService) BulkUpdateSops(ids []int64, updates map[string]interface{}) error {
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
			if err := s.sopRestore(ids); err != nil {
				return err
			}
			err := repo.UpdateManySops(ids, updates)
			return gorm_err.TranslateGormError(err)
		}
	}

	updatedSops, err := s.repo.FindSopsByIDs(ids)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	if err := s.BulkUpdateGraphSops(toSopNodeSlice(updatedSops)); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	err = repo.UpdateManySops(ids, updates)
	return gorm_err.TranslateGormError(err)
}

func (s *sopService) GetSopsByIDs(ids []int64) ([]*models.Sop, error) {
	data, err := s.repo.FindSopsByIDs(ids)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *sopService) BulkDeleteSops(ids []int64, isPermanent bool) (err error) {
	repo := s.repo

	if isPermanent {
		repo = repo.WithUnscoped()
		err = repo.RemoveManySops(ids)
		return gorm_err.TranslateGormError(err)
	}

	if err := s.BulkDeleteGraphSops(ids); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	err = repo.RemoveManySops(ids)
	return gorm_err.TranslateGormError(err)
}

func (s *sopService) CountSops(filter dto.SopFilterDto) (int64, error) {
	repo := s.repo
	if filter.TitleID != 0 {
		repo = repo.WithJoins("JOIN sop_titles ON sops.id = sop_titles.sop_id").
			WithWhere("sop_titles.title_id = ?", filter.TitleID)
	}

	data, err := repo.CountSops()
	if err != nil {
		return 0, gorm_err.TranslateGormError(err)
	}

	return data, nil
}

func (s *sopService) GetAllGraphSops(filter dto.SopFilterDto) ([]*graphdb.SopNode, error) {
	return s.graphRepo.GetAllGraphSops(filter)
}

func (s *sopService) GetGraphSopByID(id int64) (*graphdb.SopNode, error) {
	return s.graphRepo.GetGraphSopByID(id)
}

func (s *sopService) InsertGraphSop(data *graphdb.SopNode) error {
	return s.graphRepo.InsertGraphSop(data)
}

func (s *sopService) UpdateGraphSop(data *graphdb.SopNode) error {
	return s.graphRepo.UpdateGraphSop(data)
}

func (s *sopService) DeleteGraphSop(sopId int64) error {
	return s.graphRepo.DeleteGraphSop(sopId)
}

func (s *sopService) BulkInsertGraphSops(data []*graphdb.SopNode) error {
	return s.graphRepo.BulkInsertGraphSops(data)
}

func (s *sopService) BulkUpdateGraphSops(data []*graphdb.SopNode) error {
	return s.graphRepo.BulkUpdateGraphSops(data)
}

func (s *sopService) BulkDeleteGraphSops(ids []int64) error {
	return s.graphRepo.BulkDeleteGraphSops(ids)
}

func (s *sopService) CountGraphSops(filter dto.SopFilterDto) (int64, error) {
	return s.graphRepo.CountGraphSops(filter)
}

func toSopNode(m *models.Sop) *graphdb.SopNode {
	return &graphdb.SopNode{
		ID:          m.ID,
		Name:        m.Name,
		Code:        m.Code,
		Description: helper.ToJSONString(m.Description),
		ParentJobID: m.ParentJobID,
		CreatedAt:   m.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   m.UpdatedAt.Format(time.RFC3339Nano),
	}
}

func toSopNodeSlice(m []*models.Sop) []*graphdb.SopNode {
	result := make([]*graphdb.SopNode, 0, len(m))
	for _, sop := range m {
		result = append(result, toSopNode(sop))
	}
	return result
}

func (s *sopService) sopRestore(ids []int64) error {
	var sops []models.Sop
	if err := s.GetDB().Unscoped().Preload("HasJobs").Preload("HasDivisions").Where("id IN ? AND deleted_at IS NOT NULL", ids).Find(&sops).Error; err != nil {
		return err
	}

	for _, sop := range sops {
		sop.Name = strings.TrimPrefix(sop.Name, "DELETED-")
		var divisionID int64
		if len(sop.HasDivisions) > 0 {
			divisionID = sop.HasDivisions[0].ID
		}
		newCode := sop.GenerateSopCode(s.GetDB(), divisionID, sop.ID)
		if newCode != "" {
			if err := s.GetDB().Unscoped().Model(&sop).Where("id = ?", sop.ID).Update("code", newCode).Update("name", sop.Name).Error; err != nil {
				return err
			}
			sop.Code = newCode
		}

		var updatedSop models.Sop
		if err := s.GetDB().Unscoped().Preload("HasJobs").Where("id = ?", sop.ID).First(&updatedSop).Error; err != nil {
			return err
		}

		if updatedSop.CreatedAt.IsZero() {
			updatedSop.CreatedAt = time.Now()
		}
		updatedSop.UpdatedAt = time.Now()

		if err := s.removeDeletedAtFromGraph(updatedSop.ID); err != nil {
			return fmt.Errorf("failed to remove deleted_at from graph: %w", err)
		}

		if err := s.insertGraphSopWithJobs(&updatedSop); err != nil {
			return fmt.Errorf("failed to restore SOP graph: %w", err)
		}
	}
	return nil
}

func (s *sopService) insertGraphSopWithJobs(data *models.Sop) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"name":        data.Name,
		"code":        data.Code,
		"description": helper.ToJSONString(data.Description),
		"id":          data.ID,
		"parentJobId": data.ParentJobID,
		"createdAt":   data.CreatedAt.Format(time.RFC3339Nano),
		"updatedAt":   data.UpdatedAt.Format(time.RFC3339Nano),
	}

	if err := graph.
		WithMerge("(s:SOP {id: $id})").
		WithSet("s.name = $name, s.code = $code, s.description = $description, s.created_at = datetime($createdAt), s.updated_at = datetime($updatedAt)", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to merge SOP node: %w", err)
	}

	if len(data.HasJobs) > 0 {
		fmt.Printf("Creating relationships for %d jobs\n", len(data.HasJobs))
		for _, job := range data.HasJobs {
			fmt.Printf("Creating Job node and HAS_JOB relationship: SOP ID %d -> Job ID %d\n", data.ID, job.ID)

			jobParams := map[string]any{
				"sopId":      data.ID,
				"jobId":      job.ID,
				"jobName":    job.Name,
				"jobCode":    job.Code,
				"jobDesc":    helper.ToJSONString(job.Description),
				"jobCreated": job.CreatedAt.Format(time.RFC3339Nano),
				"jobUpdated": job.UpdatedAt.Format(time.RFC3339Nano),
				"jobIndex":   job.Index,
			}

			if err := graph.
				WithMerge("(j:Job {id: $jobId})").
				WithSet("j.name = $jobName, j.index = $jobIndex, j.code = $jobCode, j.description = $jobDesc, j.created_at = datetime($jobCreated), j.updated_at = datetime($jobUpdated)", nil).
				WithWith("j").
				WithMatch("(s:SOP {id: $sopId})").
				WithMerge("(s)-[:HAS_JOB]->(j)").
				WithParams(jobParams).
				RunWrite(); err != nil {
				fmt.Printf("Failed to create job node/relationship: %v\n", err)
				return fmt.Errorf("failed to create Job node and HAS_JOB relationship for job %d: %w", job.ID, err)
			}
		}
	} else {
		fmt.Println("No jobs found to create relationships")
	}

	return nil
}

func (s *sopService) removeDeletedAtFromGraph(sopID int64) error {
	graph := builder.NewGraphRepository()
	params := map[string]interface{}{
		"sopId": sopID,
	}

	if err := graph.
		WithMatch("(s:SOP {id: $sopId})").
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
		return fmt.Errorf("failed to remove deleted_at from SOP graph with id %d: %w", sopID, err)
	}
	return nil
}

func GetSOPGraphs() (any, error) {
	graph := builder.NewGraphRepository()

	result, err := graph.
		WithMatch("(s:SOP)").
		WithWhere("s.deleted_at IS NULL", nil).
		WithOptionalMatch("(s)-[:HAS_JOB]->(j:Job)").
		WithWhere("j.deleted_at IS NULL", nil).
		WithWith("s, collect(apoc.map.removeKey(apoc.convert.toMap(j), 'deleted_at')) AS jobs").
		WithWith("apoc.map.removeKey(apoc.convert.toMap(s), 'deleted_at') AS sMap").
		WithWith("apoc.map.setKey(sMap, 'has_job', jobs) AS sop").
		WithReturn("sop").
		RunWriteWithReturn()
	if err != nil {
		return nil, err
	}

	return helper.Neo4jFormatter(result), nil
}
