package graphdb

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"jk-api/api/http/controllers/v1/dto"
	"jk-api/pkg/neo4j/builder"
)

type SopJobNode struct {
	ID           int64                `json:"id"`
	Name         string               `json:"name"`
	Alias        string               `json:"alias"`
	Type         string               `json:"type"`
	Code         string               `json:"code"`
	Description  string               `json:"description"`
	TitleID      *int64               `json:"title_id"`
	SopID        int64                `json:"sop_id"`
	ReferenceID  *int64               `json:"reference_id"`
	Index        int                  `json:"index"`
	FlowchartID  *int64               `json:"flowchart_id"`
	NextIndex    *int                 `json:"next_index"`
	PrevIndex    *int                 `json:"prev_index"`
	IsPublished  *bool                 `json:"is_published"`
	IsHide       *bool                `json:"is_hide"`
	CreatedAt    string               `json:"created_at"`
	UpdatedAt    string               `json:"updated_at"`
	HasTitle     *SopJobTitleNode      `json:"has_title,omitempty"`
	HasReference *SopJobReferenceNode  `json:"has_reference,omitempty"`
}

type SopJobTitleNode struct {
	ID         int64  `json:"id"`
	Code       string `json:"code"`
	Color      string `json:"color"`
	Name       string `json:"name"`
	DivisionID int64  `json:"divisionId"`
}

type SopJobReferenceNode struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

type SopJobRepository interface {
	GetAllGraphSopJobs(filter dto.SopJobFilterDto) ([]*SopJobNode, error)
	GetGraphSopJobByID(id int64) (*SopJobNode, error)
	InsertGraphSopJob(data *SopJobNode) error
	UpdateGraphSopJob(data *SopJobNode) error
	DeleteGraphSopJob(sopJobId int64) error

	BulkInsertGraphSopJobs(data []*SopJobNode) error
	BulkUpdateGraphSopJobs(data []*SopJobNode) error
	BulkDeleteGraphSopJobs(ids []int64) error

	CountGraphSopJobs(filter dto.SopJobFilterDto) (int64, error)
}

type sopJobRepository struct{}

func NewSopJobRepository() SopJobRepository {
	return &sopJobRepository{}
}

func mapToSopJobNode(data map[string]any) *SopJobNode {
	if data == nil {
		return nil
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil
	}

	var node SopJobNode
	if err := json.Unmarshal(jsonBytes, &node); err != nil {
		return nil
	}

	return &node
}

