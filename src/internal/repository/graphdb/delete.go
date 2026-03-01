package graphdb

import (
	"fmt"
	"jk-api/pkg/neo4j/builder"
	"time"
)

func RemoveNodes(elementIds []string) error {
	if len(elementIds) == 0 {
		return nil
	}

	repo := builder.NewGraphRepository()
	params := map[string]any{
		"elementIds": elementIds,
		"deletedAt":  time.Now().Unix(),
	}

	err := repo.
		WithUnwind("$elementIds", "eid").
		WithMatch("(n)").
		WithWhere("elementId(n) = eid", nil).
		WithOptionalMatch("(n)-[:HAS_SOP]->(sop)").
		WithWith("collect(n) AS nodes, collect(sop) AS sops").
		WithWith("apoc.coll.subtract(nodes, sops) AS toDelete").
		WithUnwind("toDelete", "del").
		WithSet("del.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWrite()
	if err != nil {
		return fmt.Errorf("failed to bulk delete nodes: %w", err)
	}
	return nil
}
