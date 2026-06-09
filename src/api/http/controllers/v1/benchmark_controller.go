package controllers

import (
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/internal/container"
	"jk-api/internal/shared/helper"
	"time"

	"github.com/gofiber/fiber/v2"
)

type BenchmarkResult struct {
	TestName    string  `json:"test_name"`
	Target      string  `json:"target"`
	Requests    uint64  `json:"requests"`
	SuccessRate float64 `json:"success_rate"`
	MeanLatency string  `json:"mean_latency"`
	P99Latency  string  `json:"p99_latency"`
}

// =================================================================
// 1. SKENARIO SEDERHANA: DIVISIONS (Tanpa banyak relasi)
// =================================================================

func RunBenchmarkDivisionSQL(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "http://localhost:5000/api/v1/divisions/sql"
		dbQueryFunc := func() time.Duration {
			start := time.Now()
			_, _, _ = cn.DivisionHandler.GetAllDivisionsHandler(dto.DivisionFilterDto{})
			return time.Since(start)
		}
		result := helper.RunVegetaLoadTest("GET", targetURL, "SQL - Get Divisions", 5, 50, dbQueryFunc)
		return c.JSON(fiber.Map{"message": "Benchmark SQL Divisions selesai", "data": result})
	}
}

func RunBenchmarkDivisionGraph(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "http://localhost:5000/api/v1/divisions/graph"
		dbQueryFunc := func() time.Duration {
			start := time.Now()
			_, _, _ = cn.DivisionHandler.GetAllDivisionsGraphHandler(dto.DivisionFilterDto{})
			return time.Since(start)
		}
		result := helper.RunVegetaLoadTest("GET", targetURL, "Graph - Get Divisions", 5, 50, dbQueryFunc)
		return c.JSON(fiber.Map{"message": "Benchmark Graph Divisions selesai", "data": result})
	}
}

// =================================================================
// 2. SKENARIO MENENGAH: SOP (Relasi ke Divisi)
// =================================================================

func RunBenchmarkSopSQL(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "http://localhost:5000/api/v1/sops/sql/"
		dbQueryFunc := func() time.Duration {
			start := time.Now()
			_, _, _ = cn.SopHandler.GetAllSopsHandler(dto.SopFilterDto{})
			return time.Since(start)
		}
		result := helper.RunVegetaLoadTest("GET", targetURL, "SQL - Get All SOPs", 5, 50, dbQueryFunc)
		return c.JSON(fiber.Map{"message": "Benchmark SQL SOPs selesai", "data": result})
	}
}

func RunBenchmarkSopGraph(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "http://localhost:5000/api/v1/sops/graph/"
		dbQueryFunc := func() time.Duration {
			start := time.Now()
			_, _, _ = cn.SopHandler.GetAllSopsGraphHandler(dto.SopFilterDto{})
			return time.Since(start)
		}
		result := helper.RunVegetaLoadTest("GET", targetURL, "Graph - Get All SOPs", 5, 50, dbQueryFunc)
		return c.JSON(fiber.Map{"message": "Benchmark Graph SOPs selesai", "data": result})
	}
}

// =================================================================
// 3. SKENARIO KOMPLEKS: SOP JOBS (Struktur Linked-List & Multi-Join)
// =================================================================

func RunBenchmarkSopJobSQL(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "http://localhost:5000/api/v1/sop-jobs/sql/"

		filter := dto.SopJobFilterDto{
			Preload:       true,
			SopName:       "brainstorming",
			DivisionNames: []string{"Product"},
			Limit:         10,
			ReferenceType: "spk",
			ShowDeleted:   false,
		}

		dbQueryFunc := func() time.Duration {
			start := time.Now()
			_, _, _ = cn.SopJobHandler.GetAllSopJobsHandler(filter)
			return time.Since(start)
		}
		result := helper.RunVegetaLoadTest("GET", targetURL, "SQL - Complex SOP Jobs", 5, 20, dbQueryFunc)
		return c.JSON(fiber.Map{"message": "Benchmark SQL SOP Jobs selesai", "data": result})
	}
}

