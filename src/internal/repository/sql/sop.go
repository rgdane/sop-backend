package sql

import (
	"fmt"
	"jk-api/pkg/gorm/builder"

	"jk-api/internal/config"
	"jk-api/internal/database/models"

	"gorm.io/gorm"
)

type SopRepository interface {
	WithTx(tx *gorm.DB) SopRepository
	WithPreloads(preloads ...string) SopRepository
	WithAssociations(associations ...string) SopRepository
	WithReplacements(replacements map[string]interface{}) SopRepository
	WithJoins(joins ...string) SopRepository
	WithWhere(query interface{}, args ...interface{}) SopRepository
	WithOrder(order string) SopRepository
	WithLimit(limit int) SopRepository
	WithCursor(cursor int) SopRepository
	WithUnscoped() SopRepository

	InsertSop(data *models.Sop) (*models.Sop, error)
	InsertManySops(data []*models.Sop) ([]*models.Sop, error)
	UpdateSop(id int64, updates map[string]interface{}) (*models.Sop, error)
	UpdateManySops(ids []int64, updates map[string]interface{}) error
	RemoveSop(id int64) error
	RemoveManySops(ids []int64) error

	FindSops() ([]models.Sop, error)
	FindSopByID(id int64) (*models.Sop, error)
	FindSopsByIDs(ids []int64) ([]*models.Sop, error)
	CountSops() (int64, error)
}

type sopRepository struct {
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

func NewSopRepository() SopRepository {
	return &sopRepository{db: config.DB}
}

func (repo *sopRepository) clone() *sopRepository {
	clone := *repo
	return &clone
}

func (repo *sopRepository) WithTx(tx *gorm.DB) SopRepository {
	clone := repo.clone()
	clone.db = tx
	return clone
}

func (repo *sopRepository) WithPreloads(preloads ...string) SopRepository {
	clone := repo.clone()
	clone.preloads = append(clone.preloads, preloads...)
	return clone
}

func (repo *sopRepository) WithAssociations(associations ...string) SopRepository {
	clone := repo.clone()
	clone.associations = append(clone.associations, associations...)
	return clone
}

func (repo *sopRepository) WithReplacements(replacements map[string]interface{}) SopRepository {
	clone := repo.clone()
	clone.replacements = replacements
	return clone
}

func (repo *sopRepository) WithJoins(joins ...string) SopRepository {
	clone := repo.clone()
	clone.joins = append(clone.joins, joins...)
	return clone
}

func (repo *sopRepository) WithWhere(query interface{}, args ...interface{}) SopRepository {
	clone := repo.clone()
	clone.whereClauses = append(clone.whereClauses, func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
	return clone
}

func (repo *sopRepository) WithOrder(order string) SopRepository {
	clone := repo.clone()
	clone.order = order
	return clone
}

func (repo *sopRepository) WithLimit(limit int) SopRepository {
	clone := repo.clone()
	clone.limit = &limit
	return clone
}

func (repo *sopRepository) WithCursor(cursor int) SopRepository {
	fmt.Println("with cursor: ", cursor)
	clone := repo.clone()
	clone.cursor = &cursor
	return clone
}

func (repo *sopRepository) WithUnscoped() SopRepository {
	clone := repo.clone()
	clone.unscoped = true
	return clone
}

func (repo *sopRepository) getQueryBuilder() *builder.QueryBuilder[models.Sop] {
	db := repo.db
	if repo.unscoped {
		db = db.Unscoped()
	}

	qb := builder.NewQueryBuilder[models.Sop](db).
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

func (repo *sopRepository) InsertSop(data *models.Sop) (*models.Sop, error) {
	if err := repo.getQueryBuilder().Create(data); err != nil {
		return nil, err
	}

	return data, nil
}

func (repo *sopRepository) InsertManySops(data []*models.Sop) ([]*models.Sop, error) {
	if err := repo.getQueryBuilder().CreateMany(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *sopRepository) UpdateSop(id int64, updates map[string]interface{}) (*models.Sop, error) {
	return repo.getQueryBuilder().UpdateByID(id, updates)
}

func (repo *sopRepository) UpdateManySops(ids []int64, updates map[string]interface{}) error {
	_, err := repo.getQueryBuilder().UpdateMany(ids, updates)
	return err
}

func (repo *sopRepository) RemoveSop(id int64) error {
	return repo.getQueryBuilder().Delete(id)
}

func (repo *sopRepository) RemoveManySops(ids []int64) error {
	return repo.getQueryBuilder().Delete(ids)
}

func (repo *sopRepository) FindSops() ([]models.Sop, error) {
	return repo.getQueryBuilder().FindAll()
}

func (repo *sopRepository) FindSopByID(id int64) (*models.Sop, error) {
	return repo.getQueryBuilder().FindByID(id)
}

func (repo *sopRepository) FindSopsByIDs(ids []int64) ([]*models.Sop, error) {
	if len(ids) == 0 {
		return []*models.Sop{}, nil
	}

	return repo.getQueryBuilder().WithWhere(func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN ?", ids)
	}).FindAllPtr()
}

func (repo *sopRepository) CountSops() (int64, error) {
	var count int64
	err := repo.db.Model(&models.Sop{}).
		Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}
