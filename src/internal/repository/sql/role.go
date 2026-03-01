package sql

import (
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/pkg/gorm/builder"

	"gorm.io/gorm"
)

type RoleRepository interface {
	WithTx(tx *gorm.DB) RoleRepository
	WithPreloads(preloads ...string) RoleRepository
	WithAssociations(associations ...string) RoleRepository
	WithReplacements(replacements map[string]interface{}) RoleRepository
	WithJoins(joins ...string) RoleRepository
	WithWhere(query interface{}, args ...interface{}) RoleRepository
	WithOrder(order string) RoleRepository
	WithLimit(limit int) RoleRepository
	WithCursor(cursor int) RoleRepository

	InsertRole(data *models.Role) (*models.Role, error)
	UpdateRole(id int64, updates map[string]interface{}) (*models.Role, error)
	UpdateManyRoles(ids []int64, updates map[string]interface{}) error
	RemoveRole(id int64) error
	RemoveManyRoles(ids []int64) error

	FindRole() ([]models.Role, error)
	FindRoleByID(id int64) (*models.Role, error)
}

type roleRepository struct {
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

func NewRoleRepository() RoleRepository {
	return &roleRepository{db: config.DB}
}

func (repo *roleRepository) clone() *roleRepository {
	clone := *repo
	return &clone
}

func (repo *roleRepository) WithTx(tx *gorm.DB) RoleRepository {
	clone := repo.clone()
	clone.db = tx
	return clone
}

func (repo *roleRepository) WithPreloads(preloads ...string) RoleRepository {
	clone := repo.clone()
	clone.preloads = append(clone.preloads, preloads...)
	return clone
}

func (repo *roleRepository) WithAssociations(associations ...string) RoleRepository {
	clone := repo.clone()
	clone.associations = append(clone.associations, associations...)
	return clone
}

func (repo *roleRepository) WithReplacements(replacements map[string]interface{}) RoleRepository {
	clone := repo.clone()
	clone.replacements = replacements
	return clone
}

func (repo *roleRepository) WithJoins(joins ...string) RoleRepository {
	clone := repo.clone()
	clone.joins = append(clone.joins, joins...)
	return clone
}

func (repo *roleRepository) WithWhere(query interface{}, args ...interface{}) RoleRepository {
	clone := repo.clone()
	clone.whereClauses = append(clone.whereClauses, func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
	return clone
}

func (repo *roleRepository) WithOrder(order string) RoleRepository {
	clone := repo.clone()
	clone.order = order
	return clone
}

func (repo *roleRepository) WithLimit(limit int) RoleRepository {
	clone := repo.clone()
	clone.limit = &limit
	return clone
}

func (repo *roleRepository) WithCursor(cursor int) RoleRepository {
	clone := repo.clone()
	clone.cursor = &cursor
	return clone
}

func (repo *roleRepository) getQueryBuilder() *builder.QueryBuilder[models.Role] {
	qb := builder.NewQueryBuilder[models.Role](repo.db).
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

func (repo *roleRepository) InsertRole(data *models.Role) (*models.Role, error) {
	if err := repo.getQueryBuilder().Create(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *roleRepository) UpdateRole(id int64, updates map[string]interface{}) (*models.Role, error) {
	return repo.getQueryBuilder().UpdateByID(id, updates)
}

func (repo *roleRepository) UpdateManyRoles(ids []int64, updates map[string]interface{}) error {
	_, err := repo.getQueryBuilder().UpdateMany(ids, updates)
	return err
}

func (repo *roleRepository) RemoveRole(id int64) error {
	return repo.getQueryBuilder().Delete(id)
}

func (repo *roleRepository) RemoveManyRoles(ids []int64) error {
	return repo.getQueryBuilder().WithWhere(func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN ?", ids)
	}).Delete(nil)
}

func (repo *roleRepository) FindRole() ([]models.Role, error) {
	return repo.getQueryBuilder().FindAll()
}

func (repo *roleRepository) FindRoleByID(id int64) (*models.Role, error) {
	return repo.getQueryBuilder().FindByID(id)
}
