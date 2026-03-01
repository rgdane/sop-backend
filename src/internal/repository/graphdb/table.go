package graphdb

import (
	"encoding/json"
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/shared/helper"
	"jk-api/pkg/neo4j/builder"
)

func CreateTableNode(elementId string, relation string, payload dto.ColumnDto) error {
	graph := builder.NewGraphRepository()
	nodeType := payload.NodeType

	targetNodeParam := map[string]any{
		"elementId": elementId,
		"nodeType":  nodeType,
	}

	if err := graph.
		WithMatch("(n)").
		WithWhere("elementId(n) = $elementId", targetNodeParam).
		WithSet("n.nodeType = $nodeType", targetNodeParam).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update nodeType: %w", err)
	}

	if nodeType != "table" {
		return nil
	}

	switch relation {
	case "row":
		for _, col := range payload.Rows {
			colParam := map[string]any{
				"parentId":   col.ElementId,
				"value":      col.Value,
				"nodeType":   col.NodeType,
				"rowIndex":   col.RowIndex,
				"groupIndex": col.GroupIndex,
				"url":        col.Url,
				"productId":  payload.ProductId,
				"documentId": payload.DocumentId,
			}

			query := `(parent)-[:HAS_ROW]->(r:Row {
				name: $value,
				value: $value,
				nodeType: $nodeType,
				rowIndex: $rowIndex,
				groupIndex: $groupIndex,
				url: $url,
				productId: $productId,
				documentId: $documentId,
			})`

			if err := graph.
				WithMatch("(parent)").
				WithWhere("elementId(parent) = $parentId", colParam).
				WithMerge(query).
				WithParams(colParam).
				RunWrite(); err != nil {
				return fmt.Errorf("failed to merge row '%s': %w", col.Value, err)
			}
		}

	default:
		return fmt.Errorf("unsupported relation type: %s", relation)
	}

	return nil
}

func UpdateTableGraph(elementId string, relation string, payload dto.ColumnDto) error {
	graph := builder.NewGraphRepository()

	nodeType := payload.NodeType
	tableRef := payload.TableRef

	targetNodeParam := map[string]any{
		"elementId": elementId,
		"nodeType":  nodeType,
		"tableRef":  tableRef,
		"titleId":   payload.TitleId,
	}

	// Update nodeType pada target node
	if err := graph.
		WithMatch("(n)").
		WithWhere("elementId(n) = $elementId", targetNodeParam).
		WithSet("n.nodeType = $nodeType, n.tableRef = $tableRef, n.title_id = $titleId", targetNodeParam).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update nodeType: %w", err)
	}

	if nodeType == "canvas" && relation == "job" {
		return handleCanvasJob(elementId, payload)
	}

	if nodeType != "table" && nodeType != "database" {
		return nil
	}

	switch relation {
	case "column":
		return handleColumnRelation(elementId, payload)
	case "row":
		return handleRowRelation(payload)
	case "job":
		return handleJobRelation(elementId, payload)
	default:
		return nil
	}
}

func handleCanvasJob(elementId string, payload dto.ColumnDto) error {
	return handleJobRelation(elementId, payload)
}

