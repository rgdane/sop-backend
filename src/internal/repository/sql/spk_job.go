package sql

import (
	"fmt"
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/pkg/gorm/builder"
	"time"

	"gorm.io/gorm"
)

type SpkJobRepository interface {
	WithTx(tx *gorm.DB) SpkJobRepository
	WithPreloads(preloads ...string) SpkJobRepository
	WithAssociations(associations ...string) SpkJobRepository
	WithReplacements(replacements map[string]interface{}) SpkJobRepository
	WithJoins(joins ...string) SpkJobRepository
	WithWhere(query interface{}, args ...interface{}) SpkJobRepository
	WithOrder(order string) SpkJobRepository
	WithLimit(limit int) SpkJobRepository
	WithCursor(cursor int) SpkJobRepository

	InsertSpkJob(data *models.SpkJob) (*models.SpkJob, error)
	InsertManySpkJobs(data []*models.SpkJob) ([]*models.SpkJob, error)
	UpdateSpkJob(id int64, updates map[string]interface{}) (*models.SpkJob, error)
	UpdateManySpkJobs(ids []int64, updates map[string]interface{}) error
	RemoveSpkJob(id int64) error
	RemoveManySpkJobs(ids []int64) error

	FindSpkJobs() ([]models.SpkJob, error)
	FindSpkJobByID(id int64) (*models.SpkJob, error)
	FindSpkJobsByIDs(ids []int64) ([]*models.SpkJob, error)

	ReorderSpkJob(spkJobID int64, newIndex int, spkID int64) error
}

