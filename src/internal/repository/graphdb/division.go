package graphdb

import (
	"fmt"
	"strings"
	"time"

	"jk-api/api/http/controllers/v1/dto"
	"jk-api/pkg/neo4j/builder"
)

// 🔹 1. Ini adalah Struct Khusus Graph (Tidak ada hubungannya dengan GORM/SQL)
type DivisionNode struct {
	ID         int64          `json:"id"`
	Name       string         `json:"name"`
	Code       string         `json:"code"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`
}

// 🔹 2. Interface sekarang menggunakan DivisionNode
type DivisionRepository interface {
	GetAllGraphDivisions(filter dto.DivisionFilterDto) ([]*DivisionNode, error)
	GetGraphDivisionByID(id int64) (*DivisionNode, error)
	InsertGraphDivision(data *DivisionNode) error
	UpdateGraphDivision(data *DivisionNode) error
	DeleteGraphDivision(divisionId int64) error

	BulkInsertGraphDivisions(data []*DivisionNode) error
	BulkUpdateGraphDivisions(data []*DivisionNode) error
	BulkDeleteGraphDivisions(ids []int64) error
	
	CountGraphDivisions(filter dto.DivisionFilterDto) (int64, error)
}

type divisionRepository struct{}

func NewDivisionRepository() DivisionRepository {
	return &divisionRepository{}
}

func (r *divisionRepository) GetAllGraphDivisions(filter dto.DivisionFilterDto) ([]*DivisionNode, error) {
	repo := builder.NewGraphRepository()
	params := make(map[string]any)

	// 1. MATCH awal
	repo = repo.WithMatch("(d:Division)")

	// 2. Dinamis WHERE Conditions
	var conditions []string

	if filter.ShowDeleted {
		conditions = append(conditions, "d.deleted_at IS NOT NULL")
	} else {
		conditions = append(conditions, "d.deleted_at IS NULL")
	}

	if filter.Name != "" {
		// Menggunakan regex/contains untuk pencarian mirip ILIKE
		conditions = append(conditions, "toLower(d.name) CONTAINS toLower($name)")
		params["name"] = filter.Name
	}

	repo = repo.WithWhere(strings.Join(conditions, " AND "), params)

	// 3. Return Clause (Termasuk eksekusi PRELOAD menggunakan Pattern Comprehension)
	// Default return tanpa preload
	returnClause := "d {.*} AS data"

	// 4. Tambahkan ORDER BY ke dalam string Return jika ada Sort
	if filter.Sort != "" && filter.Order != "" {
		// Validasi dasar agar terhindar dari error syntax
		orderDir := strings.ToUpper(filter.Order)
		if orderDir != "ASC" && orderDir != "DESC" {
			orderDir = "ASC"
		}
		
		// Sisipkan ORDER BY di akhir clause RETURN
		returnClause += fmt.Sprintf(" ORDER BY d.%s %s", filter.Sort, orderDir)
	}

	// Terapkan Return dan Params ke Builder
	repo = repo.
		WithReturn(returnClause).
		WithParams(params)

	// 5. Eksekusi Query
	records, err := repo.RunRead()
	if err != nil {
		return nil, fmt.Errorf("failed to get divisions with filter: %w", err)
	}

	// 6. Mapping hasil ke Struct
	var divisions []*DivisionNode
	for _, record := range records {
		dataVal, ok := record.Get("data")
		if !ok {
			continue
		}

		props, ok := dataVal.(map[string]any)
		if !ok {
			continue
		}

		div := mapToDivisionNode(props)
		divisions = append(divisions, div)
	}

	return divisions, nil
}

func (r *divisionRepository) GetGraphDivisionByID(id int64) (*DivisionNode, error) {
	repo := builder.NewGraphRepository()
	params := map[string]any{
		"id": id,
	}

	records, err := repo.
		WithMatch("(d:Division)").
		WithWhere("d.id = $id AND d.deleted_at IS NULL", params).
		WithReturn("d {.*} AS data").
		WithParams(params).
		RunRead()

	if err != nil {
		return nil, fmt.Errorf("failed to get division node by ID: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("division node with ID %d not found", id)
	}

	dataVal, ok := records[0].Get("data")
	if !ok {
		return nil, fmt.Errorf("field 'data' not found in record")
	}

	props, ok := dataVal.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("data is not a map[string]any")
	}

	return mapToDivisionNode(props), nil
}

// --- HELPER FUNCTION ---
func mapToDivisionNode(props map[string]any) *DivisionNode {
	div := &DivisionNode{}

	if idVal, ok := props["id"].(int64); ok {
		div.ID = idVal
	}
	
	if nameVal, ok := props["name"].(string); ok {
		div.Name = nameVal
	}
	
	if codeVal, ok := props["code"].(string); ok {
		div.Code = codeVal
	}
	
	if createdVal, ok := props["created_at"].(string); ok {
		div.CreatedAt = createdVal
	}
	
	if updatedVal, ok := props["updated_at"].(string); ok {
		div.UpdatedAt = updatedVal
	}

	return div
}

