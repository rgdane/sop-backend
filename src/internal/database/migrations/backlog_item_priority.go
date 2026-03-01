package migrations

import (
	"log"

	"gorm.io/gorm"
)

func AddBacklogPriorityCheck(db *gorm.DB) error {
	log.Println("🔄 Running Backlog Priority Migration...")

	var tableExists bool
	checkTableSQL := `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'backlog_items')`

	if err := db.Raw(checkTableSQL).Scan(&tableExists).Error; err != nil {
		log.Printf("⚠️ Could not check if table exists: %v", err)
		return err
	}

	if !tableExists {
		log.Println("⚠️ Table 'backlog_items' does not exist yet, skipping constraint migration")
		log.Println("✅ Backlog Priority Migration Completed (skipped - table not found)")
		return nil
	}

	var constraintExists bool
	checkConstraintSQL := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.check_constraints 
			WHERE constraint_name = 'chk_backlog_items_priority'
		)`

	if err := db.Raw(checkConstraintSQL).Scan(&constraintExists).Error; err != nil {
		log.Printf("⚠️ Could not check constraint existence: %v", err)
		constraintExists = false
	}

	if constraintExists {
		log.Println("✅ Priority constraint already exists, skipping")
		log.Println("✅ Backlog Priority Migration Completed")
		return nil
	}

	var columnExists bool
	checkColumnSQL := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_name = 'backlog_items' AND column_name = 'priority'
		)`

	if err := db.Raw(checkColumnSQL).Scan(&columnExists).Error; err != nil {
		log.Printf("⚠️ Could not check column existence: %v", err)
		return err
	}

	if !columnExists {
		log.Println("⚠️ Priority column does not exist yet, skipping constraint migration")
		log.Println("✅ Backlog Priority Migration Completed (skipped - column not found)")
		return nil
	}

	// Drop existing constraint if any (ignore error if doesn't exist)
	dropConstraintSQL := `ALTER TABLE backlog_items DROP CONSTRAINT IF EXISTS chk_backlog_items_priority`
	if err := db.Exec(dropConstraintSQL).Error; err != nil {
		log.Printf("⚠️ Could not drop existing constraint (might not exist): %v", err)
		// Continue anyway
	}

	// Add the constraint
	addConstraintSQL := `
		ALTER TABLE backlog_items
		ADD CONSTRAINT chk_backlog_items_priority
		CHECK (priority IN ('Low','Medium','High'))`

	if err := db.Exec(addConstraintSQL).Error; err != nil {
		log.Printf("❌ Failed to add priority constraint: %v", err)
		return err
	}

	log.Println("✅ Added priority check constraint to backlog_items")
	log.Println("✅ Backlog Priority Migration Completed")
	return nil
}
