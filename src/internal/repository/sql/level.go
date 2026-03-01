package sql

import (
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/pkg/gorm/builder"

	"gorm.io/gorm"
)

type LevelRepository interface {
	WithTx(tx *gorm.DB) LevelRepository
	WithPreloads(preloads ...string) LevelRepository
	WithAssociations(associations ...string) LevelRepository
	WithReplacements(replacements map[string]interface{}) LevelRepository
	WithJoins(joins ...string) LevelRepository
	WithWhere(query interface{}, args ...interface{}) LevelRepository
	WithOrder(order string) LevelRepository
	WithLimit(limit int) LevelRepository
	WithCursor(cursor int) LevelRepository
	WithUnscoped() LevelRepository

	InsertLevel(data *models.Level) (*models.Level, error)
	InsertManyLevels(data []*models.Level) ([]*models.Level, error)
	UpdateLevel(id int64, updates map[string]interface{}) (*models.Level, error)
	UpdateManyLevels(ids []int64, updates map[string]interface{}) error
	RemoveLevel(id int64) error
	RemoveManyLevels(ids []int64) error

	FindLevel() ([]models.Level, error)
	FindLevelByID(id int64) (*models.Level, error)
	FindLevelsByIDs(ids []int64) ([]*models.Level, error)
}

type levelRepository struct {
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

func NewLevelRepository() LevelRepository {
	return &levelRepository{db: config.DB}
}

// --- Chainable Configs ---

func (repo *levelRepository) clone() *levelRepository {
	clone := *repo
	return &clone
}

func (repo *levelRepository) WithTx(tx *gorm.DB) LevelRepository {
	clone := repo.clone()
	clone.db = tx
	return clone
}

func (repo *levelRepository) WithPreloads(preloads ...string) LevelRepository {
	clone := repo.clone()
	clone.preloads = append(clone.preloads, preloads...)
	return clone
}

func (repo *levelRepository) WithAssociations(associations ...string) LevelRepository {
	clone := repo.clone()
	clone.associations = append(clone.associations, associations...)
	return clone
}

func (repo *levelRepository) WithReplacements(replacements map[string]interface{}) LevelRepository {
	clone := repo.clone()
	clone.replacements = replacements
	return clone
}

func (repo *levelRepository) WithJoins(joins ...string) LevelRepository {
	clone := repo.clone()
	clone.joins = append(clone.joins, joins...)
	return clone
}

func (repo *levelRepository) WithWhere(query interface{}, args ...interface{}) LevelRepository {
	clone := repo.clone()
	clone.whereClauses = append(clone.whereClauses, func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
	return clone
}

func (repo *levelRepository) WithOrder(order string) LevelRepository {
	clone := repo.clone()
	clone.order = order
	return clone
}

func (repo *levelRepository) WithLimit(limit int) LevelRepository {
	clone := repo.clone()
	clone.limit = &limit
	return clone
}

func (repo *levelRepository) WithCursor(cursor int) LevelRepository {
	clone := repo.clone()
	clone.cursor = &cursor
	return clone
}

func (repo *levelRepository) WithUnscoped() LevelRepository {
	clone := repo.clone()
	clone.unscoped = true
	return clone
}

// --- Query Builder Helper ---

func (repo *levelRepository) getQueryBuilder() *builder.QueryBuilder[models.Level] {
	db := repo.db
	if repo.unscoped {
		db = db.Unscoped()
	}

	qb := builder.NewQueryBuilder[models.Level](db).
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

// --- CRUD Methods ---

func (repo *levelRepository) InsertLevel(data *models.Level) (*models.Level, error) {
	if err := repo.getQueryBuilder().Create(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *levelRepository) InsertManyLevels(data []*models.Level) ([]*models.Level, error) {
	if err := repo.getQueryBuilder().CreateMany(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *levelRepository) UpdateLevel(id int64, updates map[string]interface{}) (*models.Level, error) {
	return repo.getQueryBuilder().UpdateByID(id, updates)
}

func (repo *levelRepository) UpdateManyLevels(ids []int64, updates map[string]interface{}) error {
	_, err := repo.getQueryBuilder().UpdateMany(ids, updates)
	return err
}

func (repo *levelRepository) RemoveLevel(id int64) error {
	return repo.getQueryBuilder().Delete(id)
}

func (repo *levelRepository) RemoveManyLevels(ids []int64) error {
	return repo.getQueryBuilder().Delete(ids)
}

func (repo *levelRepository) FindLevel() ([]models.Level, error) {
	return repo.getQueryBuilder().FindAll()
}

func (repo *levelRepository) FindLevelByID(id int64) (*models.Level, error) {
	return repo.getQueryBuilder().FindByID(id)
}

func (repo *levelRepository) FindLevelsByIDs(ids []int64) ([]*models.Level, error) {
	if len(ids) == 0 {
		return []*models.Level{}, nil
	}

	return repo.getQueryBuilder().WithWhere(func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN ?", ids)
	}).FindAllPtr()
}
