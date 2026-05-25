package sql

import (
	"jk-api/internal/config"

	"gorm.io/gorm"
)

// SqlCountResult digunakan untuk memetakan kolom hasil raw query UNION ALL
type SqlCountResult struct {
	SqlDivision  int64 `gorm:"column:sql_divisions"`
	SqlTitle     int64 `gorm:"column:sql_titles"`
	SqlFlowchart int64 `gorm:"column:sql_flowcharts"`
	SqlSop       int64 `gorm:"column:sql_sops"`
	SqlSopJob    int64 `gorm:"column:sql_sop_jobs"`
	SqlSpk       int64 `gorm:"column:sql_spks"`
	SqlSpkJob    int64 `gorm:"column:sql_spk_jobs"`
}

type DashboardRepository interface {
	WithTx(tx *gorm.DB) DashboardRepository
	GetSqlCounts() ([]SqlCountResult, error)
}

type dashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository() DashboardRepository {
	return &dashboardRepository{db: config.DB}
}

func (repo *dashboardRepository) clone() *dashboardRepository {
	clone := *repo
	return &clone
}

// Mengikuti pola chaining WithTx milik Anda untuk transaksi aman jika dibutuhkan
func (repo *dashboardRepository) WithTx(tx *gorm.DB) DashboardRepository {
	clone := repo.clone()
	clone.db = tx
	return clone
}

func (repo *dashboardRepository) GetSqlCounts() ([]SqlCountResult, error) {
	var results []SqlCountResult

	query := `
		SELECT
			(SELECT COUNT(*) FROM divisions)  AS sql_divisions,
			(SELECT COUNT(*) FROM titles)     AS sql_titles,
			(SELECT COUNT(*) FROM flowcharts) AS sql_flowcharts,
			(SELECT COUNT(*) FROM sops)       AS sql_sops,
			(SELECT COUNT(*) FROM sop_jobs)   AS sql_sop_jobs,
			(SELECT COUNT(*) FROM spks)       AS sql_spks,
			(SELECT COUNT(*) FROM spk_jobs)   AS sql_spk_jobs
	`

	if err := repo.db.Raw(query).Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}
