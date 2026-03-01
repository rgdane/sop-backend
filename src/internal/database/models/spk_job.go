package models

import (
	"time"

	"gorm.io/gorm"
)

type SpkJob struct {
	ID          int64   `gorm:"primaryKey;autoIncrement:false;type:bigint;default:nextval('spk_jobs_seq'::regclass)" json:"id"`
	Name        string  `gorm:"size:255;not null" json:"name" validate:"required"`
	Description *string `gorm:"type:text" json:"description"`

	SpkID   int64  `gorm:"not null;index:idx_spk_jobs_spk_id" json:"spk_id"`
	SopID   *int64 `gorm:"index:idx_spk_jobs_sop_id" json:"sop_id"`
	TitleID *int64 `gorm:"index:idx_spk_jobs_title_id" json:"title_id"`
	Index   int    `gorm:"default:0" json:"index"`

	// Flowchart
	FlowchartID *int64 `gorm:"default:1;index:idx_spk_jobs_flowchart_id" json:"flowchart_id"`
	NextIndex   *int   `gorm:"default:null" json:"next_index"`
	PrevIndex   *int   `gorm:"default:null" json:"prev_index"`

	CreatedAt time.Time `gorm:"autoCreateTime;index:idx_spk_jobs_created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relasi
	HasSop       *Sop       `gorm:"foreignKey:SopID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"has_sop"`
	HasTitle     *Title     `gorm:"foreignKey:TitleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"has_title"`
	HasFlowchart *Flowchart `gorm:"foreignKey:FlowchartID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"has_flowchart"`
	HasSpk       *Spk       `gorm:"foreignKey:SpkID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"has_spk"`
}

func (s *SpkJob) AfterCreate(tx *gorm.DB) error {
	if s.Index == 0 {
		var maxIndex int
		if err := tx.Model(&SpkJob{}).
			Where("spk_id = ?", s.SpkID).
			Select("COALESCE(MAX(index), 0)").
			Scan(&maxIndex).Error; err != nil {
			return err
		}
		s.Index = maxIndex + 1
		if err := tx.Save(s).Error; err != nil {
			return err
		}
	}
	return nil
}
