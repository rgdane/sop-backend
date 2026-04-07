package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/api/http/middleware"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func TitleRoutes(router fiber.Router, c *container.AppContainer) {
	app := router.Group("/titles", middleware.JWTMiddleware())

	// 1. RUTE KHUSUS / STATIS (Daftarkan paling atas)
	app.Post("/bulk-create", controllers.BulkCreateTitles(c))
	app.Put("/bulk-update", controllers.BulkUpdateTitles(c))
	app.Delete("/bulk-delete", controllers.BulkDeleteTitles(c))

	// GRAPH ENDPOINTS
	app.Get("/graph/", controllers.GetGraphTitles(c))
	app.Get("/graph/:id", controllers.GetGraphTitleByID(c))
	app.Post("/graph/", controllers.CreateGraphTitle(c))
	app.Put("/graph/:id", controllers.UpdateGraphTitle(c))
	app.Delete("/graph/:id", controllers.DeleteGraphTitle(c))

	// SQL ONLY ENDPOINTS
	app.Get("/sql/", controllers.GetTitles(c))
	app.Get("/sql/:id", controllers.GetTitleByID(c))
	app.Post("/sql/", controllers.CreateSqlTitle(c))
	app.Put("/sql/:id", controllers.UpdateSqlTitle(c))
	app.Delete("/sql/:id", controllers.DeleteSqlTitle(c))

	// 2. RUTE DASAR & DINAMIS (Daftarkan paling bawah)
	app.Get("/", controllers.GetTitles(c))
	app.Post("/", controllers.CreateTitle(c))
	
	// Parameter :id harus selalu di paling bawah agar tidak menangkap rute lain!
	app.Get("/:id", controllers.GetTitleByID(c))
	app.Put("/:id", controllers.UpdateTitle(c))
	app.Delete("/:id", controllers.DeleteTitle(c))
}