package migrations

import (
	"log"

	"gorm.io/gorm"
)

func SetupSequenceTable(db *gorm.DB) error {
	sequences := map[string]string{
		"users_seq":                      "users",
		"roles_seq":                      "roles",
		"permissions_seq":                "permissions",
		"departments_seq":                "departments",
		"divisions_seq":                  "divisions",
		"levels_seq":                     "levels",
		"positions_seq":                  "positions",
		"titles_seq":                     "titles",
		"sops_seq":                       "sops",
		"spks_seq":                       "spks",
		"spk_jobs_seq":                   "spk_jobs",
		"spk_txs_seq":                    "spk_txs",
		"sop_jobs_seq":                   "sop_jobs",
		"flowcharts_seq":                 "flowcharts",
		"database_node_seq":              "database_nodes",
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
