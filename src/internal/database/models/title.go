package models

import (
	"time"

	"gorm.io/gorm"
)

type Title struct {
	ID    int64  `gorm:"primaryKey;autoIncrement:false;type:bigint;default:nextval('titles_seq'::regclass)" json:"id"`
	Code  string `gorm:"size:255;unique:uni_title_code" json:"code"`
	Color string `gorm:"column:color;size:20" json:"color"`
	Name  string `gorm:"column:name;type:varchar(255);not null;index:idx_titles_name" json:"name"`

	PositionID *int64         `gorm:"column:position_id;index:idx_titles_position_id" json:"position_id"`
	LevelID    *int64         `gorm:"column:level_id;index:idx_titles_level_id" json:"level_id"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index:idx_titles_deleted_at" json:"deleted_at:omitempty"`

	// Relations
	HasPosition *Position `gorm:"foreignKey:PositionID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"position"`
	HasLevel    *Level    `gorm:"foreignKey:LevelID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"level"`
	HasSops     []Sop     `gorm:"many2many:sop_titles;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"sops"`
	HasSpks     []Spk     `gorm:"many2many:spk_titles;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"spks"`
}

func (Title) TableName() string {
	return "titles"
}
