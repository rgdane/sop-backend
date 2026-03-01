package migrations

import (
	"log"

	"gorm.io/gorm"
)

func SetupShiftTables(db *gorm.DB) error {
	log.Println("Setting up shift-related tables...")

	if err := createShiftViews(db); err != nil {
		log.Printf("Warning creating views: %v", err)
	}

	log.Println("Shift tables setup complete")
	return nil
}

func createShiftViews(db *gorm.DB) error {
	views := []struct {
		name string
		sql  string
	}{
		{
			name: "v_shift_schedules_with_details",
			sql: `
				CREATE OR REPLACE VIEW v_shift_schedules_with_details AS
				SELECT 
					ls.id,
					ls.reference,
					ls.user_id,
					ls.shift_id,
					ls.date,
					ls.created_at,
					ls.updated_at,
					u.name as user_name,
					s.name as shift_name,
					s.time as shift_time,
					ts.id as technical_support_id,
					CASE 
						WHEN ls.reference = 'technical' THEN 'Technical Support'
						WHEN ls.reference = 'customer_service' THEN 'Customer Service'
						ELSE 'Unknown'
					END as reference_display
				FROM shift_schedules ls
				INNER JOIN users u ON ls.user_id = u.id
				INNER JOIN shifts s ON ls.shift_id = s.id
				LEFT JOIN technical_supports ts 
					ON ls.user_id = ts.user_id 
					AND ls.reference = 'technical'
			`,
		},
		{
			name: "v_technical_support_shifts",
			sql: `
				CREATE OR REPLACE VIEW v_technical_support_shifts AS
				SELECT 
					ts.id as technical_support_id,
					ts.user_id,
					u.name as user_name,
					COUNT(ls.id) as total_shifts,
					COUNT(DISTINCT ls.shift_id) as unique_shifts,
					COUNT(DISTINCT ls.date) as days_worked,
					MIN(ls.date) as first_shift_date,
					MAX(ls.date) as last_shift_date
				FROM technical_supports ts
				INNER JOIN users u ON ts.user_id = u.id
				LEFT JOIN shift_schedules ls 
					ON ts.user_id = ls.user_id 
					AND ls.reference = 'technical'
				GROUP BY ts.id, ts.user_id, u.name
			`,
		},
		{
			name: "v_shift_statistics",
			sql: `
				CREATE OR REPLACE VIEW v_shift_statistics AS
				SELECT 
					s.id as shift_id,
					s.name as shift_name,
					s.time as shift_time,
					COUNT(ls.id) as total_logs,
					COUNT(DISTINCT ls.user_id) as unique_users,
					COUNT(DISTINCT CASE WHEN ls.reference = 'technical' THEN ls.id END) as technical_count,
					COUNT(DISTINCT CASE WHEN ls.reference = 'customer_service' THEN ls.id END) as customer_service_count
				FROM shifts s
				LEFT JOIN shift_schedules ls ON s.id = ls.shift_id
				GROUP BY s.id, s.name, s.time
			`,
		},
	}

	for _, view := range views {
		if err := db.Exec(view.sql).Error; err != nil {
			log.Printf("⚠️  Could not create view %s: %v", view.name, err)
			continue
		}
		log.Printf("✅ Created view: %s", view.name)
	}

	return nil
}

func DropShiftViews(db *gorm.DB) error {
	views := []string{
		"v_shift_schedules_with_details",
		"v_technical_support_shifts",
		"v_shift_statistics",
	}

	for _, view := range views {
		sql := "DROP VIEW IF EXISTS " + view + " CASCADE"
		if err := db.Exec(sql).Error; err != nil {
			log.Printf("⚠️  Could not drop view %s: %v", view, err)
			continue
		}
		log.Printf("🗑️  Dropped view: %s", view)
	}

	return nil
}

func AddShiftTableComments(db *gorm.DB) error {
	comments := []string{
		"COMMENT ON TABLE shifts IS 'Master data untuk shift kerja (pagi, siang, malam, dll)'",
		"COMMENT ON TABLE technical_supports IS 'Daftar user yang berperan sebagai Technical Support'",
		"COMMENT ON TABLE shift_schedules IS 'Jadwal shift harian untuk TS dan CS. Reference: ts=Technical Support, cs=Customer Service'",

		"COMMENT ON COLUMN shift_schedules.reference IS 'Tipe referensi: ts (Technical Support) atau cs (Customer Service)'",
		"COMMENT ON COLUMN shift_schedules.user_id IS 'User yang mengambil shift ini'",
		"COMMENT ON COLUMN shift_schedules.shift_id IS 'Shift yang diambil'",
		"COMMENT ON COLUMN shift_schedules.date IS 'Tanggal shift dilaksanakan'",

		"COMMENT ON COLUMN technical_supports.user_id IS 'User ID yang terdaftar sebagai Technical Support (UNIQUE)'",
	}

	for _, comment := range comments {
		if err := db.Exec(comment).Error; err != nil {
			log.Printf("⚠️  Could not add comment: %v", err)
			continue
		}
	}

	log.Println("✅ Added table comments")
	return nil
}
