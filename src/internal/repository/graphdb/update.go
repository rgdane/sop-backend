package graphdb

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/pkg/neo4j/builder"
	"strings"
)

func UpdateGraph(elementId string, payload dto.NodeData) error {
	graph := builder.NewGraphRepository()
	params := map[string]any{
		"elementId": elementId,
	}

	setClauses := []string{}

	for key, value := range payload.Props {
		if value != nil {
			paramKey := key
			params[paramKey] = value
			setClauses = append(setClauses, fmt.Sprintf("n.%s = $%s", key, paramKey))
		}
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setClause := strings.Join(setClauses, ", ")

	if err := graph.
		WithMatch("(n)").
		WithWhere("elementId(n) = $elementId", params).
		WithSet(setClause, params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update graph node: %w", err)
	}

	return nil
}

func MergeGraph(payload dto.GraphNode) error {
	graph := builder.NewGraphRepository()

	if payload.ElementId == "" {
		return fmt.Errorf("parent elementId is required")
	}
	if len(payload.Node.Labels) == 0 {
		return fmt.Errorf("node labels are required")
	}
	if payload.Node.Relationship == "" {
		return fmt.Errorf("relationship is required")
	}

	params := map[string]any{
		"parentElementId": payload.ElementId,
	}

	labelStr := strings.Join(payload.Node.Labels, ":")

	// build SET child.x = $child_x
	setProps := []string{}
	for k, v := range payload.Node.Props {
		if v != nil {
			p := "child_" + k
			params[p] = v
			setProps = append(setProps, fmt.Sprintf("child.%s = $%s", k, p))
		}
	}
	setClause := strings.Join(setProps, ", ")

	oldRel := fmt.Sprintf("(parent)-[old:%s]->(oldChild)", payload.Node.Relationship)
	newRel := fmt.Sprintf("(parent)-[:%s]->(child)", payload.Node.Relationship)

	q := graph.
		WithMatch("(parent)").
		WithWhere("elementId(parent) = $parentElementId", params).
		WithOptionalMatch(oldRel).
		WithDelete("old")
	// --------------------------
	// CASE 1: child by elementId
	// --------------------------
	if payload.Node.ElementId != "" {
		params["childElementId"] = payload.Node.ElementId
		childMatch := fmt.Sprintf("(child:%s)", labelStr)

		q = q.
			WithMatch(childMatch).
			WithWhere("elementId(child) = $childElementId", params)
	} else {
		// --------------------------
		// CASE 2: child by props.id
		// --------------------------
		id := payload.Node.Props["id"]
		if id == nil {
			return fmt.Errorf("either elementId or props.id is required for child node")
		}

		params["childId"] = id
		childMerge := fmt.Sprintf("(child:%s {id: $childId})", labelStr)

		q = q.WithMerge(childMerge)
	}

	if setClause != "" {
		q = q.WithSet(setClause, params)
	}

	err := q.
		WithMerge(newRel).
		RunWrite()

	if err != nil {
		return fmt.Errorf("failed to merge graph: %w", err)
	}

	return nil
}

func UpdateMultipleGraph(payload []dto.NodeData) error {
	graph := builder.NewGraphRepository()

	nodes := make([]map[string]any, len(payload))
	for i, node := range payload {
		nodes[i] = map[string]any{
			"elementId": node.ElementId,
			"props":     node.Props,
		}
	}

	params := map[string]any{
		"nodes": nodes,
	}

	if err := graph.
		WithUnwind("$nodes", "n").
		WithMatch("(x)").                             // match semua Node
		WithWhere("elementId(x) = n.elementId", nil). // filter by elementId
		WithSet("x += n.props", nil).                 // update props
		WithParams(params).
		RunWrite(); err != nil {
		return err
	}

	return nil
}