func RunBenchmarkSopJobGraph(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "http://localhost:5000/api/v1/sop-jobs/graph/"

		filter := dto.SopJobFilterDto{
			Preload:       true,
			SopName:       "brainstorming",
			DivisionNames: []string{"Product"},
			Limit:         10,
			ReferenceType: "spk",
			ShowDeleted:   false,
		}

		dbQueryFunc := func() time.Duration {
			start := time.Now()
			_, _, _ = cn.SopJobHandler.GetAllSopJobsGraphHandler(filter)
			return time.Since(start)
		}
		result := helper.RunVegetaLoadTest("GET", targetURL, "Graph - Complex SOP Jobs", 5, 20, dbQueryFunc)
		return c.JSON(fiber.Map{"message": "Benchmark Graph SOP Jobs selesai", "data": result})
	}
}

// =================================================================
// 4. SKENARIO FILTER: SOP JOBS BY TITLE
// =================================================================

func RunBenchmarkSopJobTitleSQL(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "http://localhost:5000/api/v1/sop-jobs/title-sql/?title_name=Product"

		titleName := "Product"

		dbQueryFunc := func() time.Duration {
			start := time.Now()
			_, _, _ = cn.SopJobHandler.GetJobsByTitleNameSqlHandler(titleName)
			return time.Since(start)
		}
		result := helper.RunVegetaLoadTest("GET", targetURL, "SQL - SOP Jobs by Title", 5, 20, dbQueryFunc)
		return c.JSON(fiber.Map{"message": "Benchmark SQL SOP Jobs by Title selesai", "data": result})
	}
}

func RunBenchmarkSopJobTitleGraph(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "http://localhost:5000/api/v1/sop-jobs/title-graph/?title_name=Product"

		titleName := "Product"

		dbQueryFunc := func() time.Duration {
			start := time.Now()
			_, _, _ = cn.SopJobHandler.GetJobsByTitleNameGraphHandler(titleName)
			return time.Since(start)
		}
		result := helper.RunVegetaLoadTest("GET", targetURL, "Graph - SOP Jobs by Title", 5, 20, dbQueryFunc)
		return c.JSON(fiber.Map{"message": "Benchmark Graph SOP Jobs by Title selesai", "data": result})
	}
}

// =================================================================
// 5. SKENARIO FILTER: SOP JOBS BY DIVISION
// =================================================================

func RunBenchmarkSopJobDivisionSQL(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "http://localhost:5000/api/v1/sop-jobs/division-sql/?division_name=Product"

		divisionName := "Product"

		dbQueryFunc := func() time.Duration {
			start := time.Now()
			_, _, _ = cn.SopJobHandler.GetJobsByDivisionNameSqlHandler(divisionName)
			return time.Since(start)
		}
		result := helper.RunVegetaLoadTest("GET", targetURL, "SQL - SOP Jobs by Division", 5, 20, dbQueryFunc)
		return c.JSON(fiber.Map{"message": "Benchmark SQL SOP Jobs by Division selesai", "data": result})
	}
}

func RunBenchmarkSopJobDivisionGraph(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "http://localhost:5000/api/v1/sop-jobs/division-graph/?division_name=Product"

		divisionName := "Product"

		dbQueryFunc := func() time.Duration {
			start := time.Now()
			_, _, _ = cn.SopJobHandler.GetJobsByDivisionNameGraphHandler(divisionName)
			return time.Since(start)
		}
		result := helper.RunVegetaLoadTest("GET", targetURL, "Graph - SOP Jobs by Division", 5, 20, dbQueryFunc)
		return c.JSON(fiber.Map{"message": "Benchmark Graph SOP Jobs by Division selesai", "data": result})
	}
}

// =================================================================
// 6. SKENARIO FILTER: SOP JOBS BY DIVISION & TITLE
// =================================================================

func RunBenchmarkSopJobDivisionTitleSQL(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "http://localhost:5000/api/v1/sop-jobs/division-title-sql/?division_name=Product&title_name=Product"

		divisionName := "Product"
		titleName := "Product"

		dbQueryFunc := func() time.Duration {
			start := time.Now()
			_, _, _ = cn.SopJobHandler.GetJobsByDivisionAndTitleSqlHandler(divisionName, titleName)
			return time.Since(start)
		}
		result := helper.RunVegetaLoadTest("GET", targetURL, "SQL - SOP Jobs by Division & Title", 5, 20, dbQueryFunc)
		return c.JSON(fiber.Map{"message": "Benchmark SQL SOP Jobs by Division & Title selesai", "data": result})
	}
}

