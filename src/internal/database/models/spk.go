package models

import (
	"time"

	"gorm.io/gorm"
)

type Spk struct {
	ID          int64   `gorm:"primaryKey;autoIncrement:false;type:bigint;default:nextval('spks_seq'::regclass)" json:"id"`
	Name        string  `gorm:"column:name;size:255;index:idx_spks_name" json:"name"`
	Code        string  `gorm:"column:code;size:255;unique:idx_spks_code" json:"code"`
	Description *string `gorm:"column:description;type:text" json:"description"`

	CreatedAt time.Time      `gorm:"autoCreateTime;index:idx_spks_created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_spks_deleted_at" json:"deleted_at"`

	HasTitles []Title  `gorm:"many2many:spk_titles;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"has_titles"`
	HasJobs   []SpkJob `gorm:"foreignKey:SpkID;references:ID;constraint:OnDelete:CASCADE;" json:"has_jobs"`
}

func (*Spk) TableName() string {
	return "spks"
}
