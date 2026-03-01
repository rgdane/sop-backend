package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var AppConfig *Config

type Config struct {
	AppName string
	AppPort string
	AppHost string

	SwaggerHost string

	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresConn     string
	PostgresSSLMode  string

	GCPBucketName string

	Neo4jURI      string
	Neo4jUser     string
	Neo4jPassword string

	OmniChannelURI string
}

func LoadConfig() error {
	_ = godotenv.Load()

	AppConfig = &Config{
		AppName:     getEnv("APP_NAME", "JalanKerja API"),
		AppPort:     getEnv("PORT", "8080"),
		AppHost:     getEnv("HOST", "localhost"),
		SwaggerHost: getEnv("SWAGGER_HOST", "localhost:5000"),

		PostgresHost:     getEnv("DB_HOST", "localhost"),
		PostgresPort:     getEnv("DB_PORT", "5432"),
		PostgresUser:     getEnv("DB_USERNAME", "postgres"),
		PostgresPassword: getEnv("DB_PASSWORD", ""),
		PostgresDB:       getEnv("DB_DATABASE", "jalankerja"),
		PostgresConn:     getEnv("DB_CONNECTION", "postgres"),
		PostgresSSLMode:  getEnv("DB_SSLMODE", "disable"),

		GCPBucketName: getEnv("GCP_BUCKET_NAME", ""),

		Neo4jURI:      getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUser:     getEnv("NEO4J_USER", "neo4j"),
		Neo4jPassword: getEnv("NEO4J_PASSWORD", "password"),

		OmniChannelURI: getEnv("OMNI_CHANNEL_URI", "http://localhost:3000"),
	}

	return nil
}

func GetPostgresUrl() string {
	return getDsn()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDsn() string {
	host := AppConfig.PostgresHost
	user := AppConfig.PostgresUser
	password := AppConfig.PostgresPassword
	dbname := AppConfig.PostgresDB
	port := AppConfig.PostgresPort
	conn := AppConfig.PostgresConn
	ssl_mode := AppConfig.PostgresSSLMode

	if host == "" || user == "" || dbname == "" || port == "" || conn == "" {
		logrus.Fatal("❌ Missing required DB config in .env")
		os.Exit(1)
	}

	dsn := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s",
		conn, user, password, host, port, dbname, ssl_mode)

	return dsn
}
