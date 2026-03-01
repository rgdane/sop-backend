package bootstrap

func Setup() {
	LoadConfig()

	InitLogger()
	InitPostgres()
	InitFirebaseApp()
	InitStorage()
	InitNeo4j()
	InitOmniChannel()

	runMigrate()
	InitFiber()
}

func runMigrate() {
	// migrations.Migrate()
	// seeders.InitSeeder(config.DB)
}
