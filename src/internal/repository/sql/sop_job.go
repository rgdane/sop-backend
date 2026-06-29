package sql

import (
	"jk-api/internal/config"
	"jk-api/internal/database/models"
	"jk-api/internal/shared/helper"
	"jk-api/pkg/gorm/builder"
	"time"

	"gorm.io/gorm"
)

type SopJobRepository interface {
	WithTx(tx *gorm.DB) SopJobRepository
	WithSelect(fields ...string) SopJobRepository
	WithPreloads(preloads ...string) SopJobRepository
	WithAssociations(associations ...string) SopJobRepository
	WithReplacements(replacements map[string]any) SopJobRepository
	WithJoins(joins ...string) SopJobRepository
	WithWhere(query any, args ...any) SopJobRepository
	WithOrder(order string) SopJobRepository
	WithLimit(limit int) SopJobRepository
	WithOffset(offset int) SopJobRepository
	WithCursor(cursor int) SopJobRepository
	WithUnscoped() SopJobRepository

	InsertSopJob(data *models.SopJob) (*models.SopJob, error)
	InsertManySopJobs(data []*models.SopJob) ([]*models.SopJob, error)
	UpdateSopJob(id int64, updates map[string]any) (*models.SopJob, error)
	UpdateManySopJobs(ids []int64, updates map[string]any) error
	RemoveSopJob(id int64) error
	RemoveManySopJobs(ids []int64) error

	FindSopJob() ([]models.SopJob, error)
	FindSopJobWithJoins() ([]models.SopJob, error)
	FindSopJobByID(id int64) (*models.SopJob, error)
	FindSopJobByIDs(ids []int64) ([]*models.SopJob, error)
	ReorderSopJob(sopJobID int64, newIndex int, sopID int64) error
	CountSopJobs() (int64, error)
	FindSopJobsByTitleName(titleName string) ([]models.SopJob, error)
	FindSopJobsByDivisionName(divisionName string) ([]models.SopJob, error)
	FindSopJobsByDivisionAndTitle(divisionName, titleName string) ([]models.SopJob, error)
	FindSopJobsByReferenceDivisionName(divisionName string) ([]models.SopJob, error)
	FindSopJobsByDivisionTitlePublished(divisionName, jobNamePattern, spkName string) ([]models.SopJob, error)
}

type sopJobRepository struct {
	db           *gorm.DB
	fields       []string
	preloads     []string
	associations []string
	replacements map[string]interface{}
	joins        []string
	whereClauses []func(*gorm.DB) *gorm.DB
	order        string
	limit        *int
	offset       *int
	cursor       *int
	unscoped     bool
}

func NewSopJobRepository() SopJobRepository {
	return&sopJobRepository{db: config.DB}
}

func (repo *sopJobRepository) clone() *sopJobRepository {
	clone := *repo
	return &clone
}

func (repo *sopJobRepository) WithTx(tx *gorm.DB) SopJobRepository {
	clone := repo.clone()
	clone.db = tx
	return clone
}

func (repo *sopJobRepository) WithSelect(fields ...string) SopJobRepository {
	clone := repo.clone()
	clone.fields = append(clone.fields, fields...)
	return clone
}

func (repo *sopJobRepository) WithPreloads(preloads ...string) SopJobRepository {
	clone := repo.clone()
	clone.preloads = append(clone.preloads, preloads...)
	return clone
}

func (repo *sopJobRepository) WithAssociations(associations ...string) SopJobRepository {
	clone := repo.clone()
	clone.associations = append(clone.associations, associations...)
	return clone
}

func (repo *sopJobRepository) WithReplacements(replacements map[string]interface{}) SopJobRepository {
	clone := repo.clone()
	clone.replacements = replacements
	return clone
}

func (repo *sopJobRepository) WithJoins(joins ...string) SopJobRepository {
	clone := repo.clone()
	clone.joins = append(clone.joins, joins...)
	return clone
}

func (repo *sopJobRepository) WithWhere(query interface{}, args ...interface{}) SopJobRepository {
	clone := repo.clone()
	clone.whereClauses = append(clone.whereClauses, func(db *gorm.DB) *gorm.DB {
		return db.Where(query, args...)
	})
	return clone
}

func (repo *sopJobRepository) WithOrder(order string) SopJobRepository {
	clone := repo.clone()
	clone.order = order
	return clone
}

func (repo *sopJobRepository) WithLimit(limit int) SopJobRepository {
	clone := repo.clone()
	clone.limit = &limit
	return clone
}

