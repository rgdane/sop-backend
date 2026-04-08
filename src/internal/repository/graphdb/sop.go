package graphdb

import (
	"fmt"
	"strings"
	"time"

	"jk-api/api/http/controllers/v1/dto"
	"jk-api/pkg/neo4j/builder"
)

type SopNode struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	ParentJobID *int64 `json:"parent_job_id,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type SopRepository interface {
	GetAllGraphSops(filter dto.SopFilterDto) ([]*SopNode, error)
	GetGraphSopByID(id int64) (*SopNode, error)
	InsertGraphSop(data *SopNode) error
	UpdateGraphSop(data *SopNode) error
	DeleteGraphSop(sopId int64) error

	BulkInsertGraphSops(data []*SopNode) error
	BulkUpdateGraphSops(data []*SopNode) error
	BulkDeleteGraphSops(ids []int64) error

	CountGraphSops(filter dto.SopFilterDto) (int64, error)
}

type sopRepository struct{}

func NewSopRepository() SopRepository {
	return &sopRepository{}
}

func (r *sopRepository) GetAllGraphSops(filter dto.SopFilterDto) ([]*SopNode, error) {
	repo := builder.NewGraphRepository()
	params := make(map[string]any)

	repo = repo.WithMatch("(s:SOP)")

	var conditions []string

	if filter.ShowDeleted {
		conditions = append(conditions, "s.deleted_at IS NOT NULL")
	} else {
		conditions = append(conditions, "s.deleted_at IS NULL")
	}

	if filter.Name != "" {
		conditions = append(conditions, "toLower(s.name) CONTAINS toLower($name)")
		params["name"] = filter.Name
	}

	if filter.Code != nil && *filter.Code != "" {
		conditions = append(conditions, "toLower(s.code) CONTAINS toLower($code)")
		params["code"] = *filter.Code
	}

	repo = repo.WithWhere(strings.Join(conditions, " AND "), params)

	returnClause := "s {.*} AS data"

	if filter.Sort != "" && filter.Order != "" {
		orderDir := strings.ToUpper(filter.Order)
		if orderDir != "ASC" && orderDir != "DESC" {
			orderDir = "ASC"
		}
		returnClause += fmt.Sprintf(" ORDER BY s.%s %s", filter.Sort, orderDir)
	}

	repo = repo.
		WithReturn(returnClause).
		WithParams(params)

	records, err := repo.RunRead()
	if err != nil {
		return nil, fmt.Errorf("failed to get SOPs with filter: %w", err)
	}

	var sops []*SopNode
	for _, record := range records {
		dataVal, ok := record.Get("data")
		if !ok {
			continue
		}

		props, ok := dataVal.(map[string]any)
		if !ok {
			continue
		}

		sop := mapToSopNode(props)
		sops = append(sops, sop)
	}

	return sops, nil
}

func (r *sopRepository) GetGraphSopByID(id int64) (*SopNode, error) {
	repo := builder.NewGraphRepository()
	params := map[string]any{
		"id": id,
	}

	records, err := repo.
		WithMatch("(s:SOP)").
		WithWhere("s.id = $id AND s.deleted_at IS NULL", params).
		WithReturn("s {.*} AS data").
		WithParams(params).
		RunRead()

	if err != nil {
		return nil, fmt.Errorf("failed to get SOP node by ID: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("SOP node with ID %d not found", id)
	}

	dataVal, ok := records[0].Get("data")
	if !ok {
		return nil, fmt.Errorf("field 'data' not found in record")
	}

	props, ok := dataVal.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("data is not a map[string]any")
	}

	return mapToSopNode(props), nil
}

func mapToSopNode(props map[string]any) *SopNode {
	sop := &SopNode{}

	if idVal, ok := props["id"].(int64); ok {
		sop.ID = idVal
	}

	if nameVal, ok := props["name"].(string); ok {
		sop.Name = nameVal
	}

	if codeVal, ok := props["code"].(string); ok {
		sop.Code = codeVal
	}

	if descVal, ok := props["description"].(string); ok {
		sop.Description = descVal
	}

	if parentJobVal, ok := props["parent_job_id"].(int64); ok {
		sop.ParentJobID = &parentJobVal
	}

	if createdVal, ok := props["created_at"].(string); ok {
		sop.CreatedAt = createdVal
	}

	if updatedVal, ok := props["updated_at"].(string); ok {
		sop.UpdatedAt = updatedVal
	}

	return sop
}

func (r *sopRepository) InsertGraphSop(data *SopNode) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"id":          data.ID,
		"name":        data.Name,
		"code":        data.Code,
		"description": data.Description,
		"parentJobId": data.ParentJobID,
		"createdAt":   data.CreatedAt,
		"updatedAt":   data.UpdatedAt,
	}

	if err := graph.
		WithMerge("(s:SOP {id: $id})").
		WithSet(`s.name = $name, 
			s.code = $code, 
			s.description = $description,
			s.parent_job_id = $parentJobId,
			s.created_at = datetime($createdAt), 
			s.updated_at = datetime($updatedAt)`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to merge SOP node: %w", err)
	}

	return nil
}

func (r *sopRepository) UpdateGraphSop(data *SopNode) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"id":          data.ID,
		"name":        data.Name,
		"code":        data.Code,
		"description": data.Description,
		"parentJobId": data.ParentJobID,
	}

	if err := graph.
		WithMatch("(s:SOP {id: $id})").
		WithSet("s.name = $name, s.code = $code, s.description = $description, s.parent_job_id = $parentJobId", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update SOP graph with id %d: %w", data.ID, err)
	}

	return nil
}

func (r *sopRepository) DeleteGraphSop(sopId int64) error {
	graph := builder.NewGraphRepository()

	params := map[string]interface{}{
		"docId":     sopId,
		"deletedAt": time.Now().Format(time.RFC3339Nano),
	}

	if err := graph.
		WithMatch("(s:SOP {id: $docId})").
		WithSet("s.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to soft delete SOP graph with id %d: %w", sopId, err)
	}

	return nil
}

func (r *sopRepository) BulkInsertGraphSops(data []*SopNode) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	sopNodes := make([]map[string]any, 0, len(data))
	for _, sop := range data {
		sopNodes = append(sopNodes, map[string]any{
			"id":          sop.ID,
			"code":        sop.Code,
			"name":        sop.Name,
			"description": sop.Description,
			"parentJobId": sop.ParentJobID,
		})
	}

	params := map[string]any{"sops": sopNodes}

	if err := graph.
		WithUnwind("$sops", "s").
		WithMerge("(sop:SOP {id: s.id})").
		WithSet(`sop.code = s.code, sop.name = s.name, sop.description = s.description, sop.parent_job_id = s.parentJobId`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk insert SOP nodes: %w", err)
	}

	return nil
}

func (r *sopRepository) BulkUpdateGraphSops(data []*SopNode) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	sopNodes := make([]map[string]any, 0, len(data))
	for _, sop := range data {
		sopNodes = append(sopNodes, map[string]any{
			"id":          sop.ID,
			"code":        sop.Code,
			"name":        sop.Name,
			"description": sop.Description,
			"parentJobId": sop.ParentJobID,
		})
	}

	params := map[string]any{"sops": sopNodes}

	if err := graph.
		WithUnwind("$sops", "s").
		WithMatch("(sop:SOP {id: s.id})").
		WithSet("sop.code = s.code, sop.name = s.name, sop.description = s.description, sop.parent_job_id = s.parentJobId", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk update SOP nodes: %w", err)
	}

	return nil
}

func (r *sopRepository) BulkDeleteGraphSops(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	params := map[string]any{
		"sopIds":    ids,
		"deletedAt": time.Now().Format(time.RFC3339Nano),
	}

	if err := graph.
		WithMatch("(s:SOP)").
		WithWhere("s.id IN $sopIds", nil).
		WithSet("s.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk soft delete SOP nodes: %w", err)
	}

	return nil
}

func (r *sopRepository) CountGraphSops(filter dto.SopFilterDto) (int64, error) {
	repo := builder.NewGraphRepository()
	params := make(map[string]any)

	repo = repo.WithMatch("(s:SOP)")

	var conditions []string

	if filter.ShowDeleted {
		conditions = append(conditions, "s.deleted_at IS NOT NULL")
	} else {
		conditions = append(conditions, "s.deleted_at IS NULL")
	}

	if filter.Name != "" {
		conditions = append(conditions, "toLower(s.name) CONTAINS toLower($name)")
		params["name"] = filter.Name
	}

	repo = repo.
		WithWhere(strings.Join(conditions, " AND "), params).
		WithReturn("count(s) AS total").
		WithParams(params)

	records, err := repo.RunRead()
	if err != nil {
		return 0, fmt.Errorf("failed to count SOP nodes: %w", err)
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
