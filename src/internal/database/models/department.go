package models

import (
	"time"

	"gorm.io/gorm"
)

type Department struct {
	ID        int64          `gorm:"primaryKey;autoIncrement:false;type:bigint;default:nextval('departments_seq'::regclass)" json:"id"`
	Name      string         `gorm:"size:255;unique:uni_departments_name;not null" json:"name" validate:"required,min=3"`
	Code      *string        `gorm:"size:255;unique:uni_departments_code" json:"code"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index:idx_departments_deleted_at" json:"deleted_at"`
}

func (*Department) TableName() string {
	return "departments"
}
