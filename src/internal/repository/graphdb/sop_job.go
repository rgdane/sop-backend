package graphdb

import (
	"fmt"
	"strings"
	"time"

	"jk-api/api/http/controllers/v1/dto"
	"jk-api/pkg/neo4j/builder"
)

type SopJobNode struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Alias       string `json:"alias"`
	Type        string `json:"type"`
	Code        string `json:"code"`
	Description string `json:"description"`
	TitleID     *int64 `json:"title_id"`
	SopID       int64  `json:"sop_id"`
	ReferenceID *int64 `json:"reference_id"`
	Index       int    `json:"index"`
	FlowchartID *int64 `json:"flowchart_id"`
	NextIndex   *int   `json:"next_index"`
	PrevIndex   *int   `json:"prev_index"`
	IsPublished *bool  `json:"is_published"`
	IsHide      *bool  `json:"is_hide"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type SopJobRepository interface {
	GetAllGraphSopJobs(filter dto.SopJobFilterDto) ([]*SopJobNode, error)
	GetGraphSopJobByID(id int64) (*SopJobNode, error)
	InsertGraphSopJob(data *SopJobNode) error
	UpdateGraphSopJob(data *SopJobNode) error
	DeleteGraphSopJob(sopJobId int64) error

	BulkInsertGraphSopJobs(data []*SopJobNode) error
	BulkUpdateGraphSopJobs(data []*SopJobNode) error
	BulkDeleteGraphSopJobs(ids []int64) error

	CountGraphSopJobs(filter dto.SopJobFilterDto) (int64, error)
}

type sopJobRepository struct{}

func NewSopJobRepository() SopJobRepository {
	return &sopJobRepository{}
}

func (r *sopJobRepository) GetAllGraphSopJobs(filter dto.SopJobFilterDto) ([]*SopJobNode, error) {
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

	if filter.SopID != 0 {
		conditions = append(conditions, "j.sop_id = $sopId")
		params["sopId"] = filter.SopID
	}

	if filter.TitleID != 0 {
		conditions = append(conditions, "j.title_id = $titleId")
		params["titleId"] = filter.TitleID
	}

	if filter.Type != nil && *filter.Type != "" {
		conditions = append(conditions, "j.type = $type")
		params["type"] = *filter.Type
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
		return nil, fmt.Errorf("failed to get SOP Jobs with filter: %w", err)
	}

	var sopJobs []*SopJobNode
	for _, record := range records {
		dataVal, ok := record.Get("data")
		if !ok {
			continue
		}

		props, ok := dataVal.(map[string]any)
		if !ok {
			continue
		}

		sopJob := mapToSopJobNode(props)
		sopJobs = append(sopJobs, sopJob)
	}

	return sopJobs, nil
}

func (r *sopJobRepository) GetGraphSopJobByID(id int64) (*SopJobNode, error) {
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
		return nil, fmt.Errorf("failed to get SOP Job node by ID: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("SOP Job node with ID %d not found", id)
	}

	dataVal, ok := records[0].Get("data")
	if !ok {
		return nil, fmt.Errorf("field 'data' not found in record")
	}

	props, ok := dataVal.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("data is not a map[string]any")
	}

	return mapToSopJobNode(props), nil
}

func mapToSopJobNode(props map[string]any) *SopJobNode {
	sopJob := &SopJobNode{}

	if idVal, ok := props["id"].(int64); ok {
		sopJob.ID = idVal
	}

	if nameVal, ok := props["name"].(string); ok {
		sopJob.Name = nameVal
	}

	if aliasVal, ok := props["alias"].(string); ok {
		sopJob.Alias = aliasVal
	}

	if typeVal, ok := props["type"].(string); ok {
		sopJob.Type = typeVal
	}

	if codeVal, ok := props["code"].(string); ok {
		sopJob.Code = codeVal
	}

	if descVal, ok := props["description"].(string); ok {
		sopJob.Description = descVal
	}

	if titleIDVal, ok := props["title_id"].(int64); ok {
		sopJob.TitleID = &titleIDVal
	}

	if sopIDVal, ok := props["sop_id"].(int64); ok {
		sopJob.SopID = sopIDVal
	}

	if refIDVal, ok := props["reference_id"].(int64); ok {
		sopJob.ReferenceID = &refIDVal
	}

	if indexVal, ok := props["index"].(int64); ok {
		sopJob.Index = int(indexVal)
	}

	if flowchartIDVal, ok := props["flowchart_id"].(int64); ok {
		sopJob.FlowchartID = &flowchartIDVal
	}

	if nextIndexVal, ok := props["next_index"].(int64); ok {
		nextIdx := int(nextIndexVal)
		sopJob.NextIndex = &nextIdx
	}

	if prevIndexVal, ok := props["prev_index"].(int64); ok {
		prevIdx := int(prevIndexVal)
		sopJob.PrevIndex = &prevIdx
	}

	if isPubVal, ok := props["is_published"].(bool); ok {
		sopJob.IsPublished = &isPubVal
	}

	if isHideVal, ok := props["is_hide"].(bool); ok {
		sopJob.IsHide = &isHideVal
	}

	if createdVal, ok := props["created_at"].(string); ok {
		sopJob.CreatedAt = createdVal
	}

	if updatedVal, ok := props["updated_at"].(string); ok {
		sopJob.UpdatedAt = updatedVal
	}

	return sopJob
}

func (r *sopJobRepository) InsertGraphSopJob(data *SopJobNode) error {
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
		"createdAt":   data.CreatedAt,
		"updatedAt":   data.UpdatedAt,
	}

	if err := graph.
		WithMerge("(j:Job {id: $id})").
		WithSet(`j.name = $name, 
			j.alias = $alias, 
			j.type = $type, 
			j.code = $code, 
			j.description = $description,
			j.title_id = $titleId, 
			j.sop_id = $sopId, 
			j.reference_id = $referenceId, 
			j.index = $index,
			j.flowchart_id = $flowchartId,
			j.next_index = $nextIndex, 
			j.prev_index = $prevIndex,
			j.is_published = $isPublished,
			j.is_hide = $isHide,
			j.created_at = datetime($createdAt), 
			j.updated_at = datetime($updatedAt)`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to merge Job node: %w", err)
	}

	return nil
}

func (r *sopJobRepository) UpdateGraphSopJob(data *SopJobNode) error {
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
		WithSet(`j.name = $name, 
			j.alias = $alias, 
			j.type = $type, 
			j.code = $code, 
			j.description = $description,
			j.title_id = $titleId, 
			j.sop_id = $sopId, 
			j.reference_id = $referenceId, 
			j.index = $index,
			j.flowchart_id = $flowchartId,
			j.next_index = $nextIndex, 
			j.prev_index = $prevIndex,
			j.is_published = $isPublished,
			j.is_hide = $isHide`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update Job graph with id %d: %w", data.ID, err)
	}

	return nil
}

func (r *sopJobRepository) DeleteGraphSopJob(sopJobId int64) error {
	graph := builder.NewGraphRepository()

	params := map[string]interface{}{
		"docId":     sopJobId,
		"deletedAt": time.Now().Format(time.RFC3339Nano),
	}

	if err := graph.
		WithMatch("(j:Job {id: $docId})").
		WithSet("j.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to soft delete Job graph with id %d: %w", sopJobId, err)
	}

	return nil
}

func (r *sopJobRepository) BulkInsertGraphSopJobs(data []*SopJobNode) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	sopJobNodes := make([]map[string]any, 0, len(data))
	for _, sopJob := range data {
		sopJobNodes = append(sopJobNodes, map[string]any{
			"id":          sopJob.ID,
			"name":        sopJob.Name,
			"alias":       sopJob.Alias,
			"type":        sopJob.Type,
			"code":        sopJob.Code,
			"description": sopJob.Description,
			"titleId":     sopJob.TitleID,
			"sopId":       sopJob.SopID,
			"referenceId": sopJob.ReferenceID,
			"index":       sopJob.Index,
			"flowchartId": sopJob.FlowchartID,
			"nextIndex":   sopJob.NextIndex,
			"prevIndex":   sopJob.PrevIndex,
			"isPublished": sopJob.IsPublished,
			"isHide":      sopJob.IsHide,
		})
	}

	params := map[string]any{"jobs": sopJobNodes}

	if err := graph.
		WithUnwind("$jobs", "j").
		WithMerge("(job:Job {id: j.id})").
		WithSet(`job.name = j.name, 
			job.alias = j.alias, 
			job.type = j.type, 
			job.code = j.code, 
			job.description = j.description,
			job.title_id = j.titleId, 
			job.sop_id = j.sopId, 
			job.reference_id = j.referenceId, 
			job.index = j.index,
			job.flowchart_id = j.flowchartId,
			job.next_index = j.nextIndex, 
			job.prev_index = j.prevIndex,
			job.is_published = j.isPublished,
			job.is_hide = j.isHide`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk insert Job nodes: %w", err)
	}

	return nil
}

func (r *sopJobRepository) BulkUpdateGraphSopJobs(data []*SopJobNode) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	sopJobNodes := make([]map[string]any, 0, len(data))
	for _, sopJob := range data {
		sopJobNodes = append(sopJobNodes, map[string]any{
			"id":          sopJob.ID,
			"name":        sopJob.Name,
			"alias":       sopJob.Alias,
			"type":        sopJob.Type,
			"code":        sopJob.Code,
			"description": sopJob.Description,
			"titleId":     sopJob.TitleID,
			"sopId":       sopJob.SopID,
			"referenceId": sopJob.ReferenceID,
			"index":       sopJob.Index,
			"flowchartId": sopJob.FlowchartID,
			"nextIndex":   sopJob.NextIndex,
			"prevIndex":   sopJob.PrevIndex,
			"isPublished": sopJob.IsPublished,
			"isHide":      sopJob.IsHide,
		})
	}

	params := map[string]any{"jobs": sopJobNodes}

	if err := graph.
		WithUnwind("$jobs", "j").
		WithMatch("(job:Job {id: j.id})").
		WithSet(`job.name = j.name, 
			job.alias = j.alias, 
			job.type = j.type, 
			job.code = j.code, 
			job.description = j.description,
			job.title_id = j.titleId, 
			job.sop_id = j.sopId, 
			job.reference_id = j.referenceId, 
			job.index = j.index,
			job.flowchart_id = j.flowchartId,
			job.next_index = j.nextIndex, 
			job.prev_index = j.prevIndex,
			job.is_published = j.isPublished,
			job.is_hide = j.isHide`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk update Job nodes: %w", err)
	}

	return nil
}

func (r *sopJobRepository) BulkDeleteGraphSopJobs(ids []int64) error {
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

func (r *sopJobRepository) CountGraphSopJobs(filter dto.SopJobFilterDto) (int64, error) {
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

	if filter.SopID != 0 {
		conditions = append(conditions, "j.sop_id = $sopId")
		params["sopId"] = filter.SopID
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
