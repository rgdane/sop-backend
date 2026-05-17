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
			// Asumsi nama method handler-mu untuk graph
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
			// Asumsi kamu punya SopFilterDto, sesuaikan jika namanya beda
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

		refID := int64(31206)
		filter := dto.SopJobFilterDto{
			Preload:       true,
			SopName:       "100163",
			DivisionNames: []string{"Finance"},
			MinIndex:      5,
			ReferenceID:   &refID,
			ReferenceType: "spk",
			ShowDeleted:   false,
		}

		dbQueryFunc := func() time.Duration {
			start := time.Now()
			_, _, _ = cn.SopJobHandler.GetAllSopJobsHandler(filter)
			return time.Since(start)
		}
		result := helper.RunVegetaLoadTest("GET", targetURL, "SQL - Complex SOP Jobs", 5, 50, dbQueryFunc)
		return c.JSON(fiber.Map{"message": "Benchmark SQL SOP Jobs selesai", "data": result})
	}
}

func RunBenchmarkSopJobGraph(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		targetURL := "http://localhost:5000/api/v1/sop-jobs/graph/"

		refID := int64(31206)
		filter := dto.SopJobFilterDto{
			Preload:       true,
			SopName:       "100163",
			DivisionNames: []string{"Finance"},
			MinIndex:      5,
			ReferenceID:   &refID,
			ReferenceType: "spk",
			ShowDeleted:   false,
		}

		dbQueryFunc := func() time.Duration {
			start := time.Now()
			_, _, _ = cn.SopJobHandler.GetAllSopJobsGraphHandler(filter)
			return time.Since(start)
		}
		result := helper.RunVegetaLoadTest("GET", targetURL, "Graph - Complex SOP Jobs", 5, 50, dbQueryFunc)
		return c.JSON(fiber.Map{"message": "Benchmark Graph SOP Jobs selesai", "data": result})
	}
}
