package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func BenchmarkRoutes(router fiber.Router, c *container.AppContainer) {
	benchmark := router.Group("/benchmark")

	// ==========================================
	// 1. SKENARIO FILTER: SOP JOBS BY TITLE
	// ==========================================
	benchmark.Get("/first/sql", controllers.RunBenchmarkSopJobTitleSQL(c))
	benchmark.Get("/first/graph", controllers.RunBenchmarkSopJobTitleGraph(c))

	// ==========================================
	// 2. SKENARIO FILTER: SOP JOBS BY DIVISION
	// ==========================================
	benchmark.Get("/second/sql", controllers.RunBenchmarkSopJobDivisionSQL(c))
	benchmark.Get("/second/graph", controllers.RunBenchmarkSopJobDivisionGraph(c))

	// ==========================================
	// 3. SKENARIO FILTER: SOP JOBS BY DIVISION & TITLE
	// ==========================================
	benchmark.Get("/third/sql", controllers.RunBenchmarkSopJobDivisionTitleSQL(c))
	benchmark.Get("/third/graph", controllers.RunBenchmarkSopJobDivisionTitleGraph(c))

	// ==========================================
	// 4. SKENARIO FILTER: SOP JOBS BY REFERENCE DIVISION
	// ==========================================
	benchmark.Get("/fourth/sql", controllers.RunBenchmarkSopJobReferenceDivisionSQL(c))
	benchmark.Get("/fourth/graph", controllers.RunBenchmarkSopJobReferenceDivisionGraph(c))

	// ==========================================
	// 5. SKENARIO FILTER: SOP JOBS BY DIVISION, TITLE COLOR & PUBLISHED
	// ==========================================
	benchmark.Get("/fifth/sql", controllers.RunBenchmarkSopJobDivisionTitlePublishedSQL(c))
	benchmark.Get("/fifth/graph", controllers.RunBenchmarkSopJobDivisionTitlePublishedGraph(c))
}