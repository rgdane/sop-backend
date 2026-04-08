package graphdb

import (
	"fmt"
	"strings"
	"time"

	"jk-api/api/http/controllers/v1/dto"
	"jk-api/pkg/neo4j/builder"
)

type SpkJobNode struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SpkID       int64  `json:"spk_id"`
	SopID       *int64 `json:"sop_id"`
	TitleID     *int64 `json:"title_id"`
	Index       int    `json:"index"`
	FlowchartID *int64 `json:"flowchart_id"`
	NextIndex   *int   `json:"next_index"`
	PrevIndex   *int   `json:"prev_index"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type SpkJobRepository interface {
	GetAllGraphSpkJobs(filter dto.SpkJobFilterDto) ([]*SpkJobNode, error)
	GetGraphSpkJobByID(id int64) (*SpkJobNode, error)
	InsertGraphSpkJob(data *SpkJobNode) error
	UpdateGraphSpkJob(data *SpkJobNode) error
	DeleteGraphSpkJob(spkJobId int64) error

	BulkInsertGraphSpkJobs(data []*SpkJobNode) error
	BulkUpdateGraphSpkJobs(data []*SpkJobNode) error
	BulkDeleteGraphSpkJobs(ids []int64) error

	CountGraphSpkJobs(filter dto.SpkJobFilterDto) (int64, error)
}

type spkJobRepository struct{}

func NewSpkJobRepository() SpkJobRepository {
	return &spkJobRepository{}
}

func (r *spkJobRepository) GetAllGraphSpkJobs(filter dto.SpkJobFilterDto) ([]*SpkJobNode, error) {
	repo := builder.NewGraphRepository()
	params := make(map[string]any)

	repo = repo.WithMatch("(j:Job)")

	var conditions []string

	if filter.ShowDeleted {
		conditions = append(conditions, "j.deleted_at IS NOT NULL")
	} else {
		conditions = append(conditions, "j.deleted_at IS NULL")
	}

	if filter.Name != "" {
		conditions = append(conditions, "toLower(j.name) CONTAINS toLower($name)")
		params["name"] = filter.Name
	}

	if filter.SpkID != 0 {
		conditions = append(conditions, "j.spk_id = $spkId")
		params["spkId"] = filter.SpkID
	}

	if filter.SopID != 0 {
		conditions = append(conditions, "j.sop_id = $sopId")
		params["sopId"] = filter.SopID
	}

	if filter.TitleID != 0 {
		conditions = append(conditions, "j.title_id = $titleId")
		params["titleId"] = filter.TitleID
	}

	repo = repo.WithWhere(strings.Join(conditions, " AND "), params)

	returnClause := "j {.*} AS data"

	if filter.Sort != "" && filter.Order != "" {
		orderDir := strings.ToUpper(filter.Order)
		if orderDir != "ASC" && orderDir != "DESC" {
			orderDir = "ASC"
		}
		returnClause += fmt.Sprintf(" ORDER BY j.%s %s", filter.Sort, orderDir)
	}

	repo = repo.
		WithReturn(returnClause).
		WithParams(params)

	records, err := repo.RunRead()
	if err != nil {
		return nil, fmt.Errorf("failed to get SPK Jobs with filter: %w", err)
	}

	var spkJobs []*SpkJobNode
	for _, record := range records {
		dataVal, ok := record.Get("data")
		if !ok {
			continue
		}

		props, ok := dataVal.(map[string]any)
		if !ok {
			continue
		}

		spkJob := mapToSpkJobNode(props)
		spkJobs = append(spkJobs, spkJob)
	}

	return spkJobs, nil
}

func (r *spkJobRepository) GetGraphSpkJobByID(id int64) (*SpkJobNode, error) {
	repo := builder.NewGraphRepository()
	params := map[string]any{
		"id": id,
	}

	records, err := repo.
		WithMatch("(j:Job)").
		WithWhere("j.id = $id AND j.deleted_at IS NULL", params).
		WithReturn("j {.*} AS data").
		WithParams(params).
		RunRead()

	if err != nil {
		return nil, fmt.Errorf("failed to get SPK Job node by ID: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("SPK Job node with ID %d not found", id)
	}

	dataVal, ok := records[0].Get("data")
	if !ok {
		return nil, fmt.Errorf("field 'data' not found in record")
	}

	props, ok := dataVal.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("data is not a map[string]any")
	}

	return mapToSpkJobNode(props), nil
}

func mapToSpkJobNode(props map[string]any) *SpkJobNode {
	spkJob := &SpkJobNode{}

	if idVal, ok := props["id"].(int64); ok {
		spkJob.ID = idVal
	}

	if nameVal, ok := props["name"].(string); ok {
		spkJob.Name = nameVal
	}

	if descVal, ok := props["description"].(string); ok {
		spkJob.Description = descVal
	}

	if spkIDVal, ok := props["spk_id"].(int64); ok {
		spkJob.SpkID = spkIDVal
	}

	if sopIDVal, ok := props["sop_id"].(int64); ok {
		spkJob.SopID = &sopIDVal
	}

	if titleIDVal, ok := props["title_id"].(int64); ok {
		spkJob.TitleID = &titleIDVal
	}

	if indexVal, ok := props["index"].(int64); ok {
		spkJob.Index = int(indexVal)
	}

	if flowchartIDVal, ok := props["flowchart_id"].(int64); ok {
		spkJob.FlowchartID = &flowchartIDVal
	}

	if nextIndexVal, ok := props["next_index"].(int64); ok {
		nextIdx := int(nextIndexVal)
		spkJob.NextIndex = &nextIdx
	}

	if prevIndexVal, ok := props["prev_index"].(int64); ok {
		prevIdx := int(prevIndexVal)
		spkJob.PrevIndex = &prevIdx
	}

	if createdVal, ok := props["created_at"].(string); ok {
		spkJob.CreatedAt = createdVal
	}

	if updatedVal, ok := props["updated_at"].(string); ok {
		spkJob.UpdatedAt = updatedVal
	}

	return spkJob
}

func (r *spkJobRepository) InsertGraphSpkJob(data *SpkJobNode) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"id":          data.ID,
		"name":        data.Name,
		"description": data.Description,
		"spkId":       data.SpkID,
		"sopId":       data.SopID,
		"titleId":     data.TitleID,
		"index":       data.Index,
		"flowchartId": data.FlowchartID,
		"nextIndex":   data.NextIndex,
		"prevIndex":   data.PrevIndex,
		"createdAt":   data.CreatedAt,
		"updatedAt":   data.UpdatedAt,
	}

	if err := graph.
		WithMerge("(j:Job {id: $id})").
		WithSet(`j.name = $name, 
			j.description = $description,
			j.spk_id = $spkId,
			j.sop_id = $sopId,
			j.title_id = $titleId,
			j.index = $index,
			j.flowchart_id = $flowchartId,
			j.next_index = $nextIndex,
			j.prev_index = $prevIndex,
			j.created_at = datetime($createdAt),
			j.updated_at = datetime($updatedAt)`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to merge Job node: %w", err)
	}

	return nil
}

func (r *spkJobRepository) UpdateGraphSpkJob(data *SpkJobNode) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"id":          data.ID,
		"name":        data.Name,
		"description": data.Description,
		"spkId":       data.SpkID,
		"sopId":       data.SopID,
		"titleId":     data.TitleID,
		"index":       data.Index,
		"flowchartId": data.FlowchartID,
		"nextIndex":   data.NextIndex,
		"prevIndex":   data.PrevIndex,
	}

	if err := graph.
		WithMatch("(j:Job {id: $id})").
		WithSet(`j.name = $name, 
			j.description = $description,
			j.spk_id = $spkId,
			j.sop_id = $sopId,
			j.title_id = $titleId,
			j.index = $index,
			j.flowchart_id = $flowchartId,
			j.next_index = $nextIndex,
			j.prev_index = $prevIndex`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update Job graph with id %d: %w", data.ID, err)
	}

	return nil
}

func (r *spkJobRepository) DeleteGraphSpkJob(spkJobId int64) error {
	graph := builder.NewGraphRepository()

	params := map[string]interface{}{
		"docId":     spkJobId,
		"deletedAt": time.Now().Format(time.RFC3339Nano),
	}

	if err := graph.
		WithMatch("(j:Job {id: $docId})").
		WithSet("j.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to soft delete Job graph with id %d: %w", spkJobId, err)
	}

	return nil
}

func (r *spkJobRepository) BulkInsertGraphSpkJobs(data []*SpkJobNode) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	spkJobNodes := make([]map[string]any, 0, len(data))
	for _, spkJob := range data {
		spkJobNodes = append(spkJobNodes, map[string]any{
			"id":          spkJob.ID,
			"name":        spkJob.Name,
			"description": spkJob.Description,
			"spkId":       spkJob.SpkID,
			"sopId":       spkJob.SopID,
			"titleId":     spkJob.TitleID,
			"index":       spkJob.Index,
			"flowchartId": spkJob.FlowchartID,
			"nextIndex":   spkJob.NextIndex,
			"prevIndex":   spkJob.PrevIndex,
		})
	}

	params := map[string]any{"jobs": spkJobNodes}

	if err := graph.
		WithUnwind("$jobs", "j").
		WithMerge("(job:Job {id: j.id})").
		WithSet(`job.name = j.name, 
			job.description = j.description,
			job.spk_id = j.spkId,
			job.sop_id = j.sopId,
			job.title_id = j.titleId,
			job.index = j.index,
			job.flowchart_id = j.flowchartId,
			job.next_index = j.nextIndex,
			job.prev_index = j.prevIndex`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk insert Job nodes: %w", err)
	}

	return nil
}

func (r *spkJobRepository) BulkUpdateGraphSpkJobs(data []*SpkJobNode) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	spkJobNodes := make([]map[string]any, 0, len(data))
	for _, spkJob := range data {
		spkJobNodes = append(spkJobNodes, map[string]any{
			"id":          spkJob.ID,
			"name":        spkJob.Name,
			"description": spkJob.Description,
			"spkId":       spkJob.SpkID,
			"sopId":       spkJob.SopID,
			"titleId":     spkJob.TitleID,
			"index":       spkJob.Index,
			"flowchartId": spkJob.FlowchartID,
			"nextIndex":   spkJob.NextIndex,
			"prevIndex":   spkJob.PrevIndex,
		})
	}

	params := map[string]any{"jobs": spkJobNodes}

	if err := graph.
		WithUnwind("$jobs", "j").
		WithMatch("(job:Job {id: j.id})").
		WithSet(`job.name = j.name, 
			job.description = j.description,
			job.spk_id = j.spkId,
			job.sop_id = j.sopId,
			job.title_id = j.titleId,
			job.index = j.index,
			job.flowchart_id = j.flowchartId,
			job.next_index = j.nextIndex,
			job.prev_index = j.prevIndex`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk update Job nodes: %w", err)
	}

	return nil
}

func (r *spkJobRepository) BulkDeleteGraphSpkJobs(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	params := map[string]any{
		"jobIds":    ids,
		"deletedAt": time.Now().Format(time.RFC3339Nano),
	}

	if err := graph.
		WithMatch("(j:Job)").
		WithWhere("j.id IN $jobIds", nil).
		WithSet("j.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk soft delete Job nodes: %w", err)
	}

	return nil
}

func (r *spkJobRepository) CountGraphSpkJobs(filter dto.SpkJobFilterDto) (int64, error) {
	repo := builder.NewGraphRepository()
	params := make(map[string]any)

	repo = repo.WithMatch("(j:Job)")

	var conditions []string

	if filter.ShowDeleted {
		conditions = append(conditions, "j.deleted_at IS NOT NULL")
	} else {
		conditions = append(conditions, "j.deleted_at IS NULL")
	}

	if filter.Name != "" {
		conditions = append(conditions, "toLower(j.name) CONTAINS toLower($name)")
		params["name"] = filter.Name
	}

	if filter.SpkID != 0 {
		conditions = append(conditions, "j.spk_id = $spkId")
		params["spkId"] = filter.SpkID
	}

	repo = repo.
		WithWhere(strings.Join(conditions, " AND "), params).
		WithReturn("count(j) AS total").
		WithParams(params)

	records, err := repo.RunRead()
	if err != nil {
		return 0, fmt.Errorf("failed to count Job nodes: %w", err)
	}

	if len(records) > 0 {
		if totalVal, ok := records[0].Get("total"); ok {
			if total, isInt := totalVal.(int64); isInt {
				return total, nil
			}
		}
	}

	return 0, nil
}
