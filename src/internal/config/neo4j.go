package config

import (
	"context"
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/config"
)

var Neo4jDriver neo4j.DriverWithContext

func Neo4jApp() error {
	uri := AppConfig.Neo4jURI       // example: neo4j://localhost:7687
	user := AppConfig.Neo4jUser     // neo4j
	pass := AppConfig.Neo4jPassword // password
	timeout := 5 * time.Second

	driver, err := neo4j.NewDriverWithContext(
		uri,
		neo4j.BasicAuth(user, pass, ""),
		func(c *config.Config) {
			c.MaxConnectionPoolSize = 10
			c.SocketConnectTimeout = timeout
		},
	)

	if err != nil {
		return fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	if err := driver.VerifyConnectivity(context.Background()); err != nil {
		return fmt.Errorf("failed to connect to Neo4j: %w", err)
	}

	Neo4jDriver = driver
	return nil
}

func GetNeo4j() neo4j.DriverWithContext {
	return Neo4jDriver
}

func CloseNeo4j() {
	if Neo4jDriver != nil {
		Neo4jDriver.Close(context.Background())
		Logger.Info("🔌 Neo4j connection closed")
	}
}
