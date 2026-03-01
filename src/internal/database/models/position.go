package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Position struct {
	ID         int64          `gorm:"primaryKey;autoIncrement:false;type:bigint;default:nextval('positions_seq'::regclass)" json:"id"`
	Name       string         `gorm:"column:name;size:255;not null;unique:uni_positions_name" json:"name"`
	Code       string         `gorm:"column:code;size:50;unique:uni_positions_code;index:idx_positions_code" json:"code"`
	Color      string         `gorm:"column:color;size:20" json:"color"`
	DivisionID int64          `gorm:"column:division_id;index:idx_positions_division_id" json:"division_id"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime;index:idx_positions_created_at" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index:idx_positions_deleted_at" json:"deleted_at"`

	HasDivision *Division `gorm:"foreignKey:DivisionID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"has_division"`
	Titles      []Title   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"titles"`
}

func (Position) TableName() string {
	return "positions"
}

func (p *Position) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == 0 {
		if err := tx.Raw("SELECT nextval('positions_seq')").Scan(&p.ID).Error; err != nil {
			return err
		}
	}

	if p.Code == "" {
		p.Code = fmt.Sprint(p.ID)
	}

	return nil
}
