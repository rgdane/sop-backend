package builder

import (
	"fmt"
	"jk-api/internal/shared/helper"

	"gorm.io/gorm"
)

type QueryBuilder[T any] struct {
	logParams    *helper.LogParams
	db           *gorm.DB
	preloads     []string
	joins        []string
	whereClauses []func(*gorm.DB) *gorm.DB
	orderBy      string
	limit        *int
	cursor       *int
	associations []string
	replacements map[string]interface{}
	unscoped     bool
	selectFields []string
}

func NewQueryBuilder[T any](db *gorm.DB) *QueryBuilder[T] {
	return &QueryBuilder[T]{db: db}
}

// ---------- Configuration Methods ----------

func (qb *QueryBuilder[T]) WithPreloads(preloads ...string) *QueryBuilder[T] {
	qb.preloads = append(qb.preloads, preloads...)
	return qb
}

func (qb *QueryBuilder[T]) WithJoins(joins ...string) *QueryBuilder[T] {
	qb.joins = append(qb.joins, joins...)
	return qb
}

func (qb *QueryBuilder[T]) WithWhere(fn func(*gorm.DB) *gorm.DB) *QueryBuilder[T] {
	qb.whereClauses = append(qb.whereClauses, fn)
	return qb
}

func (qb *QueryBuilder[T]) WithOrder(order string) *QueryBuilder[T] {
	qb.orderBy = order
	return qb
}

func (qb *QueryBuilder[T]) WithLimit(l int) *QueryBuilder[T] {
	qb.limit = &l
	return qb
}

func (qb *QueryBuilder[T]) WithCursor(c int) *QueryBuilder[T] {
	qb.cursor = &c
	return qb
}

func (qb *QueryBuilder[T]) WithAssociations(assocs ...string) *QueryBuilder[T] {
	qb.associations = append(qb.associations, assocs...)
	return qb
}

func (qb *QueryBuilder[T]) WithReplacements(replacements map[string]interface{}) *QueryBuilder[T] {
	qb.replacements = replacements
	return qb
}

func (qb *QueryBuilder[T]) WithUnscoped() *QueryBuilder[T] {
	qb.unscoped = true
	return qb
}

func (qb *QueryBuilder[T]) WithSelect(fields ...string) *QueryBuilder[T] {
	qb.selectFields = fields
	return qb
}

// ---------- Query Execution Methods ----------

