package seeders

import "gorm.io/gorm"

func InitSeeder(db *gorm.DB) {
	SeedPermissions(db)
	SeedRoles(db)
	SeedAdmin(db)
	// SeedFlowcharts(db)
	// SeedDepartments(db)
	// SeedDivisions(db)
	// SeedLevels(db)
	// SeedPositions(db)
	// SeedTitles(db)
	// SeedProjects(db)
	// SeedSquads(db)
	// SeedSops(db)
	// SeedDatabaseNodes(db)

	// Case Report Seeders
	// Uncomment to seed case report related data
	// SeedCaseStatuses(db)
	// SeedCaseCategories(db)
	// SeedCaseBugFeatures(db)
	// SeedCaseReports(db) // Seeds 1000 case reports with searchable data

	// TODO: Buat error check setiap seeder karyawan memiliki return error
}