type spkJobRepository struct {
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

func NewSpkJobRepository() SpkJobRepository {
	return &spkJobRepository{db: config.DB}
}

func (repo *spkJobRepository) clone() *spkJobRepository {
	clone := *repo
	return &clone
}

func (repo *spkJobRepository) WithTx(tx *gorm.DB) SpkJobRepository {
	clone := repo.clone()
	clone.db = tx
	return clone
}

func (repo *spkJobRepository) WithPreloads(preloads ...string) SpkJobRepository {
	clone := repo.clone()
	clone.preloads = append(clone.preloads, preloads...)
	return clone
}

func (repo *spkJobRepository) WithAssociations(associations ...string) SpkJobRepository {
	clone := repo.clone()
	clone.associations = append(clone.associations, associations...)
	return clone
}

func (repo *spkJobRepository) WithReplacements(replacements map[string]interface{}) SpkJobRepository {
	clone := repo.clone()
	clone.replacements = replacements
	return clone
}

func (repo *spkJobRepository) WithJoins(joins ...string) SpkJobRepository {
	clone := repo.clone()
	clone.joins = append(clone.joins, joins...)
	return clone
}

func (repo *spkJobRepository) WithWhere(query interface{}, args ...interface{}) SpkJobRepository {
	clone := repo.clone()
	clone.whereClauses = append(clone.whereClauses, func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
	return clone
}

func (repo *spkJobRepository) WithOrder(order string) SpkJobRepository {
	clone := repo.clone()
	clone.order = order
	return clone
}

func (repo *spkJobRepository) WithLimit(limit int) SpkJobRepository {
	clone := repo.clone()
	clone.limit = &limit
	return clone
}

func (repo *spkJobRepository) WithCursor(cursor int) SpkJobRepository {
	clone := repo.clone()
	clone.cursor = &cursor
	return clone
}

func (repo *spkJobRepository) getQueryBuilder() *builder.QueryBuilder[models.SpkJob] {
	qb := builder.NewQueryBuilder[models.SpkJob](repo.db).
		WithPreloads(repo.preloads...).
		WithAssociations(repo.associations...).
		WithReplacements(repo.replacements).
		WithJoins(repo.joins...).
		WithOrder(repo.order)

	for _, w := range repo.whereClauses {
		qb = qb.WithWhere(w)
	}
	if repo.limit != nil {
		qb = qb.WithLimit(*repo.limit)
	}
	if repo.cursor != nil {
		qb = qb.WithCursor(*repo.cursor)
	}
	return qb
}

func (repo *spkJobRepository) InsertSpkJob(data *models.SpkJob) (*models.SpkJob, error) {
	if err := repo.getQueryBuilder().Create(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *spkJobRepository) InsertManySpkJobs(data []*models.SpkJob) ([]*models.SpkJob, error) {
	if err := repo.getQueryBuilder().CreateMany(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *spkJobRepository) UpdateSpkJob(id int64, updates map[string]interface{}) (*models.SpkJob, error) {
	return repo.getQueryBuilder().UpdateByID(id, updates)
}

func (repo *spkJobRepository) UpdateManySpkJobs(ids []int64, updates map[string]interface{}) error {
	_, err := repo.getQueryBuilder().UpdateMany(ids, updates)
	return err
}

func (repo *spkJobRepository) RemoveSpkJob(id int64) error {
	return repo.getQueryBuilder().Delete(id)
}

func (repo *spkJobRepository) RemoveManySpkJobs(ids []int64) error {
	var job models.SpkJob
	return repo.db.Where("id IN ?", ids).Delete(&job).Error
}

func (repo *spkJobRepository) FindSpkJobs() ([]models.SpkJob, error) {
	r := repo
	if r.order == "" {
		if casted, ok := r.WithOrder("index ASC").(*spkJobRepository); ok {
			r = casted
		}
	}
	return repo.getQueryBuilder().FindAll()
}

func (repo *spkJobRepository) FindSpkJobByID(id int64) (*models.SpkJob, error) {
	return repo.getQueryBuilder().FindByID(id)
}

func (repo *spkJobRepository) FindSpkJobsByIDs(ids []int64) ([]*models.SpkJob, error) {
	if len(ids) == 0 {
		return []*models.SpkJob{}, nil
	}

	return repo.getQueryBuilder().WithWhere(func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN ?", ids)
	}).FindAllPtr()
}

func (repo *spkJobRepository) ReorderSpkJob(spkJobID int64, newIndex int, spkID int64) error {
	// Get current spk job
	var currentSpkJob models.SpkJob
	if err := repo.db.Where("id = ? AND spk_id = ?", spkJobID, spkID).First(&currentSpkJob).Error; err != nil {
		return err
	}

	oldIndex := int(currentSpkJob.Index)

	if oldIndex == newIndex {
		return nil
	}

	var maxIndexRow struct {
		MaxIndex int
	}
	if err := repo.db.Model(&models.SpkJob{}).
		Select("COALESCE(MAX(index), 0) as max_index").
		Where("spk_id = ?", spkID).
		Scan(&maxIndexRow).Error; err != nil {
		return err
	}

	maxIndex := maxIndexRow.MaxIndex

	if newIndex > maxIndex {
		newIndex = maxIndex
	}
	if newIndex < 1 {
		newIndex = 1
	}

	// Reorder logic
	if oldIndex < newIndex {
		// Moving down: shift items up between oldIndex+1 and newIndex
		result := repo.db.Model(&models.SpkJob{}).
			Where("spk_id = ? AND index > ? AND index <= ?", spkID, oldIndex, newIndex).
			UpdateColumn("index", gorm.Expr("index - 1"))

		if result.Error != nil {
			return result.Error
		}
		fmt.Printf("Shifted up %d rows (moving down)\n", result.RowsAffected)

	} else {
		// Moving up: shift items down between newIndex and oldIndex-1
		result := repo.db.Model(&models.SpkJob{}).
			Where("spk_id = ? AND index >= ? AND index < ?", spkID, newIndex, oldIndex).
			UpdateColumn("index", gorm.Expr("index + 1"))

		if result.Error != nil {
			return result.Error
		}
		fmt.Printf("Shifted down %d rows (moving up)\n", result.RowsAffected)
	}

	// Update the moved spk job
	result := repo.db.Model(&models.SpkJob{}).
		Where("id = ?", spkJobID).
		UpdateColumns(map[string]interface{}{
			"index":      newIndex,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	fmt.Printf("Updated target job: SpkJobID=%d to Index=%d\n", spkJobID, newIndex)
	return nil
}
