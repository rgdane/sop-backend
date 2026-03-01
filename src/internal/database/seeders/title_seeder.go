package seeders

import (
	"jk-api/internal/database/models"

	"gorm.io/gorm"
)

func SeedTitles(db *gorm.DB) error {
	// ambil Level dan Position yang sudah ada
	var juniorLevel models.Level
	if err := db.Where("name = ?", "Junior").First(&juniorLevel).Error; err != nil {
		return err
	}

	var backendDev models.Position
	if err := db.Where("name = ?", "Backend Developer").First(&backendDev).Error; err != nil {
		return err
	}

	var frontendDev models.Position
	if err := db.Where("name = ?", "Frontend Developer").First(&frontendDev).Error; err != nil {
		return err
	}

	// seed Titles
	titles := []models.Title{
		{
			Code:       "BE-JR",
			Name:       "Junior Backend Developer",
			PositionID: &backendDev.ID,
			LevelID:    &juniorLevel.ID,
		},
		{
			Code:       "FE-JR",
			Name:       "Junior Frontend Developer",
			PositionID: &frontendDev.ID,
			LevelID:    &juniorLevel.ID,
		},
	}

	return db.Create(&titles).Error
}