func RunBenchmarkSopJobDivisionTitleGraph(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "http://localhost:5000/api/v1/sop-jobs/division-title-graph/?division_name=Product&title_name=Product"

		divisionName := "Product"
		titleName := "Product"

		dbQueryFunc := func() time.Duration {
			start := time.Now()
			_, _, _ = cn.SopJobHandler.GetJobsByDivisionAndTitleGraphHandler(divisionName, titleName)
			return time.Since(start)
		}
		result := helper.RunVegetaLoadTest("GET", targetURL, "Graph - SOP Jobs by Division & Title", 5, 20, dbQueryFunc)
		return c.JSON(fiber.Map{"message": "Benchmark Graph SOP Jobs by Division & Title selesai", "data": result})
	}
}

// =================================================================
// 7. SKENARIO FILTER: SOP JOBS BY REFERENCE DIVISION
// =================================================================

func RunBenchmarkSopJobReferenceDivisionSQL(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "http://localhost:5000/api/v1/sop-jobs/reference-division-sql/?reference_division_name=Product"

		divisionName := "Product"

		dbQueryFunc := func() time.Duration {
			start := time.Now()
			_, _, _ = cn.SopJobHandler.GetJobsByReferenceDivisionNameSqlHandler(divisionName)
			return time.Since(start)
		}
		result := helper.RunVegetaLoadTest("GET", targetURL, "SQL - SOP Jobs by Reference Division", 5, 20, dbQueryFunc)
		return c.JSON(fiber.Map{"message": "Benchmark SQL SOP Jobs by Reference Division selesai", "data": result})
	}
}

func RunBenchmarkSopJobReferenceDivisionGraph(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "http://localhost:5000/api/v1/sop-jobs/reference-division-graph/?reference_division_name=Product"

		divisionName := "Product"

		dbQueryFunc := func() time.Duration {
			start := time.Now()
			_, _, _ = cn.SopJobHandler.GetJobsByReferenceDivisionNameGraphHandler(divisionName)
			return time.Since(start)
		}
		result := helper.RunVegetaLoadTest("GET", targetURL, "Graph - SOP Jobs by Reference Division", 5, 20, dbQueryFunc)
		return c.JSON(fiber.Map{"message": "Benchmark Graph SOP Jobs by Reference Division selesai", "data": result})
	}
}

// =================================================================
// 8. SKENARIO FILTER: SOP JOBS BY DIVISION, TITLE COLOR & PUBLISHED
// =================================================================

func RunBenchmarkSopJobDivisionTitlePublishedSQL(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "http://localhost:5000/api/v1/sop-jobs/division-title-published-sql/?division_name=Product&job_name_pattern=mengoptimalkan&title_color=%23FF5733"

		divisionName := "Product"
		jobNamePattern := "mengoptimalkan"
		titleColor := "#FF5733"

		dbQueryFunc := func() time.Duration {
			start := time.Now()
			_, _, _ = cn.SopJobHandler.GetJobsByDivisionTitlePublishedSqlHandler(divisionName, jobNamePattern, titleColor)
			return time.Since(start)
		}
		result := helper.RunVegetaLoadTest("GET", targetURL, "SQL - SOP Jobs by Division/Title/Published", 5, 20, dbQueryFunc)
		return c.JSON(fiber.Map{"message": "Benchmark SQL SOP Jobs by Division/Title/Published selesai", "data": result})
	}
}

func RunBenchmarkSopJobDivisionTitlePublishedGraph(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "http://localhost:5000/api/v1/sop-jobs/division-title-published-graph/?division_name=Product&job_name_pattern=mengoptimalkan&title_color=%23FF5733"

		divisionName := "Product"
		jobNamePattern := "mengoptimalkan"
		titleColor := "#FF5733"

		dbQueryFunc := func() time.Duration {
			start := time.Now()
			_, _, _ = cn.SopJobHandler.GetJobsByDivisionTitlePublishedGraphHandler(divisionName, jobNamePattern, titleColor)
			return time.Since(start)
		}
		result := helper.RunVegetaLoadTest("GET", targetURL, "Graph - SOP Jobs by Division/Title/Published", 5, 20, dbQueryFunc)
		return c.JSON(fiber.Map{"message": "Benchmark Graph SOP Jobs by Division/Title/Published selesai", "data": result})
	}
}
