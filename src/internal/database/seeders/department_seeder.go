package seeders

import (
	"jk-api/internal/database/models"

	"gorm.io/gorm"
)

func SeedDepartments(db *gorm.DB) error {
	var code1 string = "HR"
	var code2 string = "ENG"
	var code3 string = "FIN"
	var code4 string = "OPS"

	departments := []models.Department{
		{Name: "Human Resources", Code: &code1},
		{Name: "Engineering", Code: &code2},
		{Name: "Finance", Code: &code3},
		{Name: "Operations", Code: &code4},
	}
	return db.Create(&departments).Error
}
