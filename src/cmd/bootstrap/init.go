package bootstrap

import (
	"jk-api/api/http/routes/v1"
	"jk-api/internal/config"
	"jk-api/internal/container"
)

func InitLogger() {
	config.StartLogger()
}

func LoadConfig() {
	config.LoadConfig()
}

func InitNeo4j() {
	if err := config.Neo4jApp(); err != nil {
		config.Logger.Fatalf("❌ Failed to initialize Neo4j: %v", err)
		return
	}
	config.Logger.Info("✅ Neo4j initialized")
}

func InitStorage() {
	if err := config.GCPBucketApp(nil); err != nil {
		config.Logger.Fatalf("❌ Failed to initialize GCP bucket: %v", err)
		return
	}
	config.Logger.Info("✅ GCP bucket initialized")
}

func InitFirebaseApp() {
	if err := config.FirebaseApp(); err != nil {
		config.Logger.Fatalf("❌ Failed to initialize Firebase app: %v", err)
		return
	}
	config.Logger.Info("✅ Firebase app initialized")
}

func InitPostgres() {
	if err := config.PostgresApp(); err != nil {
		config.Logger.Fatalf("❌ Failed to initialize Postgres: %v", err)
		return
	}
	config.Logger.Info("✅ Postgres initialized")
}

func InitFiber() {
	app := config.InitFiberApp()
	routes.Setup(app, container.NewAppContainer())
	config.Logger.Infof("✅ REST API started on port %s", config.AppConfig.AppPort)

	if err := app.Listen(":" + config.AppConfig.AppPort); err != nil {
		config.Logger.Fatalf("❌ Failed to start Fiber: %v", err)
		return
	}
}

func InitOmniChannel() {
	config.InitOmniChannelClient()
	config.Logger.Info("✅ OmniChannel initialized")
}
