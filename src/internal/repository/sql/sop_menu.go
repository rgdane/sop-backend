package sql

import (
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/pkg/gorm/builder"

	"gorm.io/gorm"
)

type SopMenuRepository interface {
	WithTx(tx *gorm.DB) SopMenuRepository
	WithPreloads(preloads ...string) SopMenuRepository
	WithAssociations(associations ...string) SopMenuRepository
	WithReplacements(replacements map[string]interface{}) SopMenuRepository
	WithJoins(joins ...string) SopMenuRepository
	WithWhere(query interface{}, args ...interface{}) SopMenuRepository
	WithOrder(order string) SopMenuRepository
	WithLimit(limit int) SopMenuRepository
	WithCursor(cursor int) SopMenuRepository

	InsertSopMenu(data *models.SopMenu) (*models.SopMenu, error)
	InsertManySopMenus(data []*models.SopMenu) ([]*models.SopMenu, error)

	UpdateSopMenu(id int64, updates map[string]interface{}) (*models.SopMenu, error)
	UpdateManySopMenus(ids []int64, updates map[string]interface{}) error
	RemoveSopMenu(id int64) error
	RemoveManySopMenus(ids []int64) error

	FindSopMenusByIDs(ids []int64) ([]*models.SopMenu, error)
	FindSopMenus() ([]models.SopMenu, error)
	FindSopMenuByID(id int64) (*models.SopMenu, error)
}

type sopMenuRepository struct {
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

func NewSopMenuRepository() SopMenuRepository {
	return &sopMenuRepository{db: config.DB}
}

func (repo *sopMenuRepository) clone() *sopMenuRepository {
	clone := *repo
	return &clone
}

func (repo *sopMenuRepository) WithTx(tx *gorm.DB) SopMenuRepository {
	clone := repo.clone()
	clone.db = tx
	return clone
}

func (repo *sopMenuRepository) WithPreloads(preloads ...string) SopMenuRepository {
	clone := repo.clone()
	clone.preloads = append(clone.preloads, preloads...)
	return clone
}

func (repo *sopMenuRepository) WithAssociations(associations ...string) SopMenuRepository {
	clone := repo.clone()
	clone.associations = append(clone.associations, associations...)
	return clone
}

func (repo *sopMenuRepository) WithReplacements(replacements map[string]interface{}) SopMenuRepository {
	clone := repo.clone()
	clone.replacements = replacements
	return clone
}

func (repo *sopMenuRepository) WithJoins(joins ...string) SopMenuRepository {
	clone := repo.clone()
	clone.joins = append(clone.joins, joins...)
	return clone
}

func (repo *sopMenuRepository) WithWhere(query interface{}, args ...interface{}) SopMenuRepository {
	clone := repo.clone()
	clone.whereClauses = append(clone.whereClauses, func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
	return clone
}

func (repo *sopMenuRepository) WithOrder(order string) SopMenuRepository {
	clone := repo.clone()
	clone.order = order
	return clone
}

func (repo *sopMenuRepository) WithLimit(limit int) SopMenuRepository {
	clone := repo.clone()
	clone.limit = &limit
	return clone
}

func (repo *sopMenuRepository) WithCursor(cursor int) SopMenuRepository {
	clone := repo.clone()
	clone.cursor = &cursor
	return clone
}

func (repo *sopMenuRepository) getQueryBuilder() *builder.QueryBuilder[models.SopMenu] {
	qb := builder.NewQueryBuilder[models.SopMenu](repo.db).
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

func (repo *sopMenuRepository) InsertSopMenu(data *models.SopMenu) (*models.SopMenu, error) {
	if err := repo.getQueryBuilder().Create(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *sopMenuRepository) InsertManySopMenus(data []*models.SopMenu) ([]*models.SopMenu, error) {
	if err := repo.getQueryBuilder().CreateMany(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *sopMenuRepository) UpdateSopMenu(id int64, updates map[string]interface{}) (*models.SopMenu, error) {
	return repo.getQueryBuilder().UpdateByID(id, updates)
}

func (repo *sopMenuRepository) UpdateManySopMenus(ids []int64, updates map[string]interface{}) error {
	_, err := repo.getQueryBuilder().UpdateMany(ids, updates)
	return err
}

func (repo *sopMenuRepository) RemoveSopMenu(id int64) error {
	return repo.getQueryBuilder().Delete(id)
}

func (repo *sopMenuRepository) RemoveManySopMenus(ids []int64) error {
	return repo.getQueryBuilder().Delete(ids)
}

func (repo *sopMenuRepository) FindSopMenus() ([]models.SopMenu, error) {
	return repo.getQueryBuilder().FindAll()
}

func (repo *sopMenuRepository) FindSopMenuByID(id int64) (*models.SopMenu, error) {
	return repo.getQueryBuilder().FindByID(id)
}

func (repo *sopMenuRepository) FindSopMenusByIDs(ids []int64) ([]*models.SopMenu, error) {
	if len(ids) == 0 {
		return []*models.SopMenu{}, nil
	}

	return repo.getQueryBuilder().WithWhere(func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN ?", ids)
	}).FindAllPtr()
}
