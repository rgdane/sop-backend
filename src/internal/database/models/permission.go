package models

import "time"

type Permission struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false;type:bigint;default:nextval('permissions_seq'::regclass)" json:"id"`
	Name      string    `gorm:"size:255;not null;unique:idx_permissions_name" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime;index:idx_permissions_created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	HasRole []Role `gorm:"many2many:role_has_permissions;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;" json:"has_role"`
}

func (*Permission) TableName() string {
	return "permissions"
}
