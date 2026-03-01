package seeders

import (
	"jk-api/internal/database/models"

	"gorm.io/gorm"
)

func SeedDivisions(db *gorm.DB) error {
	// ambil division Engineering misalnya
	var operations models.Department
	if err := db.Where("name = ?", "Operations").First(&operations).Error; err != nil {
		return err
	}

	divisions := []models.Division{
		{Name: "IT", DepartmentID: operations.ID, Code: "IT-001"},
		{Name: "Design", DepartmentID: operations.ID, Code: "DS-001"},
	}
	return db.Create(&divisions).Error
}