func (repo *sopJobRepository) WithOffset(offset int) SopJobRepository {
	clone := repo.clone()
	clone.offset = &offset
	return clone
}

func (repo *sopJobRepository) WithCursor(cursor int) SopJobRepository {
	clone := repo.clone()
	clone.cursor = &cursor
	return clone
}

func (repo *sopJobRepository) WithUnscoped() SopJobRepository {
	clone := repo.clone()
	clone.unscoped = true
	return clone
}

func (repo *sopJobRepository) getQueryBuilder() *builder.QueryBuilder[models.SopJob] {
	db := repo.db
	if repo.unscoped {
		db = db.Unscoped()
	}

	qb := builder.NewQueryBuilder[models.SopJob](db).
		WithSelect(repo.fields...).
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
	if repo.offset != nil {
		qb = qb.WithOffset(*repo.offset)
	}
	if repo.cursor != nil {
		qb = qb.WithCursor(*repo.cursor)
	}
	return qb
}

func (repo *sopJobRepository) InsertSopJob(data *models.SopJob) (*models.SopJob, error) {
	if err := repo.getQueryBuilder().Create(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *sopJobRepository) InsertManySopJobs(data []*models.SopJob) ([]*models.SopJob, error) {
	if err := repo.getQueryBuilder().CreateMany(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *sopJobRepository) UpdateSopJob(id int64, updates map[string]interface{}) (*models.SopJob, error) {
	return repo.getQueryBuilder().UpdateByID(id, updates)
}

func (repo *sopJobRepository) UpdateManySopJobs(ids []int64, updates map[string]interface{}) error {
	_, err := repo.getQueryBuilder().UpdateMany(ids, updates)
	return err
}

func (repo *sopJobRepository) RemoveSopJob(id int64) error {
	return repo.getQueryBuilder().Delete(id)
}

func (repo *sopJobRepository) RemoveManySopJobs(ids []int64) error {
	var job models.SopJob
	return repo.db.Where("id IN ?", ids).Delete(&job).Error
}

func (repo *sopJobRepository) FindSopJob() ([]models.SopJob, error) {
	r := repo
	if r.order == "" {
		r = repo.clone()
		r.order = "index ASC"
	}
	return repo.getQueryBuilder().FindAll()
}

type SopJobJoinResult struct {
	ID          int64     `gorm:"column:id"`
	Name        string    `gorm:"column:name"`
	Alias       string    `gorm:"column:alias"`
	Type        *string   `gorm:"column:type"`
	Code        string    `gorm:"column:code"`
	Description *string   `gorm:"column:description"`
	TitleID     *int64    `gorm:"column:title_id"`
	SopID       int64     `gorm:"column:sop_id"`
	ReferenceID *int64    `gorm:"column:reference_id"`
	Index       int       `gorm:"column:index"`
	FlowchartID *int64    `gorm:"column:flowchart_id"`
	NextIndex   *int      `gorm:"column:next_index"`
	PrevIndex   *int      `gorm:"column:prev_index"`
	IsPublished *bool     `gorm:"column:is_published"`
	IsHide      *bool     `gorm:"column:is_hide"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`

	TitleIDMapped *int64 `gorm:"column:title_id_mapped"`
	TitleName     string `gorm:"column:title_name"`
	TitleCode     string `gorm:"column:title_code"`
	TitleColor    string `gorm:"column:title_color"`

	RefSopID   *int64 `gorm:"column:ref_sop_id"`
	RefSopName string `gorm:"column:ref_sop_name"`
	RefSopCode string `gorm:"column:ref_sop_code"`

	RefSpkID   *int64 `gorm:"column:ref_spk_id"`
	RefSpkName string `gorm:"column:ref_spk_name"`
	RefSpkCode string `gorm:"column:ref_spk_code"`
}

func (repo *sopJobRepository) FindSopJobWithJoins() ([]models.SopJob, error) {
	r := repo
	if r.order == "" {
		r = repo.clone()
		r.order = "index ASC"
	}

	db := repo.db.Table("sop_jobs")
	if repo.unscoped {
		db = db.Unscoped()
	}

	if len(repo.fields) > 0 {
		db = db.Select(repo.fields)
	} else {
		db = db.Select(`
			sop_jobs.id, sop_jobs.name, sop_jobs.alias, sop_jobs.type, sop_jobs.code,
			sop_jobs.description, sop_jobs.title_id, sop_jobs.sop_id, sop_jobs.reference_id,
			sop_jobs.index, sop_jobs.flowchart_id, sop_jobs.next_index, sop_jobs.prev_index,
			sop_jobs.is_published, sop_jobs.is_hide, sop_jobs.created_at, sop_jobs.updated_at,
			titles.id as title_id_mapped, titles.name as title_name, titles.code as title_code, titles.color as title_color,
			ref_sops.id as ref_sop_id, ref_sops.name as ref_sop_name, ref_sops.code as ref_sop_code,
			ref_spks.id as ref_spk_id, ref_spks.name as ref_spk_name, ref_spks.code as ref_spk_code
		`)
	}

	for _, join := range repo.joins {
		db = db.Joins(join)
	}

	for _, where := range repo.whereClauses {
		db = where(db)
	}

	if r.order != "" {
		db = db.Order(r.order)
	}
	if repo.limit != nil {
		db = db.Limit(*repo.limit)
	}
	if repo.offset != nil {
		db = db.Offset(*repo.offset)
	}
	if repo.cursor != nil {
		db = db.Where("id > ?", *repo.cursor)
	}

	var joinResults []SopJobJoinResult
	if err := db.Find(&joinResults).Error; err != nil {
		return nil, err
	}

	results := make([]models.SopJob, 0, len(joinResults))
	for _, jr := range joinResults {
		sopJob := models.SopJob{
			ID:          jr.ID,
			Name:        jr.Name,
			Alias:       jr.Alias,
			Type:        jr.Type,
			Code:        jr.Code,
			Description: jr.Description,
			TitleID:     jr.TitleID,
			SopID:       jr.SopID,
			ReferenceID: jr.ReferenceID,
			Index:       jr.Index,
			FlowchartID: jr.FlowchartID,
			NextIndex:   jr.NextIndex,
			PrevIndex:   jr.PrevIndex,
			IsPublished: jr.IsPublished,
			IsHide:      jr.IsHide,
			CreatedAt:   jr.CreatedAt,
			UpdatedAt:   jr.UpdatedAt,
		}

		if jr.TitleIDMapped != nil {
			sopJob.HasTitle = &models.Title{
				ID:    *jr.TitleIDMapped,
				Name:  jr.TitleName,
				Code:  jr.TitleCode,
				Color: jr.TitleColor,
			}
		}

		if jr.RefSopID != nil {
			sopJob.HasReference = &models.Sop{
				ID:   *jr.RefSopID,
				Name: jr.RefSopName,
				Code: jr.RefSopCode,
			}
		} else if jr.RefSpkID != nil {
			sopJob.HasReference = &models.Spk{
				ID:   *jr.RefSpkID,
				Name: jr.RefSpkName,
				Code: jr.RefSpkCode,
			}
		}

		results = append(results, sopJob)
	}

	return results, nil
}

func (repo *sopJobRepository) FindSopJobByID(id int64) (*models.SopJob, error) {
	return repo.getQueryBuilder().FindByID(id)
}

func (repo *sopJobRepository) FindSopJobByIDs(ids []int64) ([]*models.SopJob, error) {
	if len(ids) == 0 {
		return []*models.SopJob{}, nil
	}

	return repo.getQueryBuilder().WithWhere(func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN ?", ids)
	}).FindAllPtr()
}

func (repo *sopJobRepository) ReorderSopJob(sopJobID int64, newIndex int, sopID int64) error {
	var currentSopJob models.SopJob
	if err := repo.db.Where("id = ? AND sop_id = ?", sopJobID, sopID).First(&currentSopJob).Error; err != nil {
		return err
	}

	oldIndex := currentSopJob.Index

	if oldIndex == newIndex {
		return nil
	}

	var maxIndex int64
	if err := repo.db.Model(&models.SopJob{}).Where("sop_id = ?", sopID).Count(&maxIndex).Error; err != nil {
		return err
	}

	if newIndex > int(maxIndex) {
		newIndex = int(maxIndex)
	}
	if newIndex < 1 {
		newIndex = 1
	}

	if oldIndex < newIndex {
		if err := repo.db.Model(&models.SopJob{}).
			Where("sop_id = ? AND index > ? AND index <= ?", sopID, oldIndex, newIndex).
			UpdateColumn("index", gorm.Expr("index - 1")).Error; err != nil {
			return err
		}
	} else {
		if err := repo.db.Model(&models.SopJob{}).
			Where("sop_id = ? AND index >= ? AND index < ?", sopID, newIndex, oldIndex).
			UpdateColumn("index", gorm.Expr("index + 1")).Error; err != nil {
			return err
		}
	}

	if err := repo.db.Model(&models.SopJob{}).
		Where("id = ?", sopJobID).
		UpdateColumns(map[string]interface{}{
			"index":      newIndex,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return err
	}

	return nil
}

func (repo *sopJobRepository) CountSopJobs() (int64, error) {
	var count int64
	db := repo.db
	if repo.unscoped {
		db = db.Unscoped()
	}

	for _, where := range repo.whereClauses {
		db = where(db)
	}

	err := db.Model(&models.SopJob{}).Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (repo *sopJobRepository) FindSopJobsByTitleName(titleName string) ([]models.SopJob, error) {
	db := repo.db.Table("sop_jobs j").
		Select("j.id, j.name, j.type, j.code, j.index").
		Joins("JOIN titles t ON t.id = j.title_id").
		Where("t.name = ?", titleName).
		Limit(100).
		Order("j.index ASC")

	start := time.Now()
	var results []models.SopJob
	if err := db.Find(&results).Error; err != nil {
		return nil, err
	}
	helper.RecordDBLatency(time.Since(start))

	return results, nil
}

func (repo *sopJobRepository) FindSopJobsByDivisionName(divisionName string) ([]models.SopJob, error) {
	db := repo.db.Table("sop_jobs j").
		Select("j.id, j.name, j.type, j.code, j.index").
		Joins("JOIN sops s ON s.id = j.sop_id").
		Joins("JOIN sop_divisions sd ON sd.sop_id = s.id").
		Joins("JOIN divisions d ON d.id = sd.division_id").
		Where("d.name = ?", divisionName).
		Limit(100).
		Order("j.index ASC")

	start := time.Now()
	var results []models.SopJob
	if err := db.Find(&results).Error; err != nil {
		return nil, err
	}
	helper.RecordDBLatency(time.Since(start))

	return results, nil
}

func (repo *sopJobRepository) FindSopJobsByDivisionAndTitle(divisionName, titleName string) ([]models.SopJob, error) {
	db := repo.db.Table("sop_jobs j").
		Select("j.id, j.name, j.type, j.code, j.index").
		Joins("JOIN sops s ON s.id = j.sop_id").
		Joins("JOIN sop_divisions sd ON sd.sop_id = s.id").
		Joins("JOIN divisions d ON d.id = sd.division_id").
		Joins("JOIN titles t ON t.id = j.title_id").
		Where("d.name = ? AND t.name = ?", divisionName, titleName).
		Limit(100).
		Order("j.index ASC")

	start := time.Now()
	var results []models.SopJob
	if err := db.Find(&results).Error; err != nil {
		return nil, err
	}
	helper.RecordDBLatency(time.Since(start))

	return results, nil
}

func (repo *sopJobRepository) FindSopJobsByReferenceDivisionName(divisionName string) ([]models.SopJob, error) {
	db := repo.db.Table("sop_jobs j").
		Select("j.id, j.name, j.type, j.code, j.index").
		Joins("JOIN sops ref_sops ON ref_sops.id = j.reference_id AND j.type = 'sop'").
		Joins("JOIN sop_divisions sd ON sd.sop_id = ref_sops.id").
		Joins("JOIN divisions d ON d.id = sd.division_id").
		Where("d.name = ?", divisionName).
		Limit(100).
		Order("j.index ASC")

	start := time.Now()
	var results []models.SopJob
	if err := db.Find(&results).Error; err != nil {
		return nil, err
	}
	helper.RecordDBLatency(time.Since(start))

	return results, nil
}

func (repo *sopJobRepository) FindSopJobsByDivisionTitlePublished(divisionName, jobNamePattern, spkName string) ([]models.SopJob, error) {
	db := repo.db.Table("sop_jobs j").
		Select("j.id, j.name, j.type, j.code, j.index").
		Joins("JOIN sops s ON s.id = j.sop_id AND j.type = 'spk'").
		Joins("JOIN sop_divisions sd ON sd.sop_id = s.id").
		Joins("JOIN divisions d ON d.id = sd.division_id").
		Joins("JOIN spks spk ON spk.id = j.reference_id").
		Where("d.name = ?", divisionName).
		Where("j.name LIKE ?", "%"+jobNamePattern+"%").
		Where("j.is_published = ?", true).
		Where("spk.name LIKE ?", "%"+spkName+"%").
		Where("j.reference_id IS NOT NULL").
		Limit(100).
		Order("j.index ASC")

	start := time.Now()
	var results []models.SopJob
	if err := db.Find(&results).Error; err != nil {
		return nil, err
	}
	helper.RecordDBLatency(time.Since(start))

	return results, nil
}
