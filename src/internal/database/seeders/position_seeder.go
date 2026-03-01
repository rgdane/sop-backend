package seeders

import (
	"jk-api/internal/database/models"

	"gorm.io/gorm"
)

func SeedPositions(db *gorm.DB) error {
	// ambil division Engineering misalnya
	var it models.Division
	if err := db.Where("name = ?", "IT").First(&it).Error; err != nil {
		return err
	}

	positions := []models.Position{
		{Name: "Backend Developer", DivisionID: it.ID},
		{Name: "Frontend Developer", DivisionID: it.ID},
	}
	return db.Create(&positions).Error
}
