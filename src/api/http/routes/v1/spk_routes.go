package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/api/http/middleware"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func SpkRoutes(router fiber.Router, c *container.AppContainer) {
	app := router.Group("/spks", middleware.JWTMiddleware())

	app.Post("/bulk-create", controllers.BulkCreateSpks(c))
	app.Put("/bulk-update", controllers.BulkUpdateSpks(c))
	app.Delete("/bulk-delete", controllers.BulkDeleteSpks(c))

	app.Get("/graph/", controllers.GetGraphSpks(c))
	app.Get("/graph/:id", controllers.GetGraphSpkByID(c))
	app.Post("/graph/", controllers.CreateGraphSpk(c))
	app.Put("/graph/:id", controllers.UpdateGraphSpk(c))
	app.Delete("/graph/:id", controllers.DeleteGraphSpk(c))

	app.Get("/sql/", controllers.GetSqlSpks(c))
	app.Get("/sql/:id", controllers.GetSpkByID(c))
	app.Post("/sql/", controllers.CreateSqlSpk(c))
	app.Put("/sql/:id", controllers.UpdateSqlSpk(c))
	app.Delete("/sql/:id", controllers.DeleteSqlSpk(c))

	app.Get("/", controllers.GetSpks(c))
	app.Post("/", controllers.CreateSpk(c))

	app.Get("/:id", controllers.GetSpkByID(c))
	app.Put("/:id", controllers.UpdateSpk(c))
	app.Delete("/:id", controllers.DeleteSpk(c))
}