// Refactor existing column logic ke function terpisah
func handleColumnRelation(elementId string, payload dto.ColumnDto) error {
	graph := builder.NewGraphRepository()
	// Step 1: Ambil existing columns dengan index mereka
	existingColumnsParams := map[string]any{
		"elementId": elementId,
	}
	existingRecords, err := graph.
		WithMatch("(t)").
		WithWhere("elementId(t) = $elementId", existingColumnsParams).
		WithOptionalMatch("(t)-[:HAS_COLUMN]->(c:Column)").
		WithReturn("collect({elementId: elementId(c), index: c.index}) AS columns").
		WithParams(existingColumnsParams).
		RunRead()
	if err != nil {
		return fmt.Errorf("failed to get existing columns: %w", err)
	}

	// Parse existing columns
	existingColumns := make(map[int]string) // index -> elementId
	if len(existingRecords) > 0 {
		if cols, ok := existingRecords[0].Get("columns"); ok {
			if colList, ok := cols.([]any); ok {
				for _, col := range colList {
					if colMap, ok := col.(map[string]any); ok {
						if idx, ok := colMap["index"].(int64); ok {
							if elemId, ok := colMap["elementId"].(string); ok && elemId != "" {
								existingColumns[int(idx)] = elemId
							}
						}
					}
				}
			}
		}
	}

	// Step 2: Track which indexes are in the new payload
	newIndexes := make(map[int]bool)
	for _, col := range payload.HasColumn {
		newIndexes[col.Index] = true
	}

	// Step 3: Delete columns that are NOT in the new payload
	for idx, elemId := range existingColumns {
		if !newIndexes[idx] {
			deleteParams := map[string]any{
				"elementId": elemId,
			}

			err := graph.
				WithMatch("(c:Column)").
				WithWhere("elementId(c) = $elementId", deleteParams).
				WithOptionalMatch("(c)-[r:HAS_ROW]->(row:Row)").
				WithDelete("r").
				WithWith("c").
				WithDetachDelete("c").
				WithParams(deleteParams).
				RunWrite()
			if err != nil {
				return fmt.Errorf("failed to delete old column at index %d: %w", idx, err)
			}
		}
	}

	// Step 4: Update atau create columns dari payload
	for _, col := range payload.HasColumn {
		colParam := map[string]any{
			"parentId": elementId,
			"name":     col.Name,
			"nodeType": col.NodeType,
			"options":  helper.ToJSONString(col.Options),
			"tableRef": col.TableRef,
			"index":    col.Index,
			"titleId":  col.TitleId,
		}

		records, err := graph.
			WithMatch("(parent)").
			WithWhere("elementId(parent) = $parentId", colParam).
			WithMerge("(parent)-[:HAS_COLUMN]->(c:Column {index: $index})").
			WithSet(`c.name = $name,
				c.title_id = $titleId,
				c.nodeType = $nodeType,
				c.options = $options,
				c.tableRef = $tableRef`, colParam).
			WithParams(colParam).
			WithReturn("elementId(c) AS columnId").
			RunWriteWithReturn()
		if err != nil {
			return fmt.Errorf("failed to merge column '%s': %w", col.Name, err)
		}

		if len(records) == 0 {
			return fmt.Errorf("no record returned for column '%s'", col.Name)
		}

		if len(col.HasTable) != 0 {
			nestedElementId, err := insertNestedColumn(col)
			if err != nil {
				return fmt.Errorf("failed to create nested table: %w", err)
			}

			updateParam := map[string]any{
				"columnId": col.ElementId,
				"graphRef": nestedElementId,
			}

			if err := graph.
				WithMatch("(c:Column)").
				WithWhere("elementId(c) = $columnId", updateParam).
				WithSet("c.graphRef = $graphRef", updateParam).
				WithParams(updateParam).
				RunWrite(); err != nil {
				return fmt.Errorf("failed to update column graphRef: %w", err)
			}
		}

		if len(col.HasJob) != 0 {
			fmt.Println("Inserting Job Graph for Column")
			_, err := insertJobGraph(col)
			if err != nil {
				return fmt.Errorf("failed to create job graph: %w", err)
			}
		}
	}

	return nil
}

// Refactor existing row logic ke function terpisah
func handleRowRelation(payload dto.ColumnDto) error {
	graph := builder.NewGraphRepository()
	for _, row := range payload.Rows {
		var valueString string

		switch v := row.Value.(type) {
		case string:
			valueString = v
		case []byte:
			valueString = string(v)
		default:
			jsonValue, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("failed to marshal row.Value: %w", err)
			}
			valueString = string(jsonValue)
		}

		rowParam := map[string]any{
			"value":      valueString,
			"nodeType":   row.NodeType,
			"rowIndex":   row.RowIndex,
			"groupIndex": row.GroupIndex,
			"url":        row.Url,
			"productId":  payload.ProductId,
			"projectId":  payload.ProjectId,
			"epicId":     payload.EpicId,
			"featureId":  payload.FeatureId,
			"documentId": payload.DocumentId,
		}

		mergeProps := "rowIndex: $rowIndex"
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

		setString := `
				r.name = $value,
                r.value = $value,
                r.nodeType = $nodeType,
                r.url = $url,
                r.documentId = $documentId,
				 `
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

		if err := graph.
			WithMatch("(column)").
			WithWhere("elementId(column) = $columnId", rowParam).
			WithMerge(fmt.Sprintf("(column)-[:HAS_ROW]->(r:Row {%s})", mergeProps)).
			WithSet(setString, rowParam).
			WithParams(rowParam).
			RunWrite(); err != nil {
			return fmt.Errorf("failed to merge row '%v': %w", row.Value, err)
		}
	}

	return nil
}

