  package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/api/http/middleware"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func SopRoutes(router fiber.Router, c *container.AppContainer) {
	app := router.Group("/sops", middleware.JWTMiddleware())

	// 1. RUTE KHUSUS / STATIS (Daftarkan paling atas)
	app.Post("/bulk-create", controllers.BulkCreateSops(c))
	app.Put("/bulk-update", controllers.BulkUpdateSops(c))
	app.Delete("/bulk-delete", controllers.BulkDeleteSops(c))

	// GRAPH ENDPOINTS
	app.Get("/graph/", controllers.GetGraphSops(c))
	app.Get("/graph/:id", controllers.GetGraphSopByID(c))
	app.Post("/graph/", controllers.CreateGraphSop(c))
	app.Put("/graph/:id", controllers.UpdateGraphSop(c))
	app.Delete("/graph/:id", controllers.DeleteGraphSop(c))

	// SQL ONLY ENDPOINTS
	app.Get("/sql/", controllers.GetSops(c))
	app.Get("/sql/:id", controllers.GetSopByID(c))
	app.Post("/sql/", controllers.CreateSqlSop(c))
	app.Put("/sql/:id", controllers.UpdateSqlSop(c))
	app.Delete("/sql/:id", controllers.DeleteSqlSop(c))

	// 2. RUTE DASAR & DINAMIS (Daftarkan paling bawah)
	app.Get("/", controllers.GetSops(c))
	app.Post("/", controllers.CreateSop(c))
	
	// Parameter :id harus selalu di paling bawah agar tidak menangkap rute lain!
	app.Get("/:id", controllers.GetSopByID(c))
	app.Put("/:id", controllers.UpdateSop(c))
	app.Delete("/:id", controllers.DeleteSop(c))
}