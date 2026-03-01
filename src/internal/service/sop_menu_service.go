package service

import (
	"fmt"
	"time"

	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/internal/repository/sql"
	"jk-api/pkg/errors/gorm_err"
	"jk-api/pkg/neo4j/builder"

	"gorm.io/gorm"
)

type SopMenuService interface {
	WithTx(tx *gorm.DB) SopMenuService

	CreateSopMenu(input *models.SopMenu, filter dto.SopMenuFilterDto) (*models.SopMenu, error)
	UpdateSopMenu(id int64, updates map[string]interface{}) (*models.SopMenu, error)
	DeleteSopMenu(id int64) error
	GetAllSopMenus(filter dto.SopMenuFilterDto) ([]models.SopMenu, error)
	GetSopMenuByID(id int64, filter dto.SopMenuFilterDto) (*models.SopMenu, error)
	InsertGraphSopMenu(data *models.SopMenu, projectId int64) error

	GetDB() *gorm.DB
}

type sopMenuService struct {
	repo sql.SopMenuRepository
	tx   *gorm.DB
}

func NewSopMenuService(repo sql.SopMenuRepository) SopMenuService {
	return &sopMenuService{repo: repo}
}

func (s *sopMenuService) WithTx(tx *gorm.DB) SopMenuService {
	return &sopMenuService{
		repo: s.repo.WithTx(tx),
		tx:   tx,
	}
}

func (s *sopMenuService) GetDB() *gorm.DB {
	if s.tx != nil {
		return s.tx
	}
	return config.DB
}

func (s *sopMenuService) CreateSopMenu(input *models.SopMenu, filter dto.SopMenuFilterDto) (*models.SopMenu, error) {
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	data, err := s.repo.InsertSopMenu(input)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	// if filter.IsCreateGraph {
	if err := s.InsertGraphSopMenu(data, filter.ProjectID); err != nil {
		return nil, fmt.Errorf("neo4j sync failed: %w", err)
	}
	// }

	return data, nil
}

func (s *sopMenuService) UpdateSopMenu(id int64, updates map[string]interface{}) (*models.SopMenu, error) {
	if _, exists := updates["parent_id"]; !exists {
		updates["parent_id"] = nil
	} else if parentID := updates["parent_id"]; parentID == nil || parentID == 0 || parentID == "0" {
		updates["parent_id"] = nil
	}

	if _, exists := updates["master_id"]; !exists {
		updates["master_id"] = nil
	} else if masterID := updates["master_id"]; masterID == nil || masterID == 0 || masterID == "0" {
		updates["master_id"] = nil
	}

	updates["updated_at"] = time.Now()

	data, err := s.repo.UpdateSopMenu(id, updates)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	// Update graph (termasuk update relasi HAS_SOP jika sop_id berubah)
	if err := s.updateGraphSopMenu(id, data); err != nil {
		return nil, fmt.Errorf("failed to update graph: %w", err)
	}

	return data, nil
}

func (s *sopMenuService) DeleteSopMenu(id int64) error {
	sopMenu, err := s.repo.FindSopMenuByID(id)
	if err != nil {
		return gorm_err.TranslateGormError(err)
	}

	if err := s.deleteGraphSopMenu(sopMenu); err != nil {
		return fmt.Errorf("neo4j sync failed: %w", err)
	}

	// Use transaction to ensure all operations succeed or fail together
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.SopMenu{}).
			Where("parent_id = ?", id).
			Update("parent_id", nil).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.SopMenu{}).
			Where("master_id = ?", id).
			Update("master_id", nil).Error; err != nil {
			return err
		}

		// Now delete the menu
		if err := tx.Delete(&models.SopMenu{}, id).Error; err != nil {
			return err
		}

		return nil
	})

	return gorm_err.TranslateGormError(err)
}

func (s *sopMenuService) GetAllSopMenus(filter dto.SopMenuFilterDto) ([]models.SopMenu, error) {
	repo := s.repo
	if filter.Preload {
		repo = repo.WithPreloads("HasSop.HasJobs", "HasSop.HasDivisions", "HasDivision", "HasDocuments")
	}
	if filter.Type != "" {
		repo = repo.WithWhere("type = ?", filter.Type)
	}
	if filter.DivisionID != 0 {
		repo = repo.WithWhere("division_id = ?", filter.DivisionID)
	}
	if filter.MasterID != 0 {
		repo = repo.WithWhere("master_id = ?", filter.MasterID)
	}
	if filter.Parent {
		repo = repo.WithWhere("parent_id IS NULL")
	}
	if filter.IsMaster {
		repo = repo.WithWhere("is_master = ?", true)
	}
	if filter.Name != "" {
		repo = repo.WithWhere("name ILIKE ?", "%"+filter.Name+"%")
	}

	data, err := repo.FindSopMenus()
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *sopMenuService) GetSopMenuByID(id int64, filter dto.SopMenuFilterDto) (*models.SopMenu, error) {
	repo := s.repo
	if filter.Preload {
		repo = repo.WithPreloads("HasSop.HasJobs")
	}
	data, err := repo.FindSopMenuByID(id)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *sopMenuService) InsertGraphSopMenu(data *models.SopMenu, projectId int64) error {
	graph := builder.NewGraphRepository()

	sopMenu, err := s.GetSopMenuByID(data.ID, dto.SopMenuFilterDto{Preload: true})
	if err != nil {
		return fmt.Errorf("failed to load sop menu with relations: %w", err)
	}

	// Create/merge Document
	params := map[string]interface{}{
		"docName":  data.Name,
		"docId":    data.ID,
		"multiple": data.Multiple,
		"type":     data.Type,
		"parentId": data.ParentID,
		"masterId": data.MasterID,
		"isMaster": data.IsMaster,
	}
	if err := graph.
		WithMerge("(m:Document {id: $docId})").
		WithSet("m.name = $docName, m.multiple = $multiple, m.type = $type, m.parent_id = $parentId, m.master_id = $masterId, m.is_master = $isMaster, m.updated_at = datetime()", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to create Document: %w", err)
	}

	// Jika sop_id tidak kosong, buat relasi HAS_SOP ke SOP terkait
	if data.SopID != nil && *data.SopID != 0 {
		// Cek apakah HasSop ada
		if sopMenu.HasSop == nil {
			return fmt.Errorf("sop relation not found")
		}

		sopParams := map[string]interface{}{
			"sopId": *data.SopID,
		}
		if err := graph.
			WithMatch("(s:SOP)").
			WithWhere("s.id = toInteger($sopId)", nil).
			WithMerge("(m:Document {id: $docId})").
			WithRelate("m", "HAS_SOP", "s", nil).
			WithParams(sopParams).
			WithParams(params).
			RunWrite(); err != nil {
			return fmt.Errorf("failed to relate Document -> SOP: %w", err)
		}
	}

	// Jika master_id tidak kosong, buat relasi HAS_MASTER ke master SopMenu
	if data.MasterID != nil && *data.MasterID != 0 {
		masterParams := map[string]interface{}{
			"masterId": *data.MasterID,
		}
		if err := graph.
			WithMerge("(master:Document {id: $masterId})").
			WithMerge("(child:Document {id: $docId})").
			WithRelate("child", "HAS_MASTER", "master", nil).
			WithParams(masterParams).
			WithParams(params).
			RunWrite(); err != nil {
			return fmt.Errorf("failed to relate Document -> Master Document: %w", err)
		}
	}

	return nil
}

func (s *sopMenuService) updateGraphSopMenu(id int64, data *models.SopMenu) error {
	graph := builder.NewGraphRepository()
	params := map[string]interface{}{
		"docId":    data.ID,
		"name":     data.Name,
		"multiple": data.Multiple,
		"type":     data.Type,
		"parentId": data.ParentID,
		"masterId": data.MasterID,
		"isMaster": data.IsMaster,
	}

	// Update properties Document node
	if err := graph.
		WithMerge("(m:Document {id: $docId})").
		WithSet("m.name = $name", nil).
		WithSet("m.multiple = $multiple", nil).
		WithSet("m.type = $type", nil).
		WithSet("m.parent_id = $parentId", nil).
		WithSet("m.master_id = $masterId", nil).
		WithSet("m.is_master = $isMaster", nil).
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to update graph Document: %w", err)
	}

	// Selalu hapus relasi HAS_SOP lama jika ada (terlepas dari nilai baru)
	if err := graph.
		WithMatch("(child:Document {id: $docId})").
		WithOptionalMatch("(child)-[r:HAS_SOP]->()").
		WithDelete("r").
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to delete old HAS_SOP relationship: %w", err)
	}

	// Jika sop_id tidak kosong, buat relasi HAS_SOP ke SOP terkait
	if data.SopID != nil && *data.SopID != 0 {
		sopParams := map[string]interface{}{
			"sopId": *data.SopID,
			"docId": data.ID,
		}
		if err := graph.
			WithMatch("(sop:SOP {id: $sopId})").
			WithMatch("(child:Document {id: $docId})").
			WithMerge("(child)-[:HAS_SOP]->(sop)").
			WithParams(sopParams).
			RunWrite(); err != nil {
			return fmt.Errorf("failed to relate Document -> SOP: %w", err)
		}
	}

	// Selalu hapus relasi HAS_MASTER lama jika ada (terlepas dari nilai baru)
	if err := graph.
		WithMatch("(child:Document {id: $docId})").
		WithOptionalMatch("(child)-[r:HAS_MASTER]->()").
		WithDelete("r").
		WithParams(params).
		RunWrite(); err != nil {
		return fmt.Errorf("failed to delete old HAS_MASTER relationship: %w", err)
	}

	// Jika master_id tidak kosong, buat relasi HAS_MASTER ke master SopMenu
	if data.MasterID != nil && *data.MasterID != 0 {
		masterParams := map[string]interface{}{
			"masterId": *data.MasterID,
			"docId":    data.ID,
		}
		if err := graph.
			WithMatch("(master:Document {id: $masterId})").
			WithMatch("(child:Document {id: $docId})").
			WithMerge("(child)-[:HAS_MASTER]->(master)").
			WithParams(masterParams).
			RunWrite(); err != nil {
			return fmt.Errorf("failed to relate Document -> Master Document: %w", err)
		}
	}

	return nil
}

func (s *sopMenuService) deleteGraphSopMenu(data *models.SopMenu) error {
	graph := builder.NewGraphRepository()

	params := map[string]any{
		"docId": data.ID,
	}

	records, err := graph.
		WithMatch("(d:Document {id: $docId})").
		WithOptionalMatch("(d)-[:HAS_SOP]->(s)").
		WithOptionalMatch("(s)-[:HAS_JOB]->(j)").
		WithOptionalMatch("(j)-[:HAS_ROW]->(r)").
		WithOptionalMatch("(d)-[hm:HAS_MASTER]->(master)").
		WithDetachDelete("d,r").
		WithReturn("d,r"). // Move RETURN after DELETE
		WithParams(params).
		RunWriteWithReturn()

	if err != nil {
		return fmt.Errorf("failed to delete ROW nodes: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("document with id %v not found", data.ID)
	}

	docElementId, _ := ExtractElementId(records[0].Values[0])

	fmt.Println("records: ", docElementId)
	deleteParams := map[string]any{
		"deleteDocId": docElementId,
	}

	err = graph.
		WithMatch("(n:Row {documentId: $deleteDocId})").
		WithDetachDelete("n").
		WithParams(deleteParams).
		RunWrite()

	return nil
}
