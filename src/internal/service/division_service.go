package service

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/internal/repository/sql"
	"jk-api/internal/repository/graphdb"
	"jk-api/pkg/errors/gorm_err"
	"time"

	"gorm.io/gorm"
)

type DivisionService interface {
	WithTx(tx *gorm.DB) DivisionService

	CreateDivision(input *models.Division) (*models.Division, error)
	UpdateDivision(id int64, updates map[string]interface{}) (*models.Division, error)
	DeleteDivision(id int64) error
	GetAllDivisions(filter dto.DivisionFilterDto) ([]models.Division, error)
	GetDivisionByID(id int64, filter dto.DivisionFilterDto) (*models.Division, error)
	GetDivisionsByIDs(ids []int64) ([]*models.Division, error)
	GetDB() *gorm.DB
	BulkCreateDivisions(data []*models.Division) ([]*models.Division, error)
	BulkUpdateDivisions(ids []int64, updates map[string]interface{}) error
	BulkDeleteDivisions(ids []int64) error

	GetAllGraphDivisions(filter dto.DivisionFilterDto) ([]*graphdb.DivisionNode, error)
	GetGraphDivisionByID(id int64) (*graphdb.DivisionNode, error)
	InsertGraphDivision(data *graphdb.DivisionNode) error
	UpdateGraphDivision(data *graphdb.DivisionNode) error
	DeleteGraphDivision(divisionId int64) error
	BulkInsertGraphDivisions(data []*graphdb.DivisionNode) error
	BulkUpdateGraphDivisions(data []*graphdb.DivisionNode) error
	BulkDeleteGraphDivisions(ids []int64) error

	CountGraphDivisions(filter dto.DivisionFilterDto) (int64, error)
}

type divisionService struct {
	repo sql.DivisionRepository
	graphRepo graphdb.DivisionRepository
	tx   *gorm.DB
}

func NewDivisionService(repo sql.DivisionRepository, graphRepo graphdb.DivisionRepository) DivisionService {
	return &divisionService{repo: repo, graphRepo: graphRepo}
}

func (s *divisionService) WithTx(tx *gorm.DB) DivisionService {
	return &divisionService{
		repo: s.repo.WithTx(tx),
		graphRepo: s.graphRepo,
		tx:   tx,
	}
}

func (s *divisionService) GetDB() *gorm.DB {
	if s.tx != nil {
		return s.tx
	}
	return config.DB
}

func (s *divisionService) CreateDivision(input *models.Division) (*models.Division, error) {
	input.CreatedAt = time.Now()
	input.UpdatedAt = time.Now()

	data, err := s.repo.InsertDivision(input)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	return data, nil
}

func (s *divisionService) UpdateDivision(id int64, updates map[string]interface{}) (*models.Division, error) {
	repo := s.repo

	if _, ok := updates["deleted_at"]; ok {
		repo = repo.WithUnscoped()
	}

	if _, err := repo.FindDivisionByID(id); err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	updates["updated_at"] = time.Now()
	fmt.Println(updates)

	data, err := s.repo.UpdateDivision(id, updates)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	return data, nil
}

func (s *divisionService) DeleteDivision(id int64) error {
	err := s.repo.RemoveDivision(id)

	return gorm_err.TranslateGormError(err)
}

func (s *divisionService) GetAllDivisions(filter dto.DivisionFilterDto) ([]models.Division, error) {
	repo := s.repo

	if filter.Preload {
		repo = repo.WithPreloads("")
	}

	if filter.Sort != "" && filter.Order != "" {
		orderClause := filter.Sort + " " + filter.Order
		repo = repo.WithOrder(orderClause)
	}

	if filter.ShowDeleted {
		repo = repo.WithUnscoped().WithWhere("divisions.deleted_at IS NOT NULL")
	}
	fmt.Println("filter", filter)

	data, err := repo.FindDivision()
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *divisionService) GetDivisionByID(id int64, filter dto.DivisionFilterDto) (*models.Division, error) {
	repo := s.repo
	if filter.Preload {
		repo = repo.WithPreloads("")
	}
	data, err := repo.FindDivisionByID(id)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *divisionService) BulkCreateDivisions(data []*models.Division) ([]*models.Division, error) {
	datas, err := s.repo.InsertManyDivisions(data)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}

	return datas, nil
}

func (s *divisionService) BulkUpdateDivisions(ids []int64, updates map[string]interface{}) error {
	repo := s.repo

	if _, ok := updates["deleted_at"]; ok {
		repo = s.repo.WithUnscoped()
	}

	err := repo.UpdateManyDivisions(ids, updates)

	return gorm_err.TranslateGormError(err)
}

func (s *divisionService) GetDivisionsByIDs(ids []int64) ([]*models.Division, error) {
	data, err := s.repo.FindDivisionsByIDs(ids)
	if err != nil {
		return nil, gorm_err.TranslateGormError(err)
	}
	return data, nil
}

func (s *divisionService) BulkDeleteDivisions(ids []int64) error {
	err := s.repo.RemoveManyDivisions(ids)

	return gorm_err.TranslateGormError(err)
}

func (s *divisionService) GetAllGraphDivisions(filter dto.DivisionFilterDto) ([]*graphdb.DivisionNode, error) {
		return s.graphRepo.GetAllGraphDivisions(filter)
}

func (s *divisionService) GetGraphDivisionByID(id int64) (*graphdb.DivisionNode, error) {
	return s.graphRepo.GetGraphDivisionByID(id)
}

func (s *divisionService) InsertGraphDivision(data *graphdb.DivisionNode) error {
	return s.graphRepo.InsertGraphDivision(data)
}

func (s *divisionService) UpdateGraphDivision(data *graphdb.DivisionNode) error {
	return s.graphRepo.UpdateGraphDivision(data)
}

func (s *divisionService) DeleteGraphDivision(divisionId int64) error {
	return s.graphRepo.DeleteGraphDivision(divisionId)
}

func (s *divisionService) BulkInsertGraphDivisions(data []*graphdb.DivisionNode) error {
	return s.graphRepo.BulkInsertGraphDivisions(data)
}

func (s *divisionService) BulkUpdateGraphDivisions(data []*graphdb.DivisionNode) error {
	return s.graphRepo.BulkUpdateGraphDivisions(data)
}

func (s *divisionService) BulkDeleteGraphDivisions(ids []int64) error {
	return s.graphRepo.BulkDeleteGraphDivisions(ids)
}

// Di dalam struct implementasi Service:
func (s *divisionService) CountGraphDivisions(filter dto.DivisionFilterDto) (int64, error) {
	return s.graphRepo.CountGraphDivisions(filter)
}