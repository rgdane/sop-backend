package graphdb

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/pkg/neo4j/builder"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func GetNodeById(elementId string, filter dto.GraphFilterDto, param map[string]any) (map[string]any, error) {
	repo := builder.NewGraphRepository()

	repo = repo.
		WithMatch("(a)").
		WithWhere("elementId(a) = $elementId AND a.deleted_at IS NULL", param).
		WithWith("a").
		WithCall(`
			apoc.path.expandConfig(a, {
				relationshipFilter: ">",
				minLevel: 0,
				maxLevel: 10
			})
		`).
		WithYield("path").
		WithWith("collect(path) AS paths").
		WithUnwind("paths", "p").
		WithWith(`
			p,
			nodes(p) AS pathNodes,
    	[n IN nodes(p) WHERE n:Document][0] AS rootDoc
		`)

	var conditions []string

	conditions = append(conditions, `
		all(n IN pathNodes WHERE n.deleted_at IS NULL)
	`)

	if filter.ProductId != 0 {
		conditions = append(conditions, `
			all(n IN pathNodes WHERE
				NOT (n:Row OR n:Value)
				OR n.productId = $productId
			)
		`)
	}

	if filter.ProjectId != 0 {
		conditions = append(conditions, `
			all(n IN pathNodes WHERE
				NOT (n:Row OR n:Value)
				OR n.projectId = $projectId
			)
		`)
	}

	if filter.EpicId != 0 {
		conditions = append(conditions, `
			all(n IN pathNodes WHERE
				NOT (n:Row OR n:Value)
				OR n.epicId = $epicId
			)
		`)
	}

	if filter.FeatureId != 0 {
		conditions = append(conditions, `
			all(n IN pathNodes WHERE
				NOT (n:Row OR n:Value)
				OR n.featureId = $featureId
			)
		`)
	}

	if filter.DocumentId != "" {
		conditions = append(conditions, `
			all(n IN pathNodes WHERE
				NOT (n:Row OR n:Value)
				OR n.documentId = $filterDocumentId
			)
		`)
	}

	if filter.FilterId != "" {
		conditions = append(conditions, `
			all(n IN pathNodes WHERE
				NOT (n:Row OR n:Value)
				OR n.filterId = $filterId
			)
		`)
	}

	repo = repo.
		WithWhere(strings.Join(conditions, " AND "), param).
		WithWith("collect(p) AS filteredPaths").
		WithCall("apoc.convert.toTree(filteredPaths)").
		WithYield("value").
		WithReturn("value AS data").
		WithParams(param)

	result, err := repo.RunRead()
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("graph not found for elementId %s", elementId)
	}

	data, ok := result[0].Get("data")
	if !ok {
		return nil, fmt.Errorf("field 'data' not found in record")
	}

	dataTyped, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("data is not a map[string]any")
	}

	return dataTyped, nil
}

func GetNodeByLabel(label string, filter dto.GraphFilterDto, params map[string]any) ([]neo4j.Record, error) {
	repo := builder.NewGraphRepository()
	where := []string{"a.deleted_at IS NULL"}

	// -----------------------------
	// #1 Filter awal (MATCH a)
	// -----------------------------
	if filter.Name != "" {
		params["name"] = filter.Name
		where = append(where, "a.name = $name")
	}

	repo = repo.
		WithMatch(fmt.Sprintf("(a:%s)", label)).
		WithWhere(strings.Join(where, " AND "), params).
		WithCall(`apoc.path.expandConfig(a, {
            relationshipFilter: ">",
            minLevel: 0,
            maxLevel: -1
        }) YIELD path`).
		WithWith("a, collect(path) AS paths")

	// -----------------------------
	// #2 Filter PATHS
	// -----------------------------
	var pathCond []string

	// Semua node di path harus non-deleted
	pathCond = append(pathCond, `
        all(n IN nodes(p) WHERE n.deleted_at IS NULL)
    `)

	if filter.ProductId != 0 {
		params["productId"] = filter.ProductId
		pathCond = append(pathCond, `
            all(n IN nodes(p) WHERE 
                NOT (n:Row OR n:Value) OR 
                n.productId IS NULL OR 
                n.productId = $productId
            )
        `)
	}
	if filter.ProjectId != 0 {
		params["projectId"] = filter.ProjectId
		pathCond = append(pathCond, `
            all(n IN nodes(p) WHERE 
                NOT (n:Row OR n:Value) OR 
                n.projectId IS NULL OR 
                n.projectId = $projectId
            )
        `)
	}

	// -----------------------------
	// #3 UNWIND paths + WITH p, a (PERBAIKAN PENTING)
	// -----------------------------
	repo = repo.
		WithUnwind("paths", "p").
		WithWith("a, p"). // <- FIX untuk menghindari SyntaxError
		WithWhere(strings.Join(pathCond, " AND "), params).
		WithWith("a, collect(p) AS filteredPaths").
		WithWith(`
            a,
            CASE WHEN size(filteredPaths) = 0 
                THEN [a] 
                ELSE filteredPaths 
            END AS finalPaths
        `)

	// -----------------------------
	// #4 Convert to tree
	// -----------------------------
	records, err := repo.
		WithCall("apoc.convert.toTree(finalPaths) YIELD value").
		WithReturn("value AS data").
		WithParams(params).
		RunRead()
	if err != nil {
		return nil, fmt.Errorf("failed to get graph by label: %w", err)
	}

	return records, nil
}

