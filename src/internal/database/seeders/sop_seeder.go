package seeders

import (
	"jk-api/internal/database/models"

	"gorm.io/gorm"
)

func SeedSops(db *gorm.DB) error {
	// contoh SOPs
	sops := []models.Sop{
		{
			Name: "Code Review Process",
			Code: "SOP-CR",
		},
		{
			Name: "Deployment Process",
			Code: "SOP-DEP",
		},
	}

	if err := db.Create(&sops).Error; err != nil {
		return err
	}

	// ambil Title untuk asosiasi
	var backendJunior models.Title
	if err := db.Where("code = ?", "BE-JR").First(&backendJunior).Error; err != nil {
		return err
	}

	// hubungkan SOP pertama ke Title "Junior Backend Developer"
	if err := db.Model(&sops[0]).Association("HasTitles").Append(&backendJunior); err != nil {
		return err
	}

	return nil
}
