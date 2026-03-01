package models

import (
	"time"

	"fmt"

	"gorm.io/gorm"
)

type Level struct {
	ID        int64          `gorm:"primaryKey;autoIncrement:false;type:bigint;default:nextval('levels_seq'::regclass)" json:"id"`
	Code      *string        `gorm:"size:255;unique:uni_levels_code" json:"code"`
	Name      string         `gorm:"column:name;type:varchar(255);unique:uni_levels_name;not null" json:"name"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index:idx_levels_deleted_at" json:"deleted_at"`

	Titles []Title `gorm:"foreignKey:LevelID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"titles"`
}

func (*Level) TableName() string {
	return "levels"
}

func (l *Level) BeforeCreate(tx *gorm.DB) (err error) {
	if l.ID == 0 {
		if err := tx.Raw("SELECT nextval('levels_seq')").Scan(&l.ID).Error; err != nil {
			return err
		}
	}

	if l.Code == nil || *l.Code == "" {
		code := fmt.Sprintf("%d", l.ID)
		l.Code = &code
	}

	return nil
}
