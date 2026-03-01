package graphdb

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/pkg/neo4j/builder"
	"maps"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
)

func CreateGraph(payload dto.GraphNode) (*neo4j.Record, error) {
	params := map[string]any{}
	maps.Copy(params, payload.Node.Props)
	graph := builder.NewGraphRepository()

	// labels
	labels := ""
	if len(payload.Node.Labels) > 0 {
		labels = ":" + strings.Join(payload.Node.Labels, ":")
	}

	// props untuk MERGE pattern (identifier)
	propsQuery := []string{}
	for k := range payload.Node.Props {
		propsQuery = append(propsQuery, fmt.Sprintf("%s: $%s", k, k))
	}
	propsStr := strings.Join(propsQuery, ", ")

	// CASE 1: MERGE WITH relationship
	if payload.ElementId != "" && payload.Node.Relationship != "" {
		params["elementId"] = payload.ElementId

		// SET clause dengan alias j (untuk relationship case)
		setQuery := []string{}
		for k := range payload.Node.Props {
			setQuery = append(setQuery, fmt.Sprintf("j.%s = $%s", k, k))
		}
		setStr := strings.Join(setQuery, ", ")

		records, err := graph.
			WithMatch("(n)").
			WithWhere("elementId(n) = $elementId", params).
			WithMerge(fmt.Sprintf("(n)-[:%s]->(j%s {%s})", payload.Node.Relationship, labels, propsStr)).
			WithSet(setStr, params).
			WithParams(params).
			WithReturn("j").
			RunWriteWithReturn()
		if err != nil {
			return nil, fmt.Errorf("failed to merge node: %w", err)
		}
		if len(records) == 0 {
			return nil, fmt.Errorf("no record returned")
		}

		if payload.Node.Relationship == "HAS_DOCUMENT" {
			params["sopId"] = payload.Node.Props["sopId"]
			params["docId"], err = ExtractElementId(records[0].Values[0])
			if err != nil {
				return nil, err
			}
			err := graph.
				WithMatch("(d:Document)").
				WithWhere("elementId(d) = $docId", params).
				WithMatch("(s:SOP)").
				WithWhere("elementId(s) = $sopId", params).
				WithMerge("(d)-[:HAS_SOP]->(s)").
				WithReturn("d, s").
				RunWrite()
			if err != nil {
				return nil, fmt.Errorf("failed to create HAS_DOCUMENT relationship: %w", err)
			}
		}
		return &records[0], nil
	}

	// CASE 2: MERGE node only
	// SET clause dengan alias n (untuk standalone node case)
	setQueryN := []string{}
	for k := range payload.Node.Props {
		setQueryN = append(setQueryN, fmt.Sprintf("n.%s = $%s", k, k))
	}
	setStrN := strings.Join(setQueryN, ", ")

	records, err := graph.
		WithMerge(fmt.Sprintf("(n%s {%s})", labels, propsStr)).
		WithSet(setStrN, params).
		WithParams(params).
		WithReturn("n").
		RunWriteWithReturn()
	if err != nil {
		return nil, fmt.Errorf("failed to merge node: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("no record returned")
	}

	return &records[0], nil
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