func GetNodeByProps(where []string, params map[string]any) ([]map[string]any, error) {
	repo := builder.NewGraphRepository()

	repo = repo.
		WithMatch("(data)").
		WithWhere(strings.Join(where, " AND "), params).
		WithReturn(`
			apoc.map.fromPairs(
				[key IN keys(data) |
				[
					CASE key WHEN 'commentId' THEN 'comment_id'
					ELSE apoc.text.camelCase(key) END,
					data[key]
				]
				]
			) AS data
		`)

	records, err := repo.RunRead()
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes by props: %w", err)
	}

	var results []map[string]any
	for _, r := range records {
		results = append(results, r.AsMap())
	}
	return results, nil
}

func GetDocumentNode(documentId string, filter dto.GraphFilterDto) (any, error) {
	repo := builder.NewGraphRepository()
	param := map[string]any{
		"documentId": documentId,
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

	repo = repo.
		WithMatch("(d:Document)").
		WithWhere("d.id = toInteger($documentId) AND d.deleted_at IS NULL", param).
		WithWith("d").
		WithCall(`
			apoc.path.expandConfig(d, {
				relationshipFilter: ">",
				minLevel: 0,
				maxLevel: 10
			})
		`).
		WithYield("path").
		WithWith("collect(path) AS paths").
		WithUnwind("paths", "p").
		WithWith(`
			p,
			nodes(p) AS pathNodes,
			[n IN nodes(p) WHERE n:Column AND n.index = 0] AS indexZeroColumns,
			[r IN relationships(p) WHERE type(r) = 'HAS_COLUMN'] AS hasColumnRels,
			[n IN nodes(p) WHERE n:Document][0] AS rootDoc
		`)

	var conditions []string

	conditions = append(conditions, `
		NOT any(col IN indexZeroColumns WHERE 
			any(rel IN hasColumnRels WHERE endNode(rel) = col)
		)
	`)

	conditions = append(conditions, `
		all(n IN pathNodes WHERE n.deleted_at IS NULL)
	`)

	// Tentukan node types yang akan difilter berdasarkan flag Multiple
	var filterNodeTypes string
	if filter.Multiple {
		// Filter Row, Value, DAN Document child
		filterNodeTypes = "n:Row OR n:Value OR (n:Document AND n <> rootDoc)"
	} else {
		// Filter Row dan Value saja (exclude Document child)
		filterNodeTypes = "n:Row OR n:Value"
	}

	if filter.ProductId != 0 {
		conditions = append(conditions, fmt.Sprintf(`
			all(n IN pathNodes WHERE 
				NOT (%s)
				OR n.productId = $productId
			)
		`, filterNodeTypes))
	}

	if filter.FeatureId != 0 {
		conditions = append(conditions, fmt.Sprintf(`
			all(n IN pathNodes WHERE 
				NOT (%s)
				OR n.featureId = $featureId
			)
		`, filterNodeTypes))
	}

	if filter.EpicId != 0 {
		conditions = append(conditions, fmt.Sprintf(`
			all(n IN pathNodes WHERE 
				NOT (%s)
				OR n.epicId = $epicId
			)
		`, filterNodeTypes))
	}

	if filter.ProjectId != 0 {
		conditions = append(conditions, fmt.Sprintf(`
			all(n IN pathNodes WHERE 
				NOT (%s)
				OR n.projectId = $projectId
			)
		`, filterNodeTypes))
	}

	if filter.DocumentId != "" {
		conditions = append(conditions, fmt.Sprintf(`
			all(n IN pathNodes WHERE 
				NOT (%s)
				OR n.documentId <> toString(rootDoc.id)
				OR n.documentId = $filterDocumentId
			)
		`, filterNodeTypes))
	}

	repo = repo.
		WithWhere(strings.Join(conditions, " AND "), param).
		WithWith("collect(p) AS filteredPaths").
		WithCall("apoc.convert.toTree(filteredPaths)").
		WithYield("value").
		WithReturn("value AS data").
		WithParams(param)

	result, err := repo.RunRead()
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no document graph found")
	}
	data, ok := result[0].Get("data")
	if !ok {
		return nil, fmt.Errorf("field 'data' not found in record")
	}
	return data, nil
}
