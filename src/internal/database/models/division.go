package models

import (
	"time"

	"gorm.io/gorm"
)

type Division struct {
	ID           int64          `gorm:"primaryKey;autoIncrement:false;type:bigint;default:nextval('divisions_seq'::regclass)" json:"id"`
	Name         string         `gorm:"column:name;size:255;not null;unique:uni_divisions_name" json:"name"`
	Code         string         `gorm:"column:code;size:50;unique:uni_divisions_code;index:idx_divisions_code" json:"code"`
	DepartmentID int64          `gorm:"column:department_id;not null;index:idx_divisions_department_id" json:"department_id"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime;index:idx_divisions_created_at" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index:idx_divisions_deleted_at" json:"deleted_at"`

	HasDepartment *Department `gorm:"foreignKey:DepartmentID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"has_department"`
	Positions     []Position  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"positions"`
	Sops          []Sop       `gorm:"many2many:sop_divisions;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"sops"`
}

func (Division) TableName() string {
	return "divisions"
}
