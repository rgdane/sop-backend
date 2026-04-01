package bootstrap

import (
	"jk-api/internal/config"
	"jk-api/internal/database/migrations"
	"jk-api/internal/database/seeders"
)

func Setup() {
	LoadConfig()

	InitLogger()
	InitPostgres()
	InitNeo4j()

	//runMigrate()
	InitFiber()
}

func runMigrate() {
	migrations.Migrate()
	seeders.InitSeeder(config.DB)
}
