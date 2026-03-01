package sql

import (
	"errors"
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/pkg/gorm/builder"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepository interface {
	WithTx(tx *gorm.DB) UserRepository
	WithPreloads(preloads ...string) UserRepository
	WithAssociations(associations ...string) UserRepository
	WithReplacements(replacements map[string]interface{}) UserRepository
	WithJoins(joins ...string) UserRepository
	WithWhere(query interface{}, args ...interface{}) UserRepository
	WithOrder(order string) UserRepository
	WithLimit(limit int) UserRepository
	WithCursor(cursor int) UserRepository
	WithUnscoped() UserRepository

	InsertUser(data *models.User) (*models.User, error)
	InsertManyUsers(data []*models.User) ([]*models.User, error)

	UpdateUser(id int64, updates map[string]interface{}) (*models.User, error)
	UpdateManyUsers(ids []int64, updates map[string]interface{}) error
	RemoveUser(id int64) error
	RemoveManyUsers(ids []int64) error

	FindUser() ([]models.User, error)
	FindUserByID(id int64) (*models.User, error)
	FindUserByEmail(email string) (*models.User, error)
}

type userRepository struct {
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

func NewUserRepository() UserRepository {
	return &userRepository{db: config.DB}
}

// --- Chainable Configs ---

func (repo *userRepository) clone() *userRepository {
	clone := *repo
	return &clone
}

func (repo *userRepository) WithTx(tx *gorm.DB) UserRepository {
	clone := repo.clone()
	clone.db = tx
	return clone
}

func (repo *userRepository) WithPreloads(preloads ...string) UserRepository {
	clone := repo.clone()
	clone.preloads = append(clone.preloads, preloads...)
	return clone
}

func (repo *userRepository) WithAssociations(associations ...string) UserRepository {
	clone := repo.clone()
	clone.associations = append(clone.associations, associations...)
	return clone
}

func (repo *userRepository) WithReplacements(replacements map[string]interface{}) UserRepository {
	clone := repo.clone()
	clone.replacements = replacements
	return clone
}

func (repo *userRepository) WithJoins(joins ...string) UserRepository {
	clone := repo.clone()
	clone.joins = append(clone.joins, joins...)
	return clone
}

func (repo *userRepository) WithWhere(query interface{}, args ...interface{}) UserRepository {
	clone := repo.clone()
	clone.whereClauses = append(clone.whereClauses, func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
	return clone
}

func (repo *userRepository) WithOrder(order string) UserRepository {
	clone := repo.clone()
	clone.order = order
	return clone
}

func (repo *userRepository) WithLimit(limit int) UserRepository {
	clone := repo.clone()
	clone.limit = &limit
	return clone
}

func (repo *userRepository) WithCursor(cursor int) UserRepository {
	clone := repo.clone()
	clone.cursor = &cursor
	return clone
}

func (repo *userRepository) WithUnscoped() UserRepository {
	clone := repo.clone()
	clone.unscoped = true
	return clone
}

// --- Builder ---

func (repo *userRepository) getQueryBuilder() *builder.QueryBuilder[models.User] {
	db := repo.db
	if repo.unscoped {
		db = db.Unscoped()
	}

	qb := builder.NewQueryBuilder[models.User](db).
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

// --- CRUD ---

func (repo *userRepository) InsertUser(data *models.User) (*models.User, error) {
	if err := repo.getQueryBuilder().Create(data); err != nil {
		return nil, err
	}
	if !data.IsPasswordDefault {
		data.Password = ""
	}
	return data, nil
}

func (repo *userRepository) InsertManyUsers(data []*models.User) ([]*models.User, error) {
	if err := repo.getQueryBuilder().CreateMany(data); err != nil {
		return nil, err
	}
	for i := range data {
		if !data[i].IsPasswordDefault {
			data[i].Password = ""
		}
	}
	return data, nil
}

func (repo *userRepository) UpdateUser(id int64, updates map[string]interface{}) (*models.User, error) {
	if updates["old_password"] == nil {
		data, err := repo.getQueryBuilder().UpdateByID(id, updates)
		if err != nil {
			return nil, err
		}
		if !data.IsPasswordDefault {
			data.Password = ""
		}
		return data, nil
	}

	result, err := repo.getQueryBuilder().FindByID(id)
	if err != nil {
		return nil, err
	}
	originPwd := result.Password

	if bcrypt.CompareHashAndPassword([]byte(originPwd), []byte(updates["old_password"].(string))) != nil {
		return nil, errors.New("Password lama tidak sesuai")
	}

	updates["password"] = updates["new_password"]

	delete(updates, "new_password")
	delete(updates, "old_password")

	data, err := repo.getQueryBuilder().UpdateByID(id, updates)
	if err != nil {
		return nil, err
	}
	if !data.IsPasswordDefault {
		data.Password = ""
	}
	return data, nil
}

func (repo *userRepository) UpdateManyUsers(ids []int64, updates map[string]interface{}) error {
	_, err := repo.getQueryBuilder().UpdateMany(ids, updates)
	return err
}

func (repo *userRepository) RemoveUser(id int64) error {
	return repo.getQueryBuilder().Delete(id)
}

func (repo *userRepository) RemoveManyUsers(ids []int64) error {
	return repo.getQueryBuilder().Delete(ids)
}

// --- Finders ---

func (repo *userRepository) FindUser() ([]models.User, error) {
	data, err := repo.getQueryBuilder().FindAll()
	if err != nil {
		return nil, err
	}
	for i := range data {
		if !data[i].IsPasswordDefault {
			data[i].Password = ""
		}
	}
	return data, nil
}

func (repo *userRepository) FindUserByID(id int64) (*models.User, error) {
	data, err := repo.getQueryBuilder().FindByID(id)
	if err != nil {
		return nil, err
	}
	if !data.IsPasswordDefault {
		data.Password = ""
	}
	return data, nil
}

func (repo *userRepository) FindUserByEmail(email string) (*models.User, error) {
	return repo.getQueryBuilder().
		WithWhere(func(db *gorm.DB) *gorm.DB {
			return db.Where("email = ?", email)
		}).FindOne()
}