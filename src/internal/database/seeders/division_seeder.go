package seeders

import (
	"jk-api/internal/database/models"

	"gorm.io/gorm"
)

func SeedDivisions(db *gorm.DB) error {
	// ambil division Engineering misalnya

	divisions := []models.Division{
		{Name: "IT", Code: "IT-001"},
		{Name: "Design", Code: "DS-001"},
	}
	return db.Create(&divisions).Error
}
