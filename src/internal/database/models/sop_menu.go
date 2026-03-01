package models

import "time"

type SopMenu struct {
	ID         int64     `gorm:"primaryKey;autoIncrement:false;type:bigint;default:nextval('sop_menus_seq'::regclass)" json:"id"`
	Name       string    `gorm:"size:255;not null;index:idx_sop_menus_name" json:"name"`
	SopID      *int64    `gorm:"index:idx_sop_menus_sop_id" json:"sop_id"`
	Type       string    `gorm:"size:50;index:idx_sop_menus_type" json:"type"`
	Multiple   bool      `gorm:"index:idx_sop_menus_multiple" json:"multiple"`
	DivisionID *int64    `gorm:"index:idx_sop_menus_division_id" json:"division_id"`
	ParentID   *int64    `gorm:"index:idx_sop_menus_parent_id" json:"parent_id"`
	MasterID   *int64    `gorm:"index:idx_sop_menus_master_id" json:"master_id"`
	IsMaster   bool      `gorm:"index:idx_sop_menus_is_master; default:false" json:"is_master"`
	CreatedAt  time.Time `gorm:"autoCreateTime;index:idx_sop_menus_created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	HasSop       *Sop      `gorm:"foreignKey:SopID;references:ID;constraint:OnDelete:CASCADE;" json:"has_sop,omitempty"`
	HasDivision  *Division `gorm:"foreignKey:DivisionID;references:ID;constraint:OnDelete:SET NULL;" json:"has_division,omitempty"`
	HasDocuments []SopMenu `gorm:"foreignKey:ParentID;references:ID;constraint:OnDelete:SET NULL;" json:"has_documents,omitempty"`
}

func (SopMenu) TableName() string {
	return "sop_menus"
}
