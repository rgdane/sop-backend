package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func BenchmarkRoutes(router fiber.Router, c *container.AppContainer) {
	// Buat group /benchmark di bawah /api/v1
	benchmark := router.Group("/benchmark")

	// Pasang endpointnya (arahkan ke handler RunBenchmark yang kita buat sebelumnya)
	benchmark.Get("/test-sql", controllers.RunBenchmarkSQL)
	benchmark.Get("/test-graph", controllers.RunBenchmarkGraph)
}