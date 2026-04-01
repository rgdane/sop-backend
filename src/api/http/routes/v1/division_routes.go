package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/api/http/middleware"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func DivisionRoutes(router fiber.Router, c *container.AppContainer) {
	app := router.Group("/divisions", middleware.JWTMiddleware())

	// 1. RUTE KHUSUS / STATIS (Daftarkan paling atas)
	app.Post("/bulk-create", controllers.BulkCreateDivisions(c))
	app.Put("/bulk-update", controllers.BulkUpdateDivisions(c))
	app.Delete("/bulk-delete", controllers.BulkDeleteDivisions(c))

	// GRAPH ENDPOINTS
	app.Get("/graph/", controllers.GetGraphDivisions(c))
	app.Get("/graph/:id", controllers.GetGraphDivisionByID(c))
	app.Post("/graph/", controllers.CreateGraphDivision(c))
	app.Put("/graph/:id", controllers.UpdateGraphDivision(c))
	app.Delete("/graph/:id", controllers.DeleteGraphDivision(c))

	// SQL ONLY ENDPOINTS
	app.Get("/sql/", controllers.GetDivisions(c))
	app.Get("/sql/:id", controllers.GetDivisionByID(c))
	app.Post("/sql/", controllers.CreateSqlDivision(c))
	app.Put("/sql/:id", controllers.UpdateSqlDivision(c))
	app.Delete("/sql/:id", controllers.DeleteSqlDivision(c))

	// 2. RUTE DASAR & DINAMIS (Daftarkan paling bawah)
	app.Get("/", controllers.GetDivisions(c))
	app.Post("/", controllers.CreateDivision(c))
	
	// Parameter :id harus selalu di paling bawah agar tidak menangkap rute lain!
	app.Get("/:id", controllers.GetDivisionByID(c))
	app.Put("/:id", controllers.UpdateDivision(c))
	app.Delete("/:id", controllers.DeleteDivision(c))
}