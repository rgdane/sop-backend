package seeders

import (
	"fmt"
	"jk-api/internal/database/models"

	"gorm.io/gorm"
)

func SeedFlowcharts(db *gorm.DB) error {
	flowcharts := []models.Flowchart{
		{Type: "process"},
		{Type: "decision"},
	}

	for _, f := range flowcharts {
		var existing models.Flowchart
		if err := db.Where("type = ?", f.Type).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&f).Error; err != nil {
					return fmt.Errorf("failed to seed flowchart %s: %w", f.Type, err)
				}
			} else {
				return err
			}
		}
	}

	return nil
}
