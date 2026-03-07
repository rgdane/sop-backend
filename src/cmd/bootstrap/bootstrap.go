package bootstrap

import (
	"jk-api/internal/database/migrations"
)

func Setup() {
	LoadConfig()

	InitLogger()
	InitPostgres()
	InitNeo4j()

	runMigrate()
	InitFiber()
}

func runMigrate() {
	migrations.Migrate()
	//seeders.InitSeeder(config.DB)
}
