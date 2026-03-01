package models

import (
	"fmt"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const TableNameUser = "users"

// User mapped from table <users>
type User struct {
	ID                int64          `gorm:"primaryKey;autoIncrement:false;type:bigint;default:nextval('users_seq'::regclass)" json:"id"`
	TitleID           *int64         `gorm:"column:title_id;index:idx_users_title_id" json:"title_id"`
	Code              *string        `gorm:"column:code;size:50;unique;index:idx_users_code" json:"code"`
	Name              string         `gorm:"column:name;size:255;not null;index:idx_users_name" json:"name"`
	Email             string         `gorm:"column:email;size:255;not null;unique;index:idx_users_email" json:"email"`
	EmailVerifiedAt   *time.Time     `gorm:"column:email_verified_at" json:"email_verified_at"`
	Password          string         `gorm:"column:password;not null" json:"password"`
	RememberToken     *string        `gorm:"column:remember_token;size:100" json:"remember_token"`
	CustomFields      datatypes.JSON `gorm:"column:custom_fields;type:jsonb" json:"custom_fields"`
	AvatarUrl         *string        `gorm:"column:avatar_url;size:255" json:"avatar_url"`
	IsPasswordDefault bool           `gorm:"column:is_password_default;default:true;" json:"is_password_default"`
	CreatedAt         time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index:idx_users_deleted_at" json:"deleted_at"`

	// Relations
	HasTitle               Title         `gorm:"foreignKey:TitleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"has_title"`
	HasRoles               []Role        `gorm:"many2many:user_has_roles;constraint:OnDelete:CASCADE;" json:"has_roles"`
	HasDivisions           []Division    `gorm:"many2many:user_has_divisions;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"has_divisions"`
}

func (*User) TableName() string {
	return TableNameUser
}

func (u *User) GenerateUserCode() string {
	prefix := "KRY"
	code := fmt.Sprintf("%s%04d", prefix, u.ID)
	u.Code = &code
	return code
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	prefix := "KRY"

	if u.ID == 0 {
		var nextID int64
		if err := tx.Raw(`
			SELECT COALESCE(MAX(id), 0) + 1 FROM users
		`).Scan(&nextID).Error; err != nil {
			return fmt.Errorf("gagal mengambil next user id: %w", err)
		}
		u.ID = nextID
	}

	code := fmt.Sprintf("%s%04d", prefix, u.ID)
	u.Code = &code

	return nil
}