func (r *divisionRepository) InsertGraphDivision(data *DivisionNode) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"id":        data.ID,
		"name":      data.Name,
		"code":      data.Code,
		"createdAt": data.CreatedAt,
		"updatedAt": data.UpdatedAt,
	}

	if err := graph.
		WithMerge("(s:Division {id: $id})").
		WithSet(`s.name = $name, 
			s.code = $code, 
			s.created_at = datetime($createdAt), 
			s.updated_at = datetime($updatedAt)`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to merge Division node: %w", err)
	}

	return nil
}

func (r *divisionRepository) UpdateGraphDivision(data *DivisionNode) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"id":   data.ID,
		"name": data.Name,
		"code": data.Code,
	}

	if err := graph.
		WithMatch("(s:Division {id: $id})").
		WithSet("s.name = $name, s.code = $code", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update Division graph with id %d: %w", data.ID, err)
	}

	return nil
}

func (r *divisionRepository) DeleteGraphDivision(divisionId int64) error {
	graph := builder.NewGraphRepository()

	params := map[string]interface{}{
		"docId":     divisionId,
		"deletedAt": time.Now().Format(time.RFC3339Nano),
	}

	if err := graph.
		WithMatch("(d:Division {id: $docId})").
		WithSet("d.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to soft delete Division graph with id %d: %w", divisionId, err)
	}

	return nil
}

func (r *divisionRepository) BulkInsertGraphDivisions(data []*DivisionNode) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	divisionNodes := make([]map[string]any, 0, len(data))
	for _, div := range data {
		divisionNodes = append(divisionNodes, map[string]any{
			"id":   div.ID,
			"code": div.Code,
			"name": div.Name,
		})
	}

	params := map[string]any{"divisions": divisionNodes}

	if err := graph.
		WithUnwind("$divisions", "div").
		WithMerge("(d:Division {id: div.id})").
		WithSet(`d.code = div.code, d.name = div.name`, nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk insert Division nodes: %w", err)
	}

	return nil
}

func (r *divisionRepository) BulkUpdateGraphDivisions(data []*DivisionNode) error {
	if len(data) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	divisionNodes := make([]map[string]any, 0, len(data))
	for _, div := range data {
		divisionNodes = append(divisionNodes, map[string]any{
			"id":   div.ID,
			"code": div.Code,
			"name": div.Name,
		})
	}

	params := map[string]any{"divisions": divisionNodes}

	if err := graph.
		WithUnwind("$divisions", "div").
		WithMatch("(d:Division {id: div.id})").
		WithSet("d.code = div.code, d.name = div.name", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk update Division nodes: %w", err)
	}

	return nil
}

func (r *divisionRepository) BulkDeleteGraphDivisions(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	graph := builder.NewGraphRepository()

	params := map[string]any{
		"divisionIds": ids,
		"deletedAt":   time.Now().Format(time.RFC3339Nano),
	}

	if err := graph.
		WithMatch("(d:Division)").
		WithWhere("d.id IN $divisionIds", nil).
		WithSet("d.deleted_at = $deletedAt", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to bulk soft delete Division nodes: %w", err)
	}

	return nil
}

func (r *divisionRepository) CountGraphDivisions(filter dto.DivisionFilterDto) (int64, error) {
	repo := builder.NewGraphRepository()
	params := make(map[string]any)

	// 1. Base Match
	repo = repo.WithMatch("(d:Division)")

	// 2. Dynamic WHERE Conditions (Sama persis seperti GetAll)
	var conditions []string

	if filter.ShowDeleted {
		conditions = append(conditions, "d.deleted_at IS NOT NULL")
	} else {
		conditions = append(conditions, "d.deleted_at IS NULL")
	}

	if filter.Name != "" {
		conditions = append(conditions, "toLower(d.name) CONTAINS toLower($name)")
		params["name"] = filter.Name
	}

	// 3. Gabungkan filter dan return Count
	repo = repo.
		WithWhere(strings.Join(conditions, " AND "), params).
		WithReturn("count(d) AS total"). // 🔹 Di sini kunci perbedaannya
		WithParams(params)

	// 4. Eksekusi Query
	records, err := repo.RunRead()
	if err != nil {
		return 0, fmt.Errorf("failed to count division nodes: %w", err)
	}

	// 5. Ambil hasil perhitungan
	if len(records) > 0 {
		if totalVal, ok := records[0].Get("total"); ok {
			// Driver Neo4j mengembalikan fungsi agregat count() sebagai int64
			if total, isInt := totalVal.(int64); isInt {
				return total, nil
			}
		}
	}

	return 0, nil
}