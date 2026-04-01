package seeders

import (
	"jk-api/internal/database/models"

	"gorm.io/gorm"
)

func SeedTitles(db *gorm.DB) error {

	// seed Titles
	titles := []models.Title{
		{
			Code:       "BE-JR",
			Name:       "Junior Backend Developer",
		},
		{
			Code:       "FE-JR",
			Name:       "Junior Frontend Developer",
		},
	}

	return db.Create(&titles).Error
}
