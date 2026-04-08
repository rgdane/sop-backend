package graphdb

import (
	"fmt"
	"strings"
	"time"

	"jk-api/api/http/controllers/v1/dto"
	"jk-api/pkg/neo4j/builder"
)

type SpkNode struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type SpkRepository interface {
	GetAllGraphSpks(filter dto.SpkFilterDto) ([]*SpkNode, error)
	GetGraphSpkByID(id int64) (*SpkNode, error)
	InsertGraphSpk(data *SpkNode) error
	UpdateGraphSpk(data *SpkNode) error
	DeleteGraphSpk(spkId int64) error

	BulkInsertGraphSpks(data []*SpkNode) error
	BulkUpdateGraphSpks(data []*SpkNode) error
	BulkDeleteGraphSpks(ids []int64) error

	CountGraphSpks(filter dto.SpkFilterDto) (int64, error)
}

type spkRepository struct{}

func NewSpkRepository() SpkRepository {
	return &spkRepository{}
}

func (r *spkRepository) GetAllGraphSpks(filter dto.SpkFilterDto) ([]*SpkNode, error) {
	repo := builder.NewGraphRepository()
	params := make(map[string]any)

	repo = repo.WithMatch("(s:SPK)")

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
		return nil, fmt.Errorf("failed to get SPKs with filter: %w", err)
	}

	var spks []*SpkNode
	for _, record := range records {
		dataVal, ok := record.Get("data")
		if !ok {
			continue
		}

		props, ok := dataVal.(map[string]any)
		if !ok {
			continue
		}

		spk := mapToSpkNode(props)
		spks = append(spks, spk)
	}

	return spks, nil
}

func (r *spkRepository) GetGraphSpkByID(id int64) (*SpkNode, error) {
	repo := builder.NewGraphRepository()
	params := map[string]any{
		"id": id,
	}

	records, err := repo.
		WithMatch("(s:SPK)").
		WithWhere("s.id = $id AND s.deleted_at IS NULL", params).
		WithReturn("s {.*} AS data").
		WithParams(params).
		RunRead()

	if err != nil {
		return nil, fmt.Errorf("failed to get SPK node by ID: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("SPK node with ID %d not found", id)
	}

	dataVal, ok := records[0].Get("data")
	if !ok {
		return nil, fmt.Errorf("field 'data' not found in record")
	}

	props, ok := dataVal.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("data is not a map[string]any")
	}

	return mapToSpkNode(props), nil
}

func mapToSpkNode(props map[string]any) *SpkNode {
	spk := &SpkNode{}

	if idVal, ok := props["id"].(int64); ok {
		spk.ID = idVal
	}

	if nameVal, ok := props["name"].(string); ok {
		spk.Name = nameVal
	}

	if codeVal, ok := props["code"].(string); ok {
		spk.Code = codeVal
	}

	if descVal, ok := props["description"].(string); ok {
		spk.Description = descVal
	}

	if createdVal, ok := props["created_at"].(string); ok {
		spk.CreatedAt = createdVal
	}

	if updatedVal, ok := props["updated_at"].(string); ok {
		spk.UpdatedAt = updatedVal
	}

	return spk
}

func (r *spkRepository) InsertGraphSpk(data *SpkNode) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"id":          data.ID,
		"name":        data.Name,
		"code":        data.Code,
		"description": data.Description,
		"createdAt":   data.CreatedAt,
		"updatedAt":   data.UpdatedAt,
	}

	if err := graph.
		WithMerge("(s:SPK {id: $id})").
		WithSet(`s.name = $name, 
			s.code = $code, 
			s.description = $description,
			s.created_at = datetime($createdAt), 
			s.updated_at = datetime($updatedAt)`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to merge SPK node: %w", err)
	}

	return nil
}

func (r *spkRepository) UpdateGraphSpk(data *SpkNode) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"id":          data.ID,
		"name":        data.Name,
		"code":        data.Code,
		"description": data.Description,
	}

	if err := graph.
		WithMatch("(s:SPK {id: $id})").
		WithSet("s.name = $name, s.code = $code, s.description = $description", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update SPK graph with id %d: %w", data.ID, err)
	}

	return nil
}

func (r *spkRepository) DeleteGraphSpk(spkId int64) error {
	graph := builder.NewGraphRepository()

	params := map[string]interface{}{
		"docId":     spkId,
		"deletedAt": time.Now().Format(time.RFC3339Nano),
	}

	if err := graph.
		WithMatch("(s:SPK {id: $docId})").
		WithSet("s.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to soft delete SPK graph with id %d: %w", spkId, err)
	}

	return nil
}

func (r *spkRepository) BulkInsertGraphSpks(data []*SpkNode) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	spkNodes := make([]map[string]any, 0, len(data))
	for _, spk := range data {
		spkNodes = append(spkNodes, map[string]any{
			"id":          spk.ID,
			"code":        spk.Code,
			"name":        spk.Name,
			"description": spk.Description,
		})
	}

	params := map[string]any{"spks": spkNodes}

	if err := graph.
		WithUnwind("$spks", "s").
		WithMerge("(spk:SPK {id: s.id})").
		WithSet(`spk.code = s.code, spk.name = s.name, spk.description = s.description`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk insert SPK nodes: %w", err)
	}

	return nil
}

func (r *spkRepository) BulkUpdateGraphSpks(data []*SpkNode) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	spkNodes := make([]map[string]any, 0, len(data))
	for _, spk := range data {
		spkNodes = append(spkNodes, map[string]any{
			"id":          spk.ID,
			"code":        spk.Code,
			"name":        spk.Name,
			"description": spk.Description,
		})
	}

	params := map[string]any{"spks": spkNodes}

	if err := graph.
		WithUnwind("$spks", "s").
		WithMatch("(spk:SPK {id: s.id})").
		WithSet("spk.code = s.code, spk.name = s.name, spk.description = s.description", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk update SPK nodes: %w", err)
	}

	return nil
}

func (r *spkRepository) BulkDeleteGraphSpks(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	params := map[string]any{
		"spkIds":    ids,
		"deletedAt": time.Now().Format(time.RFC3339Nano),
	}

	if err := graph.
		WithMatch("(s:SPK)").
		WithWhere("s.id IN $spkIds", nil).
		WithSet("s.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk soft delete SPK nodes: %w", err)
	}

	return nil
}

func (r *spkRepository) CountGraphSpks(filter dto.SpkFilterDto) (int64, error) {
	repo := builder.NewGraphRepository()
	params := make(map[string]any)

	repo = repo.WithMatch("(s:SPK)")

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
		return 0, fmt.Errorf("failed to count SPK nodes: %w", err)
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
