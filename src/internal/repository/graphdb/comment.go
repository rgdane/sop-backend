package graphdb

import (
	"errors"
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/pkg/neo4j/builder"

	"github.com/google/uuid"
)

var (
	ErrNoDocumentGraph = errors.New("no document graph found")
	ErrNoCommentGraph  = fmt.Errorf("no comment graph found")
)

func GetCommentNode(commentId string) (any, error) {
	graph := builder.NewGraphRepository()
	param := map[string]any{
		"commentId": commentId,
	}

	graph = graph.
		WithMatch("(c:Comment)").
		WithWhere("c.commentId = $commentId AND c.deleted_at IS NULL", param).
		WithWith("c").
		WithCall(`
			apoc.path.expandConfig(c, {
				relationshipFilter: ">",
				minLevel: 1,
				maxLevel: 10
			})
		`).
		WithYield("path").
		WithWith("collect(path) AS paths").
		WithUnwind("paths", "p").
		WithWith("p, nodes(p) AS pathNodes").
		WithWhere("all(n IN pathNodes WHERE n.deleted_at IS NULL)", param).
		WithWith("collect(p) AS filteredPaths").
		WithCall("apoc.convert.toTree(filteredPaths)").
		WithYield("value").
		WithReturn("value AS data").
		WithParams(param)

	result, err := graph.RunRead()
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, ErrNoCommentGraph
	}

	data, ok := result[0].Get("data")
	if !ok {
		return nil, fmt.Errorf("field 'data' not found in record")
	}
	return data, nil
}

func CreateCommentNode(elementId string, comment dto.CommentDto) error {
	graph := builder.NewGraphRepository()

	param := map[string]any{
		"elementId": elementId,
		"value":     comment.Text,
		"commentId": comment.Id,
	}

	if err := graph.
		WithMatch("(n)").
		WithWhere("elementId(n) = $elementId", nil).
		WithCreate("(n)-[:HAS_COMMENT]->(c:Comment {value: $value, commentId: $commentId})").
		WithParams(param).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to create comment graph: %w", err)
	}

	return nil
}

func CreateReviewGraph(elementId string, review dto.ReviewDto) error {
	graph := builder.NewGraphRepository()

	if review.Id == "" {
		review.Id = uuid.New().String()
	}

	param := map[string]any{
		"elementId": elementId,
		"value":     review.Text,
		"reviewId":  review.Id,
		"isolated":  review.Isolated,
	}

	if err := graph.
		WithMatch("(n:Comment)").
		WithWhere("n.commentId = $elementId", param).
		WithCreate("(n)-[:HAS_REVIEW]->(r:Review {value: $value, reviewId: $reviewId, isolated: $isolated})").
		WithParams(param).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to create review graph connected to comment: %w", err)
	}

	return nil
}
