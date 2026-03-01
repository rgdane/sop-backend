package migrations

import (
	"log"

	"jk-api/internal/config"
	"jk-api/internal/database/models"
)

func Migrate() {
	db := config.DB

	if err := SetupSequenceTable(db); err != nil {
		log.Fatalf("Failed to create sequences: %v", err)
	}

	if err := SetupEnumUp(db); err != nil {
		log.Fatalf("Failed to setup enums: %v", err)
	}

	if err := DropShiftViews(db); err != nil {
		log.Printf("Warning: Failed to drop shift views: %v", err)
	}

	err := db.AutoMigrate(
		&models.User{},
		&models.Department{},
		&models.Division{},
		&models.Level{},
		&models.Position{},
		&models.Title{},
		&models.Role{},
		&models.Permission{},
		&models.Sop{},
		&models.Spk{},
		&models.Flowchart{},
		&models.SpkJob{},
		&models.SopJob{},
		&models.DatabaseNode{},
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration complete")
}
