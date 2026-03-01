package migrations

import (
	"log"

	"gorm.io/gorm"
)

func SetupShiftScheduleReferenceEnum(db *gorm.DB) error {
	var exists bool
	err := db.Raw(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'shift_schedules'
		)
	`).Scan(&exists).Error

	if err != nil {
		log.Printf("Error checking if shift_schedules table exists: %v", err)
		return err
	}

	if !exists {
		log.Println("Skipping enum constraint: shift_schedules table doesn't exist yet")
		return nil
	}

	var constraintExists string
	err = db.Raw(`
		SELECT constraint_name 
		FROM information_schema.table_constraints 
		WHERE table_name = 'shift_schedules' 
		AND constraint_name = 'shift_schedules_reference_check'
		AND constraint_type = 'CHECK'
	`).Scan(&constraintExists).Error

	if err != nil {
		log.Printf("Error checking constraint: %v", err)
		return err
	}

	if constraintExists != "" {
		log.Println("CHECK constraint 'shift_schedules_reference_check' already exists")
		return nil
	}

	sql := `
		ALTER TABLE shift_schedules 
		ADD CONSTRAINT shift_schedules_reference_check 
		CHECK (reference IN ('technical', 'customer_service'))
	`

	if err := db.Exec(sql).Error; err != nil {
		log.Printf("Failed to create CHECK constraint for shift_schedules.reference: %v", err)
		return err
	}

	log.Println("Created CHECK constraint for shift_schedules.reference (technical, customer_service)")
	return nil
}
