package service

import (
	"errors"
	"fmt"

	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/repository/graphdb"
	"jk-api/pkg/neo4j/builder"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
)

var (
	ErrNoDocumentGraph = errors.New("no document graph found")
	ErrNoCommentGraph  = fmt.Errorf("no comment graph found")
)

func GetGraphById(elementId string, filter dto.GraphFilterDto) (any, error) {
	param := map[string]any{
		"elementId": elementId,
	}

	if filter.ProductId != 0 {
		param["productId"] = filter.ProductId
	}
	if filter.ProjectId != 0 {
		param["projectId"] = filter.ProjectId
	}
	if filter.EpicId != 0 {
		param["epicId"] = filter.EpicId
	}
	if filter.FeatureId != 0 {
		param["featureId"] = filter.FeatureId
	}
	if filter.DocumentId != "" {
		param["filterDocumentId"] = filter.DocumentId
	}
	if filter.FilterId != "" {
		param["filterId"] = filter.FilterId
	}

	data, err := graphdb.GetNodeById(elementId, filter, param)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph by id: %w", err)
	}

	return data, nil
}

func GetGraphByLabel(label string, filter dto.GraphFilterDto) ([]map[string]any, error) {
	params := map[string]any{}

	if filter.Name != "" {
		params["name"] = filter.Name
	}

	records, err := graphdb.GetNodeByLabel(label, filter, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph by label: %w", err)
	}

	var datas []map[string]any
	for _, r := range records {
		datas = append(datas, r.AsMap())
	}

	return datas, nil
}

func GetGraphByProps(filter dto.GraphFilterDto) (any, error) {
	params := map[string]any{}
	where := []string{"a.deleted_at IS NULL"}

	if filter.CommentId != "" {
		params["commentId"] = filter.CommentId
		where = append(where, "a.name = $name")
	}

	data, err := graphdb.GetNodeByProps(where, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get graph by props: %w", err)
	}

	return data, nil
}

func GetDocumentGraph(documentId string, filter dto.GraphFilterDto) (any, error) {
	data, err := graphdb.GetDocumentNode(documentId, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get document graph: %w", err)
	}

	return data, nil
}

func CreateTableGraph(elementId string, relation string, payload dto.ColumnDto) error {
	err := graphdb.CreateTableNode(elementId, relation, payload)
	if err != nil {
		return fmt.Errorf("failed to create table graph: %w", err)
	}

	return nil
}

func GetCommentGraph(commentId string) (any, error) {
	data, error := graphdb.GetCommentNode(commentId)
	if error != nil {
		return nil, fmt.Errorf("failed to get comment graph: %w", error)
	}

	return data, nil
}

func CreateTextGraph(elementId string, payload dto.TextDto) error {
	graph := builder.NewGraphRepository()

	param := map[string]any{
		"elementId":  elementId,
		"value":      payload.Value,
		"name":       "text value",
		"linkGPT":    payload.LinkGPT,
		"nodeType":   "text",
		"productId":  payload.ProductId,
		"projectId":  payload.ProjectId,
		"epicId":     payload.EpicId,
		"featureId":  payload.FeatureId,
		"documentId": payload.DocumentId,
		"filterId":   payload.FilterId,
	}

	mergeProps := "documentId: $documentId"
	if payload.ProductId > 0 {
		mergeProps += ", productId: $productId"
	}
	if payload.ProjectId > 0 {
		mergeProps += ", projectId: $projectId"
	}
	if payload.EpicId > 0 {
		mergeProps += ", epicId: $epicId"
	}
	if payload.FeatureId > 0 {
		mergeProps += ", featureId: $featureId"
	}
	if payload.FilterId != "" {
		mergeProps += ", filterId: $filterId"
	}

	setString := `r.value = $value,
			r.name = $name,
			r.linkGPT = $linkGPT,
			r.documentId = $documentId,
			r.nodeType = $nodeType`
	if payload.ProductId > 0 {
		setString += ", r.productId = $productId"
	}
	if payload.ProjectId > 0 {
		setString += ", r.projectId = $projectId"
	}
	if payload.EpicId > 0 {
		setString += ", r.epicId = $epicId"
	}
	if payload.FeatureId > 0 {
		setString += ", r.featureId = $featureId"
	}
	if payload.FilterId != "" {
		setString += ", r.filterId = $filterId"
	}

	graph = graph.
		WithMatch("(n)").
		WithWhere("elementId(n) = $elementId", param).
		WithSet("n.nodeType = $nodeType", param).
		WithMerge(fmt.Sprintf("(n)-[:HAS_VALUE]->(r:Row {%s})", mergeProps)).
		WithSet(setString, param).
		WithParams(param)

	if err := graph.RunWrite(); err != nil {
		return fmt.Errorf("failed to create text graph: %w", err)
	}

	return nil
}

func CreateCommentGraph(elementId string, comment dto.CommentDto) error {
	err := graphdb.CreateCommentNode(elementId, comment)
	if err != nil {
		return fmt.Errorf("failed to create comment graph: %w", err)
	}

	return nil
}

func CreateGraph(payload dto.GraphNode) (*neo4j.Record, error) {
	data, err := graphdb.CreateGraph(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create graph: %w", err)
	}

	return data, nil
}

func BulkCreateGraph(payload []dto.GraphNode) error {
	return nil
}
func CreateReviewGraph(elementId string, review dto.ReviewDto) error {
	err := graphdb.CreateReviewGraph(elementId, review)
	if err != nil {
		return fmt.Errorf("failed to create review graph: %w", err)
	}

	return nil
}

func UpdateGraph(elementId string, payload dto.NodeData) error {
	err := graphdb.UpdateGraph(elementId, payload)
	if err != nil {
		return fmt.Errorf("failed to update graph: %w", err)
	}

	return nil
}

func UpdateMultipleGraph(payload []dto.NodeData) error {
	err := graphdb.UpdateMultipleGraph(payload)
	if err != nil {
		return fmt.Errorf("failed to update graph: %w", err)
	}

	return nil
}

func MergeGraph(payload dto.GraphNode) error {
	err := graphdb.MergeGraph(payload)
	if err != nil {
		return fmt.Errorf("failed to merge graph: %w", err)
	}

	return nil
}

func UpdateTableGraph(elementId string, relation string, payload dto.ColumnDto) error {
	err := graphdb.UpdateTableGraph(elementId, relation, payload)
	if err != nil {
		return fmt.Errorf("failed to update table graph: %w", err)
	}

	return nil
}

func DeleteGraph(elementIds []string) error {
	err := graphdb.RemoveNodes(elementIds)

	if err != nil {
		return fmt.Errorf("failed to bulk delete nodes: %w", err)
	}

	return nil
}

func ExtractElementId(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case dbtype.Node:
		return v.ElementId, nil
	default:
		return "", fmt.Errorf("unexpected type for elementId: %T", value)
	}
}