func (qb *QueryBuilder[T]) FindAll() ([]T, error) {
	var results []T
	tx := qb.buildQuery()
	if err := tx.Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func (qb *QueryBuilder[T]) FindOne() (*T, error) {
	var model T
	tx := qb.buildQuery()
	if err := tx.First(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (qb *QueryBuilder[T]) FindAllPtr() ([]*T, error) {
	var result []*T
	query := qb.buildQuery()
	if err := query.Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (qb *QueryBuilder[T]) FindByID(id any) (*T, error) {
	var result T
	tx := qb.buildQuery()
	if err := tx.First(&result, id).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (qb *QueryBuilder[T]) FindFirst() (*T, error) {
	var result T
	tx := qb.buildQuery()
	if err := tx.First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func (qb *QueryBuilder[T]) Create(data *T) error {
	return qb.db.Create(data).Error
}

func (qb *QueryBuilder[T]) CreateMany(data []*T) error {
	return qb.db.Create(data).Error
}

func (qb *QueryBuilder[T]) UpdateByID(id int64, updates map[string]interface{}) (*T, error) {
	var model T

	findDB := qb.db
	if qb.unscoped {
		findDB = findDB.Unscoped()
	}

	if err := findDB.First(&model, id).Error; err != nil {
		return nil, err
	}
	if err := qb.updateAssociations(&model); err != nil {
		return nil, err
	}

	freshDB := qb.db.Session(&gorm.Session{NewDB: true})
	if qb.unscoped {
		freshDB = freshDB.Unscoped()
	}

	if err := freshDB.Model(&model).Updates(updates).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func (qb *QueryBuilder[T]) UpdateMany(ids []int64, updates map[string]interface{}) (model *T, err error) {
	db := qb.db
	if qb.unscoped {
		db = db.Unscoped()
	}
	if len(ids) > 0 {
		db = db.Where("id IN ?", ids)
	}

	db = qb.applyWhere(db)

	if err = db.Model(new(T)).Updates(updates).Error; err != nil {
		return nil, err
	}

	return model, err
}

func (qb *QueryBuilder[T]) Delete(id any) error {
	db := qb.db

	switch v := id.(type) {
	case []int64:
		if err := db.Where("id IN ?", v).Delete(new(T)).Error; err != nil {
			return err
		}
	default:
		var model T
		if err := db.First(&model, id).Error; err != nil {
			return err
		}
		if err := qb.deleteAssociations(&model); err != nil {
			return err
		}
		if err := db.Delete(&model).Error; err != nil {
			return err
		}
	}

	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(new(T)); err != nil {
		return err
	}

	tableName := stmt.Schema.Table
	rawSQL := fmt.Sprintf(`
		SELECT setval(
			COALESCE(pg_get_serial_sequence('%s','id')::regclass, ('%s_seq')::regclass),
			COALESCE(t.m, 1),
			(t.m IS NOT NULL)
		)
		FROM (SELECT MAX(id) AS m FROM %s) t;
	`, tableName, tableName, tableName)

	if err := db.Exec(rawSQL).Error; err != nil {
		return err
	}

	return nil
}

func (qb *QueryBuilder[T]) Count() (int64, error) {
	var count int64
	tx := qb.buildQuery()
	if err := tx.Model(new(T)).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// ---------- Internal Utilities ----------

func (qb *QueryBuilder[T]) buildQuery() *gorm.DB {
	tx := qb.db
	fmt.Println("qb cursor", qb.cursor)
	if qb.unscoped {
		tx = tx.Unscoped()
	}

	if len(qb.selectFields) > 0 {
		tx = tx.Select(qb.selectFields)
	}

	tx = qb.applyPreload(tx)
	for _, join := range qb.joins {
		tx = tx.Joins(join)
	}
	tx = qb.applyWhere(tx)

	if qb.orderBy != "" {
		tx = tx.Order(qb.orderBy)
	}
	if qb.limit != nil {
		tx = tx.Limit(*qb.limit)
	}
	if qb.cursor != nil {
		tx = tx.Where("id > ?", *qb.cursor)
	}

	return tx
}

func (qb *QueryBuilder[T]) applyPreload(tx *gorm.DB) *gorm.DB {
	for _, preload := range qb.preloads {
		tx = tx.Preload(preload)
	}
	return tx
}

func (qb *QueryBuilder[T]) applyWhere(tx *gorm.DB) *gorm.DB {
	for _, where := range qb.whereClauses {
		tx = where(tx)
	}
	return tx
}

func (qb *QueryBuilder[T]) updateAssociations(model any) error {
	for _, relation := range qb.associations {
		if data, ok := qb.replacements[relation]; ok {
			omit := fmt.Sprintf("%s.*", relation)
			if err := qb.db.Model(model).Omit(omit).Association(relation).Replace(data); err != nil {
				return err
			}
		}
	}
	return nil
}

func (qb *QueryBuilder[T]) deleteAssociations(model any) error {
	for _, relation := range qb.associations {
		if err := qb.db.Model(model).Association(relation).Clear(); err != nil {
			return fmt.Errorf("failed to clear association %s: %w", relation, err)
		}
	}
	return nil
}

func (qb *QueryBuilder[T]) Sum(column string, result any) error {
	tx := qb.buildQuery()
	return tx.Model(new(T)).
		Select(fmt.Sprintf("COALESCE(SUM(%s), 0)", column)).
		Scan(result).Error
}