func (r *sopJobRepository) GetAllGraphSopJobs(filter dto.SopJobFilterDto) ([]*SopJobNode, error) {
	repo := builder.NewGraphRepository()
	params := make(map[string]any)

	// Array penyimpan kondisi WHERE
	var conditions []string

	// 1. Filter Nama SOP (Menggunakan search_name dan CONTAINS murni)
	if filter.SopName != "" {
		conditions = append(conditions, "s.search_name CONTAINS $sopName")
		params["sopName"] = strings.ToLower(filter.SopName)
	}

	// 2. Filter Nama Divisi (Karena Seeder kita juga menyuntikkan d.search_name, gunakan itu)
	if len(filter.DivisionNames) > 0 {
		var lowerDivs []string
		for _, div := range filter.DivisionNames {
			lowerDivs = append(lowerDivs, strings.ToLower(div))
		}
		params["divNames"] = lowerDivs
		// Index Seek murni tanpa toLower()
		conditions = append(conditions, "d.search_name IN $divNames")
	}

	// 3. Filter Atribut Standar Lainnya
	if filter.SopID != 0 {
		conditions = append(conditions, "s.id = $sopId")
		params["sopId"] = filter.SopID
	}

	if filter.ShowDeleted {
		conditions = append(conditions, "j.deleted_at IS NOT NULL")
	} else {
		conditions = append(conditions, "j.deleted_at IS NULL")
	}

	// 4. Filter Nama Job (Menggunakan search_name dan CONTAINS murni)
	if filter.Name != "" {
		conditions = append(conditions, "j.search_name CONTAINS $jobName")
		params["jobName"] = strings.ToLower(filter.Name)
	}

	if filter.MinIndex > 0 {
		conditions = append(conditions, "j.index > $minIndex")
		params["minIndex"] = filter.MinIndex
	}

	if filter.ReferenceID != nil && *filter.ReferenceID != 0 {
		conditions = append(conditions, "j.reference_id = $referenceId")
		params["referenceId"] = *filter.ReferenceID
	}

	if filter.ReferenceType != "" {
		conditions = append(conditions, "j.type = $referenceType")
		params["referenceType"] = filter.ReferenceType
	}

	// JALUR UTAMA: Jalur traversal graf
	repo = repo.WithMatch("(d:Division)-[:HAS_SOP]->(s:SOP)-[:HAS_JOB]->(j:Job)")

	// Gabungkan semua kondisi jika ada
	if len(conditions) > 0 {
		repo = repo.WithWhere(strings.Join(conditions, " AND "), params)
	}

	// =========================================================================
	// OPTIMASI KUNCI: EARLY PAGINATION & SORTING (DILAKUKAN DI TENGAH KUERI)
	// =========================================================================
	
	// Pengurutan data (Sorting)
	orderByClause := "ORDER BY j.index ASC"
	if filter.Sort != "" && filter.Order != "" {
		orderByClause = fmt.Sprintf("ORDER BY j.%s %s", filter.Sort, strings.ToUpper(filter.Order))
	}

	// Pagination (LIMIT & SKIP) wajib untuk menahan beban data jutaan
	paginationClause := ""
	if filter.Limit > 0 {
		var offset int64 = 0
		if filter.Page > 1 {
			offset = (filter.Page - 1) * filter.Limit
		}
		// Di Neo4j, SKIP adalah pengganti OFFSET
		paginationClause = fmt.Sprintf(" SKIP %d LIMIT %d", offset, filter.Limit)
	} else {
		// Safety default limit untuk mencegah full graph scan
		paginationClause = " LIMIT 100"
	}

	// Terapkan Limit dan Sorting SEBELUM masuk ke OPTIONAL MATCH!
	// Ini akan mencegah RAM jebol karena "Cartesian Product" atau row multiplication
	withClause := fmt.Sprintf("j, s, d %s%s", orderByClause, paginationClause)
	repo = repo.WithWith(withClause)

	// =========================================================================

	// OPTIONAL MATCH ini sekarang sangat ringan karena datanya sudah di-limit misal hanya 10 atau 20 baris
	repo = repo.
		WithOptionalMatch("(j)-[:ASSIGNED_TO]->(t:Title)").
		WithWhere("(d)<-[:HAS_TITLE]-(t)", nil).
		WithWith("j, s, t").
		WithOptionalMatch("(j)-[:HAS_REFERENCE]->(ref)").
		WithWith("j, s, t, ref")

	// Proyeksi JSON
	// CATATAN: Kita tetap me-return `.name` agar UI Front-End tetap melihat huruf kapital yang cantik
	returnClause := `j {
		.id,
		.name,
		.alias,
		.type,
		.code,
		.description,
		.index,
		.is_published,
		.is_hide,
		.created_at,
		.updated_at,
		sop_id: s.id,
		title_id: t.id,
		reference_id: ref.id,
		has_title: t { .id, .code, .color, .name, divisionId: t.divisionId },
		has_reference: ref { .id, .name, .code, .description }
	} AS data`

	// Bagian akhir ini tidak butuh ORDER BY lagi, karena sudah diurutkan di atas
	repo = repo.WithReturn(returnClause).WithParams(params)

	// Eksekusi kueri ke database Neo4j
	records, err := repo.RunRead()
	if err != nil {
		return nil, fmt.Errorf("failed to get SOP Jobs with traversal: %w", err)
	}

	// Mapping hasil record Neo4j ke struct Go
	sopJobs := make([]*SopJobNode, 0, len(records))
	for _, record := range records {
		if dataVal, ok := record.Get("data"); ok {
			if props, ok := dataVal.(map[string]any); ok {
				if node := mapToSopJobNode(props); node != nil {
					sopJobs = append(sopJobs, node)
				}
			}
		}
	}

	return sopJobs, nil
}

func (r *sopJobRepository) GetGraphSopJobByID(id int64) (*SopJobNode, error) {
	repo := builder.NewGraphRepository()
	params := map[string]any{
		"id": id,
	}

	records, err := repo.
		WithMatch("(j:Job)").
		WithWhere("j.id = $id AND j.deleted_at IS NULL", params).
		WithReturn("j {.*} AS data").
		WithParams(params).
		RunRead()

	if err != nil {
		return nil, fmt.Errorf("failed to get SOP Job node by ID: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("SOP Job node with ID %d not found", id)
	}

	dataVal, ok := records[0].Get("data")
	if !ok {
		return nil, fmt.Errorf("field 'data' not found in record")
	}

	props, ok := dataVal.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("data is not a map[string]any")
	}

	return mapToSopJobNode(props), nil
}

