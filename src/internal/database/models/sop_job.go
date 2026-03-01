package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type SopJob struct {
	ID          int64   `gorm:"primaryKey;autoIncrement:false;type:bigint;default:nextval('sop_jobs_seq'::regclass)" json:"id"`
	Name        string  `gorm:"size:255;not null;index:idx_sop_jobs_name" json:"name"`
	Alias       string  `gorm:"size:255;index:idx_sop_jobs_alias" json:"alias"`
	Type        *string `gorm:"type:text;check:type IN ('sop','spk','instruction');index:idx_sop_jobs_type" json:"type"`
	Code        string  `gorm:"size:255" json:"code"`
	Description *string `gorm:"type:text" json:"description"`
	TitleID     *int64  `gorm:"index:idx_sop_jobs_title_id" json:"title_id"`
	SopID       int64   `gorm:"not null;index:idx_sop_jobs_sop_id" json:"sop_id"`
	ReferenceID *int64  `gorm:"default:null" json:"reference_id"`
	Index       int     `gorm:"default:0;index:idx_sop_jobs_sopid_index,priority:2" json:"index"`
	IsPublished *bool   `gorm:"default:false;index:idx_sop_jobs_is_published" json:"is_published"`
	IsHide      *bool   `gorm:"default:false;index:idx_sop_jobs_is_hide" json:"is_hide"`

	FlowchartID *int64 `gorm:"default:1;index:idx_sop_jobs_flowchart_id" json:"flowchart_id"`
	NextIndex   *int   `gorm:"default:null" json:"next_index"`
	PrevIndex   *int   `gorm:"default:null" json:"prev_index"`

	CreatedAt time.Time `gorm:"autoCreateTime;index:idx_sop_jobs_created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	HasSop       *Sop       `gorm:"foreignKey:SopID;references:ID;constraint:OnDelete:SET NULL;" json:"has_sop"`
	HasTitle     *Title     `gorm:"foreignKey:TitleID;references:ID;constraint:OnDelete:SET NULL;" json:"has_title"`
	HasFlowchart *Flowchart `gorm:"foreignKey:FlowchartID;references:ID;constraint:OnDelete:SET NULL;" json:"has_flowchart"`

	HasReference interface{} `gorm:"-" json:"has_reference,omitempty"`
}

func (SopJob) TableName() string {
	return "sop_jobs"
}

func (s *SopJob) BeforeCreate(tx *gorm.DB) error {
	if s.Code != "" {
		return nil
	}

	// pake sequence id sementara (currval/nextval)
	var id int64
	if err := tx.Raw("SELECT nextval('sop_jobs_seq')").Scan(&id).Error; err != nil {
		return err
	}
	s.ID = id
	s.Code = fmt.Sprintf("P%04d", s.ID)

	return nil
}

func (s *SopJob) AfterCreate(tx *gorm.DB) error {
	if s.Index == 0 {
		var maxIndex int
		if err := tx.Model(&SopJob{}).
			Where("sop_id = ?", s.SopID).
			Select("COALESCE(MAX(index), 0)").
			Scan(&maxIndex).Error; err != nil {
			return err
		}
		s.Index = maxIndex + 1
		if err := tx.Save(s).Error; err != nil {
			return err
		}
	}

	// Only set parent_job_id on Sop when the SopJob is of type 'sop'
	if s.Type == nil || *s.Type != "sop" {
		return nil
	}
	if err := tx.Model(&Sop{}).
		Where("id = ? AND parent_job_id IS NULL", s.ReferenceID).
		Update("parent_job_id", s.ID).Error; err != nil {
		return err
	}

	return nil
}
