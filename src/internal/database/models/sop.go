package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Sop struct {
	ID          int64          `gorm:"primaryKey;autoIncrement:false;type:bigint;default:nextval('sops_seq'::regclass)" json:"id"`
	Name        string         `gorm:"size:255;not null;unique:uni_sops_name" json:"name"`
	Code        string         `gorm:"size:255;index:idx_sops_code" json:"code"`
	Description *string        `gorm:"type:text;default:null" json:"description"`
	CreatedAt   time.Time      `gorm:"autoCreateTime;index:idx_sops_created_at" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index:idx_sop_deleted_at" json:"deleted_at"`
	ParentJobID *int64         `gorm:"column:parent_job_id;index:idx_sops_parent_job_id" json:"parent_job_id"`

	// Relations
	// HasParentJob *SopJob   `gorm:"foreignKey:ParentJobID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"parent_job"`
	HasTitles    []Title    `gorm:"many2many:sop_titles;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"has_titles"`
	HasJobs      []SopJob   `gorm:"foreignKey:SopID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"has_jobs"`
	HasDivisions []Division `gorm:"many2many:sop_divisions;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"has_divisions"`
}

func (Sop) TableName() string {
	return "sops"
}

func (l *Sop) AfterCreate(tx *gorm.DB) error {
	if l.Code == "" {
		var divisionID int64
		if len(l.HasDivisions) > 0 {
			divisionID = l.HasDivisions[0].ID
		}
		code := l.GenerateSopCode(tx, divisionID, l.ID)
		if code != "" {
			if err := tx.Model(l).Where("id = ?", l.ID).Update("code", code).Error; err != nil {
				return err
			}
			l.Code = code
		}
	}
	return nil
}

func (l *Sop) GenerateSopCode(tx *gorm.DB, divisionID int64, excludeID int64) string {
	if divisionID == 0 {
		return ""
	}

	// Ambil Division
	var division Division
	if err := tx.Model(&Division{}).Where("id = ?", divisionID).First(&division).Error; err != nil {
		return ""
	}

	// Ambil Department
	var department Department
	if err := tx.Model(&Department{}).Where("id = ?", division.DepartmentID).First(&department).Error; err != nil {
		return ""
	}

	if department.Code == nil || *department.Code == "" {
		return ""
	}
	if division.Code == "" {
		return ""
	}

	// Prefix: DEPT.DIV.
	prefix := fmt.Sprintf("%s.%s.", *department.Code, division.Code)

	// Ambil semua code yang match prefix
	query := tx.Model(&Sop{}).Select("code").Where("code LIKE ? AND deleted_at IS NULL", prefix+"%")
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}

	var codes []string
	if err := query.Order("code DESC").Scan(&codes).Error; err != nil && err != gorm.ErrRecordNotFound {
		return ""
	}

	// Cari sequence berikutnya di Go
	nextSeq := 1
	usedSequences := make(map[int]bool)

	for _, code := range codes {
		if strings.HasPrefix(code, prefix) {
			parts := strings.Split(code, ".")
			if len(parts) >= 3 {
				seqStr := strings.TrimLeft(parts[2], "0") // hapus leading zero
				if seqStr == "" {
					seqStr = "0"
				}
				if seq, err := strconv.Atoi(seqStr); err == nil {
					usedSequences[seq] = true
					if seq >= nextSeq {
						nextSeq = seq + 1
					}
				}
			}
		}
	}

	// Kalau ada sequence bolong, isi yang bolong dulu
	for i := 1; i < nextSeq; i++ {
		if !usedSequences[i] {
			nextSeq = i
			break
		}
	}

	// Return code dengan padding 4 digit
	return fmt.Sprintf("%s%04d", prefix, nextSeq)
}
