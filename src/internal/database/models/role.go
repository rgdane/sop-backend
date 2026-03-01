package models

import "time"

type Role struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false;type:bigint;default:nextval('roles_seq'::regclass)" json:"id"`
	Name      string    `gorm:"size:100;not null;unique:idx_roles_name" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime;index:idx_roles_created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	HasPermissions []Permission `gorm:"many2many:role_has_permissions;" json:"has_permission"`
	HasUsers       []*User      `gorm:"many2many:user_has_roles;joinForeignKey:RoleID;joinReferences:UserID;constraint:OnDelete:CASCADE;" json:"has_user"`
}

func (*Role) TableName() string {
	return "roles"
}
