package sql

import (
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/pkg/gorm/builder"

	"gorm.io/gorm"
)

type SpkRepository interface {
	WithTx(tx *gorm.DB) SpkRepository
	WithPreloads(preloads ...string) SpkRepository
	WithAssociations(associations ...string) SpkRepository
	WithReplacements(replacements map[string]interface{}) SpkRepository
	WithJoins(joins ...string) SpkRepository
	WithWhere(query interface{}, args ...interface{}) SpkRepository
	WithOrder(order string) SpkRepository
	WithLimit(limit int) SpkRepository
	WithCursor(cursor int) SpkRepository
	WithUnscoped() SpkRepository

	InsertSpk(data *models.Spk) (*models.Spk, error)
	InsertManySpks(data []*models.Spk) ([]*models.Spk, error)
	UpdateSpk(id int64, updates map[string]interface{}) (*models.Spk, error)
	UpdateManySpks(ids []int64, updates map[string]interface{}) error
	RemoveSpk(id int64) error
	RemoveManySpks(ids []int64) error

	FindSpk() ([]models.Spk, error)
	FindSpkByID(id int64) (*models.Spk, error)
	FindSpksByIDs(ids []int64) ([]*models.Spk, error)
}

type spkRepository struct {
	db           *gorm.DB
	preloads     []string
	associations []string
	replacements map[string]interface{}
	joins        []string
	whereClauses []func(*gorm.DB) *gorm.DB
	order        string
	limit        *int
	cursor       *int
	unscoped     bool
}

func NewSpkRepository() SpkRepository {
	return &spkRepository{db: config.DB}
}

func (repo *spkRepository) clone() *spkRepository {
	clone := *repo
	return &clone
}

func (repo *spkRepository) WithTx(tx *gorm.DB) SpkRepository {
	clone := repo.clone()
	clone.db = tx
	return clone
}

func (repo *spkRepository) WithPreloads(preloads ...string) SpkRepository {
	clone := repo.clone()
	clone.preloads = append(clone.preloads, preloads...)
	return clone
}

func (repo *spkRepository) WithAssociations(associations ...string) SpkRepository {
	clone := repo.clone()
	clone.associations = append(clone.associations, associations...)
	return clone
}

func (repo *spkRepository) WithReplacements(replacements map[string]interface{}) SpkRepository {
	clone := repo.clone()
	clone.replacements = replacements
	return clone
}

func (repo *spkRepository) WithJoins(joins ...string) SpkRepository {
	clone := repo.clone()
	clone.joins = append(clone.joins, joins...)
	return clone
}

func (repo *spkRepository) WithWhere(query interface{}, args ...interface{}) SpkRepository {
	clone := repo.clone()
	clone.whereClauses = append(clone.whereClauses, func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
	return clone
}

func (repo *spkRepository) WithOrder(order string) SpkRepository {
	clone := repo.clone()
	clone.order = order
	return clone
}

func (repo *spkRepository) WithLimit(limit int) SpkRepository {
	clone := repo.clone()
	clone.limit = &limit
	return clone
}

func (repo *spkRepository) WithCursor(cursor int) SpkRepository {
	clone := repo.clone()
	clone.cursor = &cursor
	return clone
}

func (repo *spkRepository) WithUnscoped() SpkRepository {
	clone := repo.clone()
	clone.unscoped = true
	return clone
}

func (repo *spkRepository) getQueryBuilder() *builder.QueryBuilder[models.Spk] {
	db := repo.db
	if repo.unscoped {
		db = db.Unscoped()
	}

	qb := builder.NewQueryBuilder[models.Spk](db).
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

func (repo *spkRepository) InsertSpk(data *models.Spk) (*models.Spk, error) {
	if err := repo.getQueryBuilder().Create(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *spkRepository) InsertManySpks(data []*models.Spk) ([]*models.Spk, error) {
	if err := repo.getQueryBuilder().CreateMany(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *spkRepository) UpdateSpk(id int64, updates map[string]interface{}) (*models.Spk, error) {
	return repo.getQueryBuilder().UpdateByID(id, updates)
}

func (repo *spkRepository) UpdateManySpks(ids []int64, updates map[string]interface{}) error {
	_, err := repo.getQueryBuilder().UpdateMany(ids, updates)
	return err
}

func (repo *spkRepository) RemoveSpk(id int64) error {
	return repo.getQueryBuilder().Delete(id)
}

func (repo *spkRepository) RemoveManySpks(ids []int64) error {
	var Spk models.Spk
	return repo.db.Where("id IN ?", ids).Delete(&Spk).Error
}

func (repo *spkRepository) FindSpk() ([]models.Spk, error) {
	return repo.getQueryBuilder().FindAll()
}

func (repo *spkRepository) FindSpkByID(id int64) (*models.Spk, error) {
	return repo.getQueryBuilder().FindByID(id)
}

func (repo *spkRepository) FindSpksByIDs(ids []int64) ([]*models.Spk, error) {
	if len(ids) == 0 {
		return []*models.Spk{}, nil
	}

	return repo.getQueryBuilder().WithWhere(func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN ?", ids)
	}).FindAllPtr()
}