func insertNestedColumn(column dto.ColumnDto) (string, error) {
	nodeData := dto.NodeData{
		Labels: []string{"Table"},
		Props: map[string]any{
			"name": fmt.Sprintf("%s Table", column.Name),
		},
		Relationship: "HAS_TABLE",
	}
	data := dto.GraphNode{
		ElementId: column.ElementId,
		Node:      nodeData,
	}

	record, err := CreateGraph(data)
	if err != nil {
		return "", fmt.Errorf("failed to create/update nested table: %w", err)
	}
	if record == nil {
		return "", fmt.Errorf("no record returned from CreateGraph")
	}

	// Extract table elementId
	tableElementId, err := ExtractElementId(record.Values[0])
	if err != nil {
		return "", err
	}

	//  Insert nested columns dengan CreateGraph (hanya 1 level)
	for _, nestedCol := range column.HasTable {
		fmt.Println("nestedCol:", nestedCol.Name)
		nestedColData := dto.NodeData{
			Labels: []string{"Column"},
			Props: map[string]any{
				"name":     nestedCol.Name,
				"nodeType": nestedCol.NodeType,
				"options":  helper.ToJSONString(nestedCol.Options),
				"tableRef": nestedCol.TableRef,
				"index":    nestedCol.Index,
			},
			Relationship: "HAS_COLUMN",
		}
		nestedColGraph := dto.GraphNode{
			ElementId: tableElementId,
			Node:      nestedColData,
		}

		_, err := CreateGraph(nestedColGraph)

		if nestedCol.NodeType == "table" {
			fmt.Printf("[TABLE 2] : Nested table in column: %s\n", nestedCol.Name)
			insertNestedColumn(nestedCol)
		}

		if nestedCol.NodeType == "canvas" {
			fmt.Println("[CANVAS 2] : Nested canvas detected, inserting job graph...")
			_, err := insertJobGraph(nestedCol)
			if err != nil {
				return "", fmt.Errorf("failed to create job graph for nested canvas column '%s': %w", nestedCol.Name, err)
			}
		}

		if err != nil {
			return "", fmt.Errorf("failed to create/update nested column '%s': %w", nestedCol.Name, err)
		}
	}

	return tableElementId, nil
}

func handleJobRelation(elementId string, payload dto.ColumnDto) error {
	graph := builder.NewGraphRepository()
	if len(payload.HasJob) == 0 {
		return nil
	}

	// Step 1: Ambil existing nested jobs
	existingJobsParams := map[string]any{
		"elementId": elementId,
	}
	existingJobRecords, err := graph.
		WithMatch("(parent)").
		WithWhere("elementId(parent) = $elementId", existingJobsParams).
		WithOptionalMatch("(parent)-[:HAS_JOB]->(j:Job)").
		WithReturn("collect({elementId: elementId(j), index: j.index}) AS jobs").
		WithParams(existingJobsParams).
		RunRead()
	if err != nil {
		return fmt.Errorf("failed to get existing nested jobs: %w", err)
	}

	// Parse existing nested jobs
	existingJobs := make(map[int]string) // index -> elementId
	if len(existingJobRecords) > 0 {
		if jobs, ok := existingJobRecords[0].Get("jobs"); ok {
			if jobList, ok := jobs.([]any); ok {
				for _, job := range jobList {
					if jobMap, ok := job.(map[string]any); ok {
						if idx, ok := jobMap["index"].(int64); ok {
							if elemId, ok := jobMap["elementId"].(string); ok && elemId != "" {
								existingJobs[int(idx)] = elemId
							}
						}
					}
				}
			}
		}
	}

	// Step 2: Track which indexes are in the new payload
	newJobIndexes := make(map[int]bool)
	for _, job := range payload.HasJob {
		newJobIndexes[job.Index] = true
	}

	// Step 3: Delete nested jobs that are NOT in the new payload (including all nested children)
	for idx, elemId := range existingJobs {
		if !newJobIndexes[idx] {
			deleteJobParams := map[string]any{
				"jobId": elemId,
			}

			// Delete job and all its nested children (jobs, columns, etc.) recursively
			err := graph.
				WithMatch("(j:Job)").
				WithWhere("elementId(j) = $jobId", deleteJobParams).
				WithOptionalMatch("(j)-[*]->(child)").
				WithDetachDelete("j, child").
				WithParams(deleteJobParams).
				RunWrite()
			if err != nil {
				return fmt.Errorf("failed to delete old nested job at index %d: %w", idx, err)
			}
			fmt.Printf("Deleted nested job at index %d with all children\n", idx)
		}
	}

	// Step 4: Create or update nested jobs from payload
	for _, job := range payload.HasJob {
		jobParam := map[string]any{
			"parentId": elementId,
			"name":     job.Name,
			"nodeType": job.NodeType,
			"index":    job.Index,
		}

		// Add optional fields if they exist
		if job.Options != nil {
			jobParam["options"] = helper.ToJSONString(job.Options)
		}
		if job.TableRef != "" {
			jobParam["tableRef"] = job.TableRef
		}

		// Build SET clause dynamically
		setClause := `
			j.name = $name,
			j.nodeType = $nodeType,
			j.index = $index
		`

		if job.Options != nil {
			setClause += `, j.options = $options`
		}
		if job.TableRef != "" {
			setClause += `, j.tableRef = $tableRef`
		}

		records, err := graph.
			WithMatch("(parent)").
			WithWhere("elementId(parent) = $parentId", jobParam).
			WithMerge("(parent)-[:HAS_JOB]->(j:Job {index: $index})").
			WithSet(setClause, jobParam).
			WithParams(jobParam).
			WithReturn("elementId(j) AS jobId").
			RunWriteWithReturn()
		if err != nil {
			return fmt.Errorf("failed to merge nested job '%s': %w", job.Name, err)
		}

		if len(records) == 0 {
			return fmt.Errorf("no record returned for nested job '%s'", job.Name)
		}

		jobId := records[0].Values[0].(string)

		// Jika nested job ini punya HasJob (job -> has_job -> job -> has_job)
		if len(job.HasJob) != 0 {
			fmt.Printf("[CANVAS] : Nested job '%s' has nested jobs, processing recursively...\n", job.Name)
			if err := handleJobRelation(jobId, job); err != nil {
				return fmt.Errorf("failed to handle nested job relation for '%s': %w", job.Name, err)
			}
		}

		// Jika nested job ini punya HasColumn (job bertipe table di dalam Canvas)
		if len(job.HasColumn) != 0 {
			fmt.Printf("[CANVAS] : Nested job '%s' has columns, processing...\n", job.Name)
			for _, nestedCol := range job.HasColumn {
				nestedColParam := map[string]any{
					"jobId":    jobId,
					"name":     nestedCol.Name,
					"nodeType": nestedCol.NodeType,
					"options":  helper.ToJSONString(nestedCol.Options),
					"tableRef": nestedCol.TableRef,
					"index":    nestedCol.Index,
				}

				_, err := graph.
					WithMatch("(j:Job)").
					WithWhere("elementId(j) = $jobId", nestedColParam).
					WithMerge("(j)-[:HAS_COLUMN]->(c:Column {index: $index})").
					WithSet(`c.name = $name,
						c.nodeType = $nodeType,
						c.options = $options,
						c.tableRef = $tableRef`, nestedColParam).
					WithParams(nestedColParam).
					WithReturn("elementId(c) AS columnId").
					RunWriteWithReturn()

				if nestedCol.NodeType == "table" {
					fmt.Printf("[CANVAS] : Nested table in column: %s\n", nestedCol.Name)
					insertNestedColumn(nestedCol)
				}

				if nestedCol.NodeType == "canvas" {
					fmt.Println("[CANVAS] : Nested canvas detected, inserting job graph...")
					_, err := insertJobGraph(nestedCol)
					if err != nil {
						return fmt.Errorf("failed to create job graph for nested canvas column '%s': %w", nestedCol.Name, err)
					}
				}

				if err != nil {
					return fmt.Errorf("failed to create column for nested job '%s': %w", job.Name, err)
				}
			}
		}
	}

	return nil
}

func insertJobGraph(column dto.ColumnDto) (string, error) {
	for _, job := range column.HasJob {

		// Step 1: Insert Job node
		jobNode := dto.NodeData{
			Labels: []string{"Job"},
			Props: map[string]any{
				"name":     job.Name,
				"nodeType": job.NodeType,
				"options":  helper.ToJSONString(job.Options),
				"tableRef": job.TableRef,
				"index":    job.Index,
			},
			Relationship: "HAS_JOB",
		}

		// IMPORTANT: create job node and capture its elementId
		jobGraph := dto.GraphNode{
			ElementId: column.ElementId, // parent = column node
			Node:      jobNode,
		}

		record, err := CreateGraph(jobGraph)
		if err != nil {
			return "", fmt.Errorf("failed inserting job '%s': %w", job.Name, err)
		}

		jobElementId, _ := ExtractElementId(record.Values[0])

		// Step 2: Insert nested columns under this job
		if len(job.HasColumn) > 0 {
			for _, nestedCol := range job.HasColumn {

				colData := dto.NodeData{
					Labels: []string{"Column"},
					Props: map[string]any{
						"name":     nestedCol.Name,
						"nodeType": nestedCol.NodeType,
						"options":  helper.ToJSONString(nestedCol.Options),
						"tableRef": nestedCol.TableRef,
						"index":    nestedCol.Index,
					},
					Relationship: "HAS_COLUMN",
				}

				colGraph := dto.GraphNode{
					ElementId: jobElementId, // parent job node
					Node:      colData,
				}

				_, err := CreateGraph(colGraph)

				if nestedCol.NodeType == "canvas" {
					fmt.Println("[CANVAS] : Nested canvas detected in job, inserting job graph...")
					_, err := insertJobGraph(nestedCol)
					if err != nil {
						return "", fmt.Errorf("failed to create job graph for nested canvas column '%s': %w", nestedCol.Name, err)
					}
				}

				if nestedCol.NodeType == "table" {
					fmt.Printf("[TABLE] : Nested table in job lo ini: %s\n", nestedCol.Name)
					fmt.Printf("[TABLE] : Nested table in job lo ini: %v\n", nestedCol.HasColumn)
					insertNestedColumn(nestedCol)
				}

				if err != nil {
					return "", fmt.Errorf("failed to create/update nested column '%s': %w", nestedCol.Name, err)
				}
			}
		}

		if len(job.HasJob) > 0 {
			fmt.Println("INI ADA NESTED JOB LAGI", jobElementId)
			fmt.Println("INI ADA NESTED JOB LAGI", job.Name)
			fmt.Println("INI ADA NESTED JOB LAGI", job.HasJob)
			for _, nestedJob := range job.HasJob {
				nestedJobColumn := dto.ColumnDto{
					ElementId: jobElementId,
					HasJob:    []dto.ColumnDto{nestedJob},
				}
				_, err := insertJobGraph(nestedJobColumn)

				if job.NodeType == "table" {
					fmt.Printf("[TABLE] : Nested table in job: %s\n", nestedJob.Name)
					insertNestedColumn(nestedJob)
				}

				if err != nil {
					return "", fmt.Errorf("failed to create nested job '%s': %w", nestedJob.Name, err)
				}
			}
		}
	}

	return column.ElementId, nil
}
