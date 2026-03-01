package migrations

import (
	"log"

	"gorm.io/gorm"
)

func SetupSequenceTable(db *gorm.DB) error {
	sequences := map[string]string{
		"users_seq":                      "users",
		"squads_seq":                     "squads",
		"roles_seq":                      "roles",
		"permissions_seq":                "permissions",
		"projects_seq":                   "projects",
		"departments_seq":                "departments",
		"divisions_seq":                  "divisions",
		"levels_seq":                     "levels",
		"positions_seq":                  "positions",
		"titles_seq":                     "titles",
		"statuses_seq":                   "statuses",
		"holidays_seq":                   "holidays",
		"leaves_seq":                     "leaves",
		"backlog_items_seq":              "backlog_items",
		"sprints_seq":                    "sprints",
		"backlogs_seq":                   "backlogs",
		"comments_seq":                   "comments",
		"activity_logs_seq":              "activity_logs",
		"notifications_seq":              "notifications",
		"cms_articles_seq":               "cms_articles",
		"cms_categories_seq":             "cms_categories",
		"cms_tags_seq":                   "cms_tags",
		"documents_seq":                  "documents",
		"sops_seq":                       "sops",
		"spks_seq":                       "spks",
		"spk_jobs_seq":                   "spk_jobs",
		"spk_txs_seq":                    "spk_txs",
		"spk_tx_results_seq":             "spk_tx_results",
		"spk_tx_versions_seq":            "spk_tx_versions",
		"sop_jobs_seq":                   "sop_jobs",
		"products_seq":                   "products",
		"features_seq":                   "features",
		"jobs_seq":                       "jobs",
		"flowcharts_seq":                 "flowcharts",
		"sprint_goals_seq":               "sprint_goals",
		"sprint_retros_seq":              "sprint_retros",
		"retro_items_seq":                "retro_items",
		"sprint_dailies_seq":             "sprint_dailies",
		"languages_seq":                  "languages",
		"color_palettes_seq":             "color_palettes",
		"colors_seq":                     "colors",
		"sop_menus_seq":                  "sop_menus",
		"typography_categories_sequence": "typography_categories",
		"typography_weights_sequence":    "typography_weights",
		"typography_scales_sequence":     "typography_scales",
		"typography_sequence":            "typography",
		"grids_seq":                      "grids",
		"grid_types_seq":                 "grid_types",
		"database_node_seq":              "database_nodes",
		"case_bug_features_seq":          "case_bug_features",
		"case_accounts_seq":              "case_accounts",
		"case_account_outlets_seq":       "case_account_outlets",
		"case_reports_seq":               "case_reports",
		"case_attachments_seq":           "case_attachments",
		"case_statuses_seq":              "case_statuses",
		"case_categories_seq":            "case_categories",
		"shifts_seq":                     "shifts",
		"technical_supports_seq":         "technical_supports",
		"customer_services_seq":          "customer_services",
		"shift_schedules_seq":            "shift_schedules",
		"shift_logs_seq":                 "shift_logs",
	}

	for seqName, tableName := range sequences {
		var existingSeq string

		err := db.Raw(`
			SELECT sequence_name 
			FROM information_schema.sequences 
			WHERE sequence_name = ? AND sequence_schema = 'public'
		`, seqName).Scan(&existingSeq).Error

		if err != nil {
			log.Printf("Error checking sequence %s: %v", seqName, err)
			continue
		}

		if existingSeq == "" {
			createSeqSQL := `
				CREATE SEQUENCE IF NOT EXISTS ` + seqName + `
					START 1
					INCREMENT 1
					MINVALUE 1
					MAXVALUE 9007199254740991;
			`

			if err := db.Exec(createSeqSQL).Error; err != nil {
				log.Printf("Failed to create sequence %s for table %s: %v", seqName, tableName, err)
				continue
			}
		} else {
			log.Printf("📝 Sequence already exists: %s", seqName)
		}
	}

	return nil
}
