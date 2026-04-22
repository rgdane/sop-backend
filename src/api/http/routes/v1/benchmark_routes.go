package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func BenchmarkRoutes(router fiber.Router, c *container.AppContainer) {
	// Buat group /benchmark di bawah /api/v1 (atau sesuai prefix utama-mu)
	benchmark := router.Group("/benchmark")

	// ==========================================
	// 1. SKENARIO SEDERHANA (DIVISIONS)
	// ==========================================
	benchmark.Get("/divisions/sql", controllers.RunBenchmarkDivisionSQL(c))
	benchmark.Get("/divisions/graph", controllers.RunBenchmarkDivisionGraph(c))

	// ==========================================
	// 2. SKENARIO MENENGAH (SOPS)
	// ==========================================
	benchmark.Get("/sops/sql", controllers.RunBenchmarkSopSQL(c))
	benchmark.Get("/sops/graph", controllers.RunBenchmarkSopGraph(c))

	// ==========================================
	// 3. SKENARIO KOMPLEKS (SOP JOBS)
	// ==========================================
	benchmark.Get("/sop-jobs/sql", controllers.RunBenchmarkSopJobSQL(c))
	benchmark.Get("/sop-jobs/graph", controllers.RunBenchmarkSopJobGraph(c))
}