func (r *sopJobRepository) InsertGraphSopJob(data *SopJobNode) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"id":          data.ID,
		"name":        data.Name,
		"alias":       data.Alias,
		"type":        data.Type,
		"code":        data.Code,
		"description": data.Description,
		"titleId":     data.TitleID,
		"sopId":       data.SopID,
		"referenceId": data.ReferenceID,
		"index":       data.Index,
		"flowchartId": data.FlowchartID,
		"nextIndex":   data.NextIndex,
		"prevIndex":   data.PrevIndex,
		"isPublished": data.IsPublished,
		"isHide":      data.IsHide,
		"createdAt":   data.CreatedAt,
		"updatedAt":   data.UpdatedAt,
	}

	if err := graph.
		WithMerge("(j:Job {id: $id})").
		WithSet(`j.name = $name, 
			j.alias = $alias, 
			j.type = $type, 
			j.code = $code, 
			j.description = $description,
			j.title_id = $titleId, 
			j.sop_id = $sopId, 
			j.reference_id = $referenceId, 
			j.index = $index,
			j.flowchart_id = $flowchartId,
			j.next_index = $nextIndex, 
			j.prev_index = $prevIndex,
			j.is_published = $isPublished,
			j.is_hide = $isHide,
			j.created_at = datetime($createdAt), 
			j.updated_at = datetime($updatedAt)`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to merge Job node: %w", err)
	}

	return nil
}

func (r *sopJobRepository) UpdateGraphSopJob(data *SopJobNode) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"id":          data.ID,
		"name":        data.Name,
		"alias":       data.Alias,
		"type":        data.Type,
		"code":        data.Code,
		"description": data.Description,
		"titleId":     data.TitleID,
		"sopId":       data.SopID,
		"referenceId": data.ReferenceID,
		"index":       data.Index,
		"flowchartId": data.FlowchartID,
		"nextIndex":   data.NextIndex,
		"prevIndex":   data.PrevIndex,
		"isPublished": data.IsPublished,
		"isHide":      data.IsHide,
	}

	if err := graph.
		WithMatch("(j:Job {id: $id})").
		WithSet(`j.name = $name, 
			j.alias = $alias, 
			j.type = $type, 
			j.code = $code, 
			j.description = $description,
			j.title_id = $titleId, 
			j.sop_id = $sopId, 
			j.reference_id = $referenceId, 
			j.index = $index,
			j.flowchart_id = $flowchartId,
			j.next_index = $nextIndex, 
			j.prev_index = $prevIndex,
			j.is_published = $isPublished,
			j.is_hide = $isHide`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update Job graph with id %d: %w", data.ID, err)
	}

	return nil
}

func (r *sopJobRepository) DeleteGraphSopJob(sopJobId int64) error {
	graph := builder.NewGraphRepository()

	params := map[string]interface{}{
		"docId":     sopJobId,
		"deletedAt": time.Now().Format(time.RFC3339Nano),
	}

	if err := graph.
		WithMatch("(j:Job {id: $docId})").
		WithSet("j.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to soft delete Job graph with id %d: %w", sopJobId, err)
	}

	return nil
}

func (r *sopJobRepository) BulkInsertGraphSopJobs(data []*SopJobNode) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	sopJobNodes := make([]map[string]any, 0, len(data))
	for _, sopJob := range data {
		sopJobNodes = append(sopJobNodes, map[string]any{
			"id":          sopJob.ID,
			"name":        sopJob.Name,
			"alias":       sopJob.Alias,
			"type":        sopJob.Type,
			"code":        sopJob.Code,
			"description": sopJob.Description,
			"titleId":     sopJob.TitleID,
			"sopId":       sopJob.SopID,
			"referenceId": sopJob.ReferenceID,
			"index":       sopJob.Index,
			"flowchartId": sopJob.FlowchartID,
			"nextIndex":   sopJob.NextIndex,
			"prevIndex":   sopJob.PrevIndex,
			"isPublished": sopJob.IsPublished,
			"isHide":      sopJob.IsHide,
		})
	}

	params := map[string]any{"jobs": sopJobNodes}

	if err := graph.
		WithUnwind("$jobs", "j").
		WithMerge("(job:Job {id: j.id})").
		WithSet(`job.name = j.name, 
			job.alias = j.alias, 
			job.type = j.type, 
			job.code = j.code, 
			job.description = j.description,
			job.title_id = j.titleId, 
			job.sop_id = j.sopId, 
			job.reference_id = j.referenceId, 
			job.index = j.index,
			job.flowchart_id = j.flowchartId,
			job.next_index = j.nextIndex, 
			job.prev_index = j.prevIndex,
			job.is_published = j.isPublished,
			job.is_hide = j.isHide`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk insert Job nodes: %w", err)
	}

	return nil
}

func (r *sopJobRepository) BulkUpdateGraphSopJobs(data []*SopJobNode) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	sopJobNodes := make([]map[string]any, 0, len(data))
	for _, sopJob := range data {
		sopJobNodes = append(sopJobNodes, map[string]any{
			"id":          sopJob.ID,
			"name":        sopJob.Name,
			"alias":       sopJob.Alias,
			"type":        sopJob.Type,
			"code":        sopJob.Code,
			"description": sopJob.Description,
			"titleId":     sopJob.TitleID,
			"sopId":       sopJob.SopID,
			"referenceId": sopJob.ReferenceID,
			"index":       sopJob.Index,
			"flowchartId": sopJob.FlowchartID,
			"nextIndex":   sopJob.NextIndex,
			"prevIndex":   sopJob.PrevIndex,
			"isPublished": sopJob.IsPublished,
			"isHide":      sopJob.IsHide,
		})
	}

	params := map[string]any{"jobs": sopJobNodes}

	if err := graph.
		WithUnwind("$jobs", "j").
		WithMatch("(job:Job {id: j.id})").
		WithSet(`job.name = j.name, 
			job.alias = j.alias, 
			job.type = j.type, 
			job.code = j.code, 
			job.description = j.description,
			job.title_id = j.titleId, 
			job.sop_id = j.sopId, 
			job.reference_id = j.referenceId, 
			job.index = j.index,
			job.flowchart_id = j.flowchartId,
			job.next_index = j.nextIndex, 
			job.prev_index = j.prevIndex,
			job.is_published = j.isPublished,
			job.is_hide = j.isHide`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk update Job nodes: %w", err)
	}

	return nil
}

func (r *sopJobRepository) BulkDeleteGraphSopJobs(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	params := map[string]any{
		"jobIds":    ids,
		"deletedAt": time.Now().Format(time.RFC3339Nano),
	}

	if err := graph.
		WithMatch("(j:Job)").
		WithWhere("j.id IN $jobIds", nil).
		WithSet("j.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk soft delete Job nodes: %w", err)
	}

	return nil
}

func (r *sopJobRepository) CountGraphSopJobs(filter dto.SopJobFilterDto) (int64, error) {
	repo := builder.NewGraphRepository()
	params := make(map[string]any)

	if filter.SopID != 0 {
		repo = repo.WithMatch("(s:SOP {id: $sopId})-[:HAS_JOB]->(j:Job)")
		params["sopId"] = filter.SopID
	} else {
		repo = repo.WithMatch("(s:SOP)-[:HAS_JOB]->(j:Job)")
	}

	var mainConditions []string
	if filter.SopName != "" {
		mainConditions = append(mainConditions, "toLower(s.name) CONTAINS toLower($sopName)")
		params["sopName"] = filter.SopName
	}
	if filter.ShowDeleted {
		mainConditions = append(mainConditions, "j.deleted_at IS NOT NULL")
	} else {
		mainConditions = append(mainConditions, "j.deleted_at IS NULL")
	}
	if filter.Name != "" {
		mainConditions = append(mainConditions, "toLower(j.name) CONTAINS toLower($name)")
		params["name"] = filter.Name
	}
	if filter.MinIndex > 0 {
		mainConditions = append(mainConditions, "j.index > $minIndex")
		params["minIndex"] = filter.MinIndex
	}
	if filter.ReferenceID != nil && *filter.ReferenceID != 0 {
		mainConditions = append(mainConditions, "j.reference_id = $referenceId")
		params["referenceId"] = *filter.ReferenceID
	}
	if filter.ReferenceType != "" {
		mainConditions = append(mainConditions, "j.type = $referenceType")
		params["referenceType"] = filter.ReferenceType
	}

	if len(filter.DivisionNames) > 0 {
		repo = repo.WithMatch("(s:SOP)<-[:HAS_SOP]-(d:Division)")
		var divConditions []string
		for i, divName := range filter.DivisionNames {
			divConditions = append(divConditions, fmt.Sprintf("toLower(d.name) = toLower($divName%d)", i))
			params[fmt.Sprintf("divName%d", i)] = divName
		}
		mainConditions = append(mainConditions, "("+strings.Join(divConditions, " OR ")+")")
	}

	if len(mainConditions) > 0 {
		repo = repo.WithWhere(strings.Join(mainConditions, " AND "), params)
	}

	repo = repo.
		WithReturn("count(j) AS total").
		WithParams(params)

	records, err := repo.RunRead()
	if err != nil {
		return 0, fmt.Errorf("failed to count Job nodes: %w", err)
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