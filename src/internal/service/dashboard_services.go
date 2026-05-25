package service

import (
	"jk-api/internal/config"
	"jk-api/internal/repository/graphdb"
	"jk-api/internal/repository/sql"

	"gorm.io/gorm"
)

type DashboardService interface {
	GetGraphCounts() (*graphdb.GraphCountResult, error)
	GetSqlCounts() ([]sql.SqlCountResult, error)
}

type dashboardService struct {
	repo sql.DashboardRepository
	graphRepo graphdb.DashboardRepository
	tx   *gorm.DB
}

func NewDashboardService(repo sql.DashboardRepository, graphRepo graphdb.DashboardRepository) DashboardService {
	return &dashboardService{repo: repo, graphRepo: graphRepo}
}

func (s *dashboardService) WithTx(tx *gorm.DB) DashboardService {
	return &dashboardService{
		repo: s.repo.WithTx(tx),
		tx:   tx,
	}
}

func (s *dashboardService) GetDB() *gorm.DB {
	if s.tx != nil {
		return s.tx
	}
	return config.DB
}

func (s *dashboardService) GetGraphCounts() (*graphdb.GraphCountResult, error) {
	return s.graphRepo.GetGraphCounts()
}

func (s *dashboardService) GetSqlCounts() ([]sql.SqlCountResult, error) {
	return s.repo.GetSqlCounts()
}