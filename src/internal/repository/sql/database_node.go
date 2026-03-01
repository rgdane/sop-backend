package sql

import (
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/pkg/gorm/builder"

	"gorm.io/gorm"
)

type DatabaseNodeRepository interface {
	WithTx(tx *gorm.DB) DatabaseNodeRepository
	WithPreloads(preloads ...string) DatabaseNodeRepository
	WithAssociations(associations ...string) DatabaseNodeRepository
	WithReplacements(replacements map[string]interface{}) DatabaseNodeRepository
	WithJoins(joins ...string) DatabaseNodeRepository
	WithWhere(query interface{}, args ...interface{}) DatabaseNodeRepository
	WithOrder(order string) DatabaseNodeRepository
	WithLimit(limit int) DatabaseNodeRepository
	WithCursor(cursor int) DatabaseNodeRepository
	WithUnscoped() DatabaseNodeRepository

	InsertDatabaseNode(data *models.DatabaseNode) (*models.DatabaseNode, error)
	InsertManyDatabaseNodes(data []*models.DatabaseNode) ([]*models.DatabaseNode, error)
	UpdateDatabaseNode(id int64, updates map[string]interface{}) (*models.DatabaseNode, error)
	UpdateManyDatabaseNodes(ids []int64, updates map[string]interface{}) error
	RemoveDatabaseNode(id int64) error
	RemoveManyDatabaseNodes(ids []int64) error
	FindDatabaseNodes() ([]models.DatabaseNode, error)
	FindDatabaseNodeByID(id int64) (*models.DatabaseNode, error)
	FindDatabaseNodesByIDs(ids []int64) ([]*models.DatabaseNode, error)
	CountDatabaseNodes() (int64, error)
}

type databaseNodeRepository struct {
	db           *gorm.DB
	replacements map[string]interface{}
	preloads     []string
	associations []string
	joins        []string
	whereClauses []func(*gorm.DB) *gorm.DB
	order        string
	limit        *int
	cursor       *int
	unscoped     bool
}

func NewDatabaseNodeRepository() DatabaseNodeRepository {
	return &databaseNodeRepository{db: config.DB}
}

// --- Chainable Config Methods ---

func (repo *databaseNodeRepository) clone() *databaseNodeRepository {
	clone := *repo
	return &clone
}

func (repo *databaseNodeRepository) WithTx(tx *gorm.DB) DatabaseNodeRepository {
	clone := repo.clone()
	clone.db = tx
	return clone
}

func (repo *databaseNodeRepository) WithPreloads(preloads ...string) DatabaseNodeRepository {
	clone := repo.clone()
	clone.preloads = append(clone.preloads, preloads...)
	return clone
}

func (repo *databaseNodeRepository) WithAssociations(associations ...string) DatabaseNodeRepository {
	clone := repo.clone()
	clone.associations = append(clone.associations, associations...)
	return clone
}

func (repo *databaseNodeRepository) WithReplacements(replacements map[string]interface{}) DatabaseNodeRepository {
	clone := repo.clone()
	clone.replacements = replacements
	return clone
}

func (repo *databaseNodeRepository) WithJoins(joins ...string) DatabaseNodeRepository {
	clone := repo.clone()
	clone.joins = append(clone.joins, joins...)
	return clone
}

func (repo *databaseNodeRepository) WithWhere(query interface{}, args ...interface{}) DatabaseNodeRepository {
	clone := repo.clone()
	clone.whereClauses = append(clone.whereClauses, func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
	return clone
}

func (repo *databaseNodeRepository) WithOrder(order string) DatabaseNodeRepository {
	clone := repo.clone()
	clone.order = order
	return clone
}

func (repo *databaseNodeRepository) WithLimit(limit int) DatabaseNodeRepository {
	clone := repo.clone()
	clone.limit = &limit
	return clone
}

func (repo *databaseNodeRepository) WithCursor(cursor int) DatabaseNodeRepository {
	clone := repo.clone()
	clone.cursor = &cursor
	return clone
}

func (repo *databaseNodeRepository) WithUnscoped() DatabaseNodeRepository {
	clone := repo.clone()
	clone.unscoped = true
	return clone
}

// --- Builder Helper ---

func (repo *databaseNodeRepository) getQueryBuilder() *builder.QueryBuilder[models.DatabaseNode] {
	db := repo.db
	if repo.unscoped {
		db = db.Unscoped()
	}

	qb := builder.NewQueryBuilder[models.DatabaseNode](db).
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

func (repo *databaseNodeRepository) InsertDatabaseNode(data *models.DatabaseNode) (*models.DatabaseNode, error) {
	if err := repo.getQueryBuilder().Create(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *databaseNodeRepository) InsertManyDatabaseNodes(data []*models.DatabaseNode) ([]*models.DatabaseNode, error) {
	if err := repo.getQueryBuilder().CreateMany(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *databaseNodeRepository) UpdateDatabaseNode(id int64, updates map[string]interface{}) (*models.DatabaseNode, error) {
	return repo.getQueryBuilder().UpdateByID(id, updates)
}

func (repo *databaseNodeRepository) UpdateManyDatabaseNodes(ids []int64, updates map[string]interface{}) error {
	_, err := repo.getQueryBuilder().UpdateMany(ids, updates)
	return err
}

func (repo *databaseNodeRepository) RemoveDatabaseNode(id int64) error {
	db := repo.db
	if repo.unscoped {
		db = db.Unscoped()
	}
	model := models.DatabaseNode{}

	model.ID = id

	if err := db.Delete(&model).Error; err != nil {
		return err
	}
	return nil
}

func (repo *databaseNodeRepository) RemoveManyDatabaseNodes(ids []int64) error {
	return repo.getQueryBuilder().WithWhere(func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN ?", ids)
	}).Delete(nil)
}

func (repo *databaseNodeRepository) FindDatabaseNodes() ([]models.DatabaseNode, error) {
	return repo.getQueryBuilder().FindAll()
}

func (repo *databaseNodeRepository) FindDatabaseNodeByID(id int64) (*models.DatabaseNode, error) {
	return repo.getQueryBuilder().FindByID(id)
}

func (repo *databaseNodeRepository) FindDatabaseNodesByIDs(ids []int64) ([]*models.DatabaseNode, error) {
	if len(ids) == 0 {
		return []*models.DatabaseNode{}, nil
	}

	return repo.getQueryBuilder().WithWhere(func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN ?", ids)
	}).FindAllPtr()
}

func (repo *databaseNodeRepository) CountDatabaseNodes() (int64, error) {
	return repo.getQueryBuilder().Count()
}
