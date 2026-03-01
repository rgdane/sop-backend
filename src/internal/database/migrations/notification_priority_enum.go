package migrations

import (
	"log"

	"gorm.io/gorm"
)

func NotificationPriorityEnum(db *gorm.DB) error {
	log.Println("🔄 Running Notification Migration...")

	// Alternative approach: Use VARCHAR with CHECK constraint
	// This is more compatible with CockroachDB

	// Check if table exists
	var tableExists bool
	checkTableSQL := `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'notifications')`
	if err := db.Raw(checkTableSQL).Scan(&tableExists).Error; err != nil {
		log.Printf("⚠️ Could not check if table exists: %v", err)
		tableExists = false
	}

	if tableExists {
		var constraintExists bool
		checkConstraintSQL := `SELECT EXISTS (SELECT 1 FROM information_schema.check_constraints WHERE constraint_name = 'notification_priority_check')`
		if err := db.Raw(checkConstraintSQL).Scan(&constraintExists).Error; err != nil {
			log.Printf("⚠️ Could not check constraint: %v", err)
			constraintExists = false
		}

		if !constraintExists {
			addConstraintSQL := `ALTER TABLE notifications ADD CONSTRAINT notification_priority_check CHECK (priority IN ('info', 'warning', 'important', 'urgent'))`

			if err := db.Exec(addConstraintSQL).Error; err != nil {
				log.Printf("❌ Failed to add constraint: %v", err)
				return err
			}

			log.Println("✅ Added priority check constraint")
		} else {
			log.Println("✅ Priority constraint already exists")
		}
	}

	log.Println("✅ Notification Migration Completed.")
	return nil
}
