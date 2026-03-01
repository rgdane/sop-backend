package sql

import (
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/pkg/gorm/builder"
	"time"

	"gorm.io/gorm"
)

type SopJobRepository interface {
	WithTx(tx *gorm.DB) SopJobRepository
	WithPreloads(preloads ...string) SopJobRepository
	WithAssociations(associations ...string) SopJobRepository
	WithReplacements(replacements map[string]any) SopJobRepository
	WithJoins(joins ...string) SopJobRepository
	WithWhere(query any, args ...any) SopJobRepository
	WithOrder(order string) SopJobRepository
	WithLimit(limit int) SopJobRepository
	WithCursor(cursor int) SopJobRepository

	InsertSopJob(data *models.SopJob) (*models.SopJob, error)
	InsertManySopJobs(data []*models.SopJob) ([]*models.SopJob, error)
	UpdateSopJob(id int64, updates map[string]any) (*models.SopJob, error)
	UpdateManySopJobs(ids []int64, updates map[string]any) error
	RemoveSopJob(id int64) error
	RemoveManySopJobs(ids []int64) error

	FindSopJob() ([]models.SopJob, error)
	FindSopJobByID(id int64) (*models.SopJob, error)
	FindSopJobByIDs(ids []int64) ([]*models.SopJob, error)
	ReorderSopJob(sopJobID int64, newIndex int, sopID int64) error
}

type sopJobRepository struct {
	db           *gorm.DB
	preloads     []string
	associations []string
	replacements map[string]interface{}
	joins        []string
	whereClauses []func(*gorm.DB) *gorm.DB
	order        string
	limit        *int
	cursor       *int
}

func NewSopJobRepository() SopJobRepository {
	return &sopJobRepository{db: config.DB}
}

func (repo *sopJobRepository) clone() *sopJobRepository {
	clone := *repo
	return &clone
}

func (repo *sopJobRepository) WithTx(tx *gorm.DB) SopJobRepository {
	clone := repo.clone()
	clone.db = tx
	return clone
}

func (repo *sopJobRepository) WithPreloads(preloads ...string) SopJobRepository {
	clone := repo.clone()
	clone.preloads = append(clone.preloads, preloads...)
	return clone
}

func (repo *sopJobRepository) WithAssociations(associations ...string) SopJobRepository {
	clone := repo.clone()
	clone.associations = append(clone.associations, associations...)
	return clone
}

func (repo *sopJobRepository) WithReplacements(replacements map[string]interface{}) SopJobRepository {
	clone := repo.clone()
	clone.replacements = replacements
	return clone
}

func (repo *sopJobRepository) WithJoins(joins ...string) SopJobRepository {
	clone := repo.clone()
	clone.joins = append(clone.joins, joins...)
	return clone
}

func (repo *sopJobRepository) WithWhere(query interface{}, args ...interface{}) SopJobRepository {
	clone := repo.clone()
	clone.whereClauses = append(clone.whereClauses, func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
	return clone
}

func (repo *sopJobRepository) WithOrder(order string) SopJobRepository {
	clone := repo.clone()
	clone.order = order
	return clone
}

func (repo *sopJobRepository) WithLimit(limit int) SopJobRepository {
	clone := repo.clone()
	clone.limit = &limit
	return clone
}

func (repo *sopJobRepository) WithCursor(cursor int) SopJobRepository {
	clone := repo.clone()
	clone.cursor = &cursor
	return clone
}

func (repo *sopJobRepository) getQueryBuilder() *builder.QueryBuilder[models.SopJob] {
	qb := builder.NewQueryBuilder[models.SopJob](repo.db).
		WithPreloads(repo.preloads...).
		WithAssociations(repo.associations...).
		WithReplacements(repo.replacements).
		WithJoins(repo.joins...).
		WithOrder(repo.order)

	for _, where := range repo.whereClauses {
		qb = qb.WithWhere(where)
	}
	if repo.limit != nil {
		qb = qb.WithLimit(*repo.limit)
	}
	if repo.cursor != nil {
		qb = qb.WithCursor(*repo.cursor)
	}
	return qb
}

func (repo *sopJobRepository) InsertSopJob(data *models.SopJob) (*models.SopJob, error) {
	if err := repo.getQueryBuilder().Create(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *sopJobRepository) InsertManySopJobs(data []*models.SopJob) ([]*models.SopJob, error) {
	if err := repo.getQueryBuilder().CreateMany(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *sopJobRepository) UpdateSopJob(id int64, updates map[string]interface{}) (*models.SopJob, error) {
	return repo.getQueryBuilder().UpdateByID(id, updates)
}

func (repo *sopJobRepository) UpdateManySopJobs(ids []int64, updates map[string]interface{}) error {
	_, err := repo.getQueryBuilder().UpdateMany(ids, updates)
	return err
}

func (repo *sopJobRepository) RemoveSopJob(id int64) error {
	return repo.getQueryBuilder().Delete(id)
}

func (repo *sopJobRepository) RemoveManySopJobs(ids []int64) error {
	var job models.SopJob
	return repo.db.Where("id IN ?", ids).Delete(&job).Error
}

func (repo *sopJobRepository) FindSopJob() ([]models.SopJob, error) {
	r := repo
	if r.order == "" {
		if casted, ok := r.WithOrder("index ASC").(*sopJobRepository); ok {
			r = casted
		}
	}
	return repo.getQueryBuilder().FindAll()
}

func (repo *sopJobRepository) FindSopJobByID(id int64) (*models.SopJob, error) {
	return repo.getQueryBuilder().FindByID(id)
}

func (repo *sopJobRepository) FindSopJobByIDs(ids []int64) ([]*models.SopJob, error) {
	if len(ids) == 0 {
		return []*models.SopJob{}, nil
	}

	return repo.getQueryBuilder().WithWhere(func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN ?", ids)
	}).FindAllPtr()
}

func (repo *sopJobRepository) ReorderSopJob(sopJobID int64, newIndex int, sopID int64) error {
	// Get current sop job
	var currentSopJob models.SopJob
	if err := repo.db.Where("id = ? AND sop_id = ?", sopJobID, sopID).First(&currentSopJob).Error; err != nil {
		return err
	}

	oldIndex := currentSopJob.Index

	// If no change needed
	if oldIndex == newIndex {
		return nil
	}

	// Validate newIndex is within valid range
	var maxIndex int64
	if err := repo.db.Model(&models.SopJob{}).Where("sop_id = ?", sopID).Count(&maxIndex).Error; err != nil {
		return err
	}

	if newIndex > int(maxIndex) {
		newIndex = int(maxIndex)
	}
	if newIndex < 1 {
		newIndex = 1
	}

	// Reorder logic
	if oldIndex < newIndex {
		// Moving down: shift items up between oldIndex+1 and newIndex
		if err := repo.db.Model(&models.SopJob{}).
			Where("sop_id = ? AND index > ? AND index <= ?", sopID, oldIndex, newIndex).
			UpdateColumn("index", gorm.Expr("index - 1")).Error; err != nil {
			return err
		}
	} else {
		// Moving up: shift items down between newIndex and oldIndex-1
		if err := repo.db.Model(&models.SopJob{}).
			Where("sop_id = ? AND index >= ? AND index < ?", sopID, newIndex, oldIndex).
			UpdateColumn("index", gorm.Expr("index + 1")).Error; err != nil {
			return err
		}
	}

	// Update the moved sop job
	if err := repo.db.Model(&models.SopJob{}).
		Where("id = ?", sopJobID).
		UpdateColumns(map[string]interface{}{
			"index":      newIndex,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return err
	}

	return nil
}
