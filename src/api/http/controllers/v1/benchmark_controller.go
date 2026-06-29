package controllers

import (
	"jk-api/internal/container"
	"jk-api/internal/shared/helper"

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

func RunBenchmarkSopJobTitleSQL(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rate := c.QueryInt("rate", 100)
		duration := c.QueryInt("duration", 5)

		targetURL := "http://localhost:5000/api/v1/sop-jobs/title-sql/?title_name=Engineer"
		result := helper.RunVegetaLoadTest("GET", targetURL, "SQL - SOP Jobs by Title", duration, rate)
		return c.JSON(fiber.Map{"message": "Benchmark SQL SOP Jobs by Title selesai", "data": result})
	}
}

func RunBenchmarkSopJobTitleGraph(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rate := c.QueryInt("rate", 100)
		duration := c.QueryInt("duration", 5)

		targetURL := "http://localhost:5000/api/v1/sop-jobs/title-graph/?title_name=Engineer"
		result := helper.RunVegetaLoadTest("GET", targetURL, "Graph - SOP Jobs by Title", duration, rate)
		return c.JSON(fiber.Map{"message": "Benchmark Graph SOP Jobs by Title selesai", "data": result})
	}
}

func RunBenchmarkSopJobDivisionSQL(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rate := c.QueryInt("rate", 100)
		duration := c.QueryInt("duration", 5)

		targetURL := "http://localhost:5000/api/v1/sop-jobs/division-sql/?division_name=Product"
		result := helper.RunVegetaLoadTest("GET", targetURL, "SQL - SOP Jobs by Division", duration, rate)
		return c.JSON(fiber.Map{"message": "Benchmark SQL SOP Jobs by Division selesai", "data": result})
	}
}

func RunBenchmarkSopJobDivisionGraph(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rate := c.QueryInt("rate", 100)
		duration := c.QueryInt("duration", 5)

		targetURL := "http://localhost:5000/api/v1/sop-jobs/division-graph/?division_name=Product"
		result := helper.RunVegetaLoadTest("GET", targetURL, "Graph - SOP Jobs by Division", duration, rate)
		return c.JSON(fiber.Map{"message": "Benchmark Graph SOP Jobs by Division selesai", "data": result})
	}
}

func RunBenchmarkSopJobDivisionTitleSQL(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rate := c.QueryInt("rate", 100)
		duration := c.QueryInt("duration", 5)

		targetURL := "http://localhost:5000/api/v1/sop-jobs/division-title-sql/?division_name=Product&title_name=Engineer"
		result := helper.RunVegetaLoadTest("GET", targetURL, "SQL - SOP Jobs by Division & Title", duration, rate)
		return c.JSON(fiber.Map{"message": "Benchmark SQL SOP Jobs by Division & Title selesai", "data": result})
	}
}

func RunBenchmarkSopJobDivisionTitleGraph(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rate := c.QueryInt("rate", 100)
		duration := c.QueryInt("duration", 5)

		targetURL := "http://localhost:5000/api/v1/sop-jobs/division-title-graph/?division_name=Product&title_name=Engineer"
		result := helper.RunVegetaLoadTest("GET", targetURL, "Graph - SOP Jobs by Division & Title", duration, rate)
		return c.JSON(fiber.Map{"message": "Benchmark Graph SOP Jobs by Division & Title selesai", "data": result})
	}
}

func RunBenchmarkSopJobReferenceDivisionSQL(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rate := c.QueryInt("rate", 100)
		duration := c.QueryInt("duration", 5)

		targetURL := "http://localhost:5000/api/v1/sop-jobs/reference-division-sql/?reference_division_name=Product"
		result := helper.RunVegetaLoadTest("GET", targetURL, "SQL - SOP Jobs by Reference Division", duration, rate)
		return c.JSON(fiber.Map{"message": "Benchmark SQL SOP Jobs by Reference Division selesai", "data": result})
	}
}

func RunBenchmarkSopJobReferenceDivisionGraph(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rate := c.QueryInt("rate", 100)
		duration := c.QueryInt("duration", 5)

		targetURL := "http://localhost:5000/api/v1/sop-jobs/reference-division-graph/?reference_division_name=Product"
		result := helper.RunVegetaLoadTest("GET", targetURL, "Graph - SOP Jobs by Reference Division", duration, rate)
		return c.JSON(fiber.Map{"message": "Benchmark Graph SOP Jobs by Reference Division selesai", "data": result})
	}
}

func RunBenchmarkSopJobDivisionTitlePublishedSQL(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rate := c.QueryInt("rate", 100)
		duration := c.QueryInt("duration", 5)

		targetURL := "http://localhost:5000/api/v1/sop-jobs/division-title-published-sql/?division_name=Product&job_name_pattern=mengoptimalkan&spk_name=tim"
		result := helper.RunVegetaLoadTest("GET", targetURL, "SQL - SOP Jobs by Division/Title/Published", duration, rate)
		return c.JSON(fiber.Map{"message": "Benchmark SQL SOP Jobs by Division/Title/Published selesai", "data": result})
	}
}

func RunBenchmarkSopJobDivisionTitlePublishedGraph(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rate := c.QueryInt("rate", 100)
		duration := c.QueryInt("duration", 5)

		targetURL := "http://localhost:5000/api/v1/sop-jobs/division-title-published-graph/?division_name=Product&job_name_pattern=mengoptimalkan&spk_name=tim"
		result := helper.RunVegetaLoadTest("GET", targetURL, "Graph - SOP Jobs by Division/Title/Published", duration, rate)
		return c.JSON(fiber.Map{"message": "Benchmark Graph SOP Jobs by Division/Title/Published selesai", "data": result})
	}
}
