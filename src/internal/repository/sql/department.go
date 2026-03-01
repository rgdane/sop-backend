package sql

import (
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/pkg/gorm/builder"

	"gorm.io/gorm"
)

type DepartmentRepository interface {
	WithTx(tx *gorm.DB) DepartmentRepository
	WithPreloads(preloads ...string) DepartmentRepository
	WithAssociations(associations ...string) DepartmentRepository
	WithReplacements(replacements map[string]interface{}) DepartmentRepository
	WithJoins(joins ...string) DepartmentRepository
	WithWhere(query interface{}, args ...interface{}) DepartmentRepository
	WithOrder(order string) DepartmentRepository
	WithLimit(limit int) DepartmentRepository
	WithCursor(cursor int) DepartmentRepository
	WithUnscoped() DepartmentRepository

	InsertDepartment(data *models.Department) (*models.Department, error)
	InsertManyDepartments(data []*models.Department) ([]*models.Department, error)
	UpdateDepartment(id int64, updates map[string]interface{}) (*models.Department, error)
	UpdateManyDepartments(ids []int64, updates map[string]interface{}) error
	RemoveDepartment(id int64) error
	RemoveManyDepartments(ids []int64) error

	FindDepartment() ([]models.Department, error)
	FindDepartmentByID(id int64) (*models.Department, error)
	FindDepartmentsByIDs(ids []int64) ([]*models.Department, error)
}

type departmentRepository struct {
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

func NewDepartmentRepository() DepartmentRepository {
	return &departmentRepository{db: config.DB}
}

func (repo *departmentRepository) clone() *departmentRepository {
	clone := *repo
	return &clone
}

func (repo *departmentRepository) WithTx(tx *gorm.DB) DepartmentRepository {
	clone := repo.clone()
	clone.db = tx
	return clone
}

func (repo *departmentRepository) WithPreloads(preloads ...string) DepartmentRepository {
	clone := repo.clone()
	clone.preloads = append(clone.preloads, preloads...)
	return clone
}

func (repo *departmentRepository) WithAssociations(associations ...string) DepartmentRepository {
	clone := repo.clone()
	clone.associations = append(clone.associations, associations...)
	return clone
}

func (repo *departmentRepository) WithReplacements(replacements map[string]interface{}) DepartmentRepository {
	clone := repo.clone()
	clone.replacements = replacements
	return clone
}

func (repo *departmentRepository) WithJoins(joins ...string) DepartmentRepository {
	clone := repo.clone()
	clone.joins = append(clone.joins, joins...)
	return clone
}

func (repo *departmentRepository) WithWhere(query interface{}, args ...interface{}) DepartmentRepository {
	clone := repo.clone()
	clone.whereClauses = append(clone.whereClauses, func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
	return clone
}

func (repo *departmentRepository) WithOrder(order string) DepartmentRepository {
	clone := repo.clone()
	clone.order = order
	return clone
}

func (repo *departmentRepository) WithLimit(limit int) DepartmentRepository {
	clone := repo.clone()
	clone.limit = &limit
	return clone
}

func (repo *departmentRepository) WithCursor(cursor int) DepartmentRepository {
	clone := repo.clone()
	clone.cursor = &cursor
	return clone
}

func (repo *departmentRepository) WithUnscoped() DepartmentRepository {
	clone := repo.clone()
	clone.unscoped = true
	return clone
}

func (repo *departmentRepository) getQueryBuilder() *builder.QueryBuilder[models.Department] {
	db := repo.db
	if repo.unscoped {
		db = db.Unscoped()
	}

	qb := builder.NewQueryBuilder[models.Department](db).
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

func (repo *departmentRepository) InsertDepartment(data *models.Department) (*models.Department, error) {
	if err := repo.getQueryBuilder().Create(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *departmentRepository) InsertManyDepartments(data []*models.Department) ([]*models.Department, error) {
	if err := repo.getQueryBuilder().CreateMany(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *departmentRepository) UpdateDepartment(id int64, updates map[string]interface{}) (*models.Department, error) {
	return repo.getQueryBuilder().UpdateByID(id, updates)
}

func (repo *departmentRepository) UpdateManyDepartments(ids []int64, updates map[string]interface{}) error {
	_, err := repo.getQueryBuilder().UpdateMany(ids, updates)
	return err
}

func (repo *departmentRepository) RemoveDepartment(id int64) error {
	return repo.getQueryBuilder().Delete(id)
}

func (repo *departmentRepository) RemoveManyDepartments(ids []int64) error {
	return repo.getQueryBuilder().Delete(ids)
}

func (repo *departmentRepository) FindDepartment() ([]models.Department, error) {
	return repo.getQueryBuilder().FindAll()
}

func (repo *departmentRepository) FindDepartmentByID(id int64) (*models.Department, error) {
	return repo.getQueryBuilder().FindByID(id)
}

func (repo *departmentRepository) FindDepartmentsByIDs(ids []int64) ([]*models.Department, error) {
	if len(ids) == 0 {
		return []*models.Department{}, nil
	}

	return repo.getQueryBuilder().WithWhere(func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN ?", ids)
	}).FindAllPtr()
}
