package seeders

import (
	"jk-api/internal/database/models"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedRoles(db *gorm.DB) error {
	var superRole models.Role
	err := db.Where("name = ?", "super").First(&superRole).Error

	if err == gorm.ErrRecordNotFound {
		superRole = models.Role{
			Name:      "super",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := db.Create(&superRole).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	var permissions []models.Permission
	if err := db.Find(&permissions).Error; err != nil {
		return err
	}

	if err := db.Model(&superRole).Association("HasPermissions").Replace(&permissions); err != nil {
		return err
	}

	return nil
}

func SeedAdmin(db *gorm.DB) error {
	var superRole models.Role
	if err := db.Where("name = ?", "super").First(&superRole).Error; err != nil {
		return err
	}

	var existing models.User
	if err := db.Where("email = ?", "admin@mail.com").First(&existing).Error; err == nil {
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := models.User{
		Name:      "Admin",
		Email:     "admin@mail.com",
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&admin).Error; err != nil {
		return err
	}

	if err := db.Model(&admin).Association("HasRoles").Append(&superRole); err != nil {
		return err
	}

	return nil
}
