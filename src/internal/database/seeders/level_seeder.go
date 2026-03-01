package seeders

import (
	"jk-api/internal/database/models"

	"gorm.io/gorm"
)

func SeedLevels(db *gorm.DB) error {
	levels := []models.Level{
		{Name: "Junior"},
		{Name: "Mid"},
		{Name: "Senior"},
	}
	return db.Create(&levels).Error
}
