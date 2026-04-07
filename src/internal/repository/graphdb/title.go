package graphdb

import (
	"fmt"
	"strings"
	"time"

	"jk-api/api/http/controllers/v1/dto"
	"jk-api/pkg/neo4j/builder"
)

type TitleNode struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	Color     string `json:"color"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type TitleRepository interface {
	GetAllGraphTitles(filter dto.TitleFilterDto) ([]*TitleNode, error)
	GetGraphTitleByID(id int64) (*TitleNode, error)
	InsertGraphTitle(data *TitleNode) error
	UpdateGraphTitle(data *TitleNode) error
	DeleteGraphTitle(titleId int64) error

	BulkInsertGraphTitles(data []*TitleNode) error
	BulkUpdateGraphTitles(data []*TitleNode) error
	BulkDeleteGraphTitles(ids []int64) error

	CountGraphTitles(filter dto.TitleFilterDto) (int64, error)
}

type titleRepository struct{}

func NewTitleRepository() TitleRepository {
	return &titleRepository{}
}

func (r *titleRepository) GetAllGraphTitles(filter dto.TitleFilterDto) ([]*TitleNode, error) {
	repo := builder.NewGraphRepository()
	params := make(map[string]any)

	repo = repo.WithMatch("(t:Title)")

	var conditions []string

	if filter.ShowDeleted {
		conditions = append(conditions, "t.deleted_at IS NOT NULL")
	} else {
		conditions = append(conditions, "t.deleted_at IS NULL")
	}

	if filter.Name != "" {
		conditions = append(conditions, "toLower(t.name) CONTAINS toLower($name)")
		params["name"] = filter.Name
	}

	repo = repo.WithWhere(strings.Join(conditions, " AND "), params)

	returnClause := "t {.*} AS data"

	if filter.Sort != "" && filter.Order != "" {
		orderDir := strings.ToUpper(filter.Order)
		if orderDir != "ASC" && orderDir != "DESC" {
			orderDir = "ASC"
		}

		returnClause += fmt.Sprintf(" ORDER BY t.%s %s", filter.Sort, orderDir)
	}

	repo = repo.
		WithReturn(returnClause).
		WithParams(params)

	records, err := repo.RunRead()
	if err != nil {
		return nil, fmt.Errorf("failed to get titles with filter: %w", err)
	}

	var titles []*TitleNode
	for _, record := range records {
		dataVal, ok := record.Get("data")
		if !ok {
			continue
		}

		props, ok := dataVal.(map[string]any)
		if !ok {
			continue
		}

		title := mapToTitleNode(props)
		titles = append(titles, title)
	}

	return titles, nil
}

func (r *titleRepository) GetGraphTitleByID(id int64) (*TitleNode, error) {
	repo := builder.NewGraphRepository()
	params := map[string]any{
		"id": id,
	}

	records, err := repo.
		WithMatch("(t:Title)").
		WithWhere("t.id = $id AND t.deleted_at IS NULL", params).
		WithReturn("t {.*} AS data").
		WithParams(params).
		RunRead()

	if err != nil {
		return nil, fmt.Errorf("failed to get title node by ID: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("title node with ID %d not found", id)
	}

	dataVal, ok := records[0].Get("data")
	if !ok {
		return nil, fmt.Errorf("field 'data' not found in record")
	}

	props, ok := dataVal.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("data is not a map[string]any")
	}

	return mapToTitleNode(props), nil
}

func mapToTitleNode(props map[string]any) *TitleNode {
	title := &TitleNode{}

	if idVal, ok := props["id"].(int64); ok {
		title.ID = idVal
	}

	if nameVal, ok := props["name"].(string); ok {
		title.Name = nameVal
	}

	if codeVal, ok := props["code"].(string); ok {
		title.Code = codeVal
	}

	if colorVal, ok := props["color"].(string); ok {
		title.Color = colorVal
	}

	if createdVal, ok := props["created_at"].(string); ok {
		title.CreatedAt = createdVal
	}

	if updatedVal, ok := props["updated_at"].(string); ok {
		title.UpdatedAt = updatedVal
	}

	return title
}

func (r *titleRepository) InsertGraphTitle(data *TitleNode) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"id":        data.ID,
		"name":      data.Name,
		"code":      data.Code,
		"color":     data.Color,
		"createdAt": data.CreatedAt,
		"updatedAt": data.UpdatedAt,
	}

	if err := graph.
		WithMerge("(t:Title {id: $id})").
		WithSet(`t.name = $name, 
			t.code = $code, 
			t.color = $color,
			t.created_at = datetime($createdAt), 
			t.updated_at = datetime($updatedAt)`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to merge Title node: %w", err)
	}

	return nil
}

func (r *titleRepository) UpdateGraphTitle(data *TitleNode) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"id":    data.ID,
		"name":  data.Name,
		"code":  data.Code,
		"color": data.Color,
	}

	if err := graph.
		WithMatch("(t:Title {id: $id})").
		WithSet("t.name = $name, t.code = $code, t.color = $color", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update Title graph with id %d: %w", data.ID, err)
	}

	return nil
}

func (r *titleRepository) DeleteGraphTitle(titleId int64) error {
	graph := builder.NewGraphRepository()

	params := map[string]interface{}{
		"docId":     titleId,
		"deletedAt": time.Now().Format(time.RFC3339Nano),
	}

	if err := graph.
		WithMatch("(t:Title {id: $docId})").
		WithSet("t.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to soft delete Title graph with id %d: %w", titleId, err)
	}

	return nil
}

func (r *titleRepository) BulkInsertGraphTitles(data []*TitleNode) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	titleNodes := make([]map[string]any, 0, len(data))
	for _, title := range data {
		titleNodes = append(titleNodes, map[string]any{
			"id":    title.ID,
			"code":  title.Code,
			"name":  title.Name,
			"color": title.Color,
		})
	}

	params := map[string]any{"titles": titleNodes}

	if err := graph.
		WithUnwind("$titles", "title").
		WithMerge("(t:Title {id: title.id})").
		WithSet(`t.code = title.code, t.name = title.name, t.color = title.color`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk insert Title nodes: %w", err)
	}

	return nil
}

func (r *titleRepository) BulkUpdateGraphTitles(data []*TitleNode) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	titleNodes := make([]map[string]any, 0, len(data))
	for _, title := range data {
		titleNodes = append(titleNodes, map[string]any{
			"id":    title.ID,
			"code":  title.Code,
			"name":  title.Name,
			"color": title.Color,
		})
	}

	params := map[string]any{"titles": titleNodes}

	if err := graph.
		WithUnwind("$titles", "title").
		WithMatch("(t:Title {id: title.id})").
		WithSet("t.code = title.code, t.name = title.name, t.color = title.color", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk update Title nodes: %w", err)
	}

	return nil
}

func (r *titleRepository) BulkDeleteGraphTitles(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	params := map[string]any{
		"titleIds":  ids,
		"deletedAt": time.Now().Format(time.RFC3339Nano),
	}

	if err := graph.
		WithMatch("(t:Title)").
		WithWhere("t.id IN $titleIds", nil).
		WithSet("t.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk soft delete Title nodes: %w", err)
	}

	return nil
}

func (r *titleRepository) CountGraphTitles(filter dto.TitleFilterDto) (int64, error) {
	repo := builder.NewGraphRepository()
	params := make(map[string]any)

	repo = repo.WithMatch("(t:Title)")

	var conditions []string

	if filter.ShowDeleted {
		conditions = append(conditions, "t.deleted_at IS NOT NULL")
	} else {
		conditions = append(conditions, "t.deleted_at IS NULL")
	}

	if filter.Name != "" {
		conditions = append(conditions, "toLower(t.name) CONTAINS toLower($name)")
		params["name"] = filter.Name
	}

	repo = repo.
		WithWhere(strings.Join(conditions, " AND "), params).
		WithReturn("count(t) AS total").
		WithParams(params)

	records, err := repo.RunRead()
	if err != nil {
		return 0, fmt.Errorf("failed to count title nodes: %w", err)
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
