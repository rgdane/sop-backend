package controllers

import (
	"jk-api/internal/shared/helper"

	"github.com/gofiber/fiber/v2"
)

// Bisa kamu taruh di file domain/models atau langsung di atas handler
type BenchmarkResult struct {
	TestName    string  `json:"test_name"`
	Target      string  `json:"target"`
	Requests    uint64  `json:"requests"`
	SuccessRate float64 `json:"success_rate"` // Persentase (0-100)
	MeanLatency string  `json:"mean_latency"` // Rata-rata response time API
	P99Latency  string  `json:"p99_latency"`  // Response time terburuk (99th percentile)
}

func RunBenchmarkSQL(c *fiber.Ctx) error {
	// Bisa juga ambil URL, durasi, atau rate dari query param jika ingin dinamis
	targetURL := "http://localhost:5000/api/v1/divisions/sql"

	// Panggil Helper
	result := helper.RunVegetaLoadTest("GET", targetURL, "SQL - Get Divisions", 5, 50)

	return c.JSON(fiber.Map{
		"message": "Benchmark SQL selesai",
		"data":    result,
	})
}

func RunBenchmarkGraph(c *fiber.Ctx) error {
	targetURL := "http://localhost:5000/api/v1/divisions/graph"
	
	// Panggil Helper dengan parameter berbeda
	result := helper.RunVegetaLoadTest("GET", targetURL, "Graph - Get Divisions", 5, 50)

	return c.JSON(fiber.Map{
		"message": "Benchmark Graph selesai",
		"data":    result,
	})
}
