package seeders

import (
	"jk-api/internal/database/models"

	"gorm.io/gorm"
)

func SeedDatabaseNodes(db *gorm.DB) error {
	databaseNodes := []models.DatabaseNode{
		{Name: "Users", TableRef: "users"},
		{Name: "Typography Categories", TableRef: "typography_categories"},
	}
	return db.Create(&databaseNodes).Error
}
