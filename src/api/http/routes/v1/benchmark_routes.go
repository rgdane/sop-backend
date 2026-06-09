package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func BenchmarkRoutes(router fiber.Router, c *container.AppContainer) {
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

	// ==========================================
	// 4. SKENARIO FILTER: SOP JOBS BY TITLE
	// ==========================================
	benchmark.Get("/first/sql", controllers.RunBenchmarkSopJobTitleSQL(c))
	benchmark.Get("/first/graph", controllers.RunBenchmarkSopJobTitleGraph(c))

	// ==========================================
	// 5. SKENARIO FILTER: SOP JOBS BY DIVISION
	// ==========================================
	benchmark.Get("/second/sql", controllers.RunBenchmarkSopJobDivisionSQL(c))
	benchmark.Get("/second/graph", controllers.RunBenchmarkSopJobDivisionGraph(c))

	// ==========================================
	// 6. SKENARIO FILTER: SOP JOBS BY DIVISION & TITLE
	// ==========================================
	benchmark.Get("/third/sql", controllers.RunBenchmarkSopJobDivisionTitleSQL(c))
	benchmark.Get("/third/graph", controllers.RunBenchmarkSopJobDivisionTitleGraph(c))

	// ==========================================
	// 7. SKENARIO FILTER: SOP JOBS BY REFERENCE DIVISION
	// ==========================================
	benchmark.Get("/fourth/sql", controllers.RunBenchmarkSopJobReferenceDivisionSQL(c))
	benchmark.Get("/fourth/graph", controllers.RunBenchmarkSopJobReferenceDivisionGraph(c))

	// ==========================================
	// 8. SKENARIO FILTER: SOP JOBS BY DIVISION, TITLE COLOR & PUBLISHED
	// ==========================================
	benchmark.Get("/fifth/sql", controllers.RunBenchmarkSopJobDivisionTitlePublishedSQL(c))
	benchmark.Get("/fifth/graph", controllers.RunBenchmarkSopJobDivisionTitlePublishedGraph(c))
}