package sql

import (
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/pkg/gorm/builder"

	"gorm.io/gorm"
)

type PermissionRepository interface {
	WithTx(tx *gorm.DB) PermissionRepository
	WithPreloads(preloads ...string) PermissionRepository
	WithAssociations(associations ...string) PermissionRepository
	WithReplacements(replacements map[string]interface{}) PermissionRepository
	WithJoins(joins ...string) PermissionRepository
	WithWhere(query interface{}, args ...interface{}) PermissionRepository
	WithOrder(order string) PermissionRepository
	WithLimit(limit int) PermissionRepository
	WithCursor(cursor int) PermissionRepository

	InsertPermission(data *models.Permission) (*models.Permission, error)
	UpdatePermission(id int64, updates map[string]interface{}) (*models.Permission, error)
	UpdateManyPermissions(ids []int64, updates map[string]interface{}) error
	RemovePermission(id int64) error
	RemoveManyPermissions(ids []int64) error

	FindPermission() ([]models.Permission, error)
	FindPermissionByID(id int64) (*models.Permission, error)
}

type permissionRepository struct {
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

func NewPermissionRepository() PermissionRepository {
	return &permissionRepository{db: config.DB}
}

func (repo *permissionRepository) clone() *permissionRepository {
	clone := *repo
	return &clone
}

func (repo *permissionRepository) WithTx(tx *gorm.DB) PermissionRepository {
	clone := repo.clone()
	clone.db = tx
	return clone
}

func (repo *permissionRepository) WithPreloads(preloads ...string) PermissionRepository {
	clone := repo.clone()
	clone.preloads = append(clone.preloads, preloads...)
	return clone
}

func (repo *permissionRepository) WithAssociations(associations ...string) PermissionRepository {
	clone := repo.clone()
	clone.associations = append(clone.associations, associations...)
	return clone
}

func (repo *permissionRepository) WithReplacements(replacements map[string]interface{}) PermissionRepository {
	clone := repo.clone()
	clone.replacements = replacements
	return clone
}

func (repo *permissionRepository) WithJoins(joins ...string) PermissionRepository {
	clone := repo.clone()
	clone.joins = append(clone.joins, joins...)
	return clone
}

func (repo *permissionRepository) WithWhere(query interface{}, args ...interface{}) PermissionRepository {
	clone := repo.clone()
	clone.whereClauses = append(clone.whereClauses, func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
	return clone
}

func (repo *permissionRepository) WithOrder(order string) PermissionRepository {
	clone := repo.clone()
	clone.order = order
	return clone
}

func (repo *permissionRepository) WithLimit(limit int) PermissionRepository {
	clone := repo.clone()
	clone.limit = &limit
	return clone
}

func (repo *permissionRepository) WithCursor(cursor int) PermissionRepository {
	clone := repo.clone()
	clone.cursor = &cursor
	return clone
}

func (repo *permissionRepository) getQueryBuilder() *builder.QueryBuilder[models.Permission] {
	qb := builder.NewQueryBuilder[models.Permission](repo.db).
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

func (repo *permissionRepository) InsertPermission(data *models.Permission) (*models.Permission, error) {
	if err := repo.getQueryBuilder().Create(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *permissionRepository) UpdatePermission(id int64, updates map[string]interface{}) (*models.Permission, error) {
	return repo.getQueryBuilder().UpdateByID(id, updates)
}

func (repo *permissionRepository) UpdateManyPermissions(ids []int64, updates map[string]interface{}) error {
	_, err := repo.getQueryBuilder().UpdateMany(ids, updates)
	return err
}

func (repo *permissionRepository) RemovePermission(id int64) error {
	return repo.getQueryBuilder().Delete(id)
}

func (repo *permissionRepository) RemoveManyPermissions(ids []int64) error {
	return repo.getQueryBuilder().
		WithWhere(func(db *gorm.DB) *gorm.DB {
			return db.Where("id IN ?", ids)
		}).Delete(nil)
}

func (repo *permissionRepository) FindPermission() ([]models.Permission, error) {
	return repo.getQueryBuilder().FindAll()
}

func (repo *permissionRepository) FindPermissionByID(id int64) (*models.Permission, error) {
	return repo.getQueryBuilder().FindByID(id)
}
