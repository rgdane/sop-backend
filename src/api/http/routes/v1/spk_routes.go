package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/api/http/middleware"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func SpkRoutes(router fiber.Router, c *container.AppContainer) {
	app := router.Group("/spks", middleware.JWTMiddleware())

	app.Get("/", controllers.GetSpks(c))
	app.Post("/", controllers.CreateSpks(c))
	app.Post("/bulk-create", controllers.BulkCreateSpks(c))
	app.Put("/bulk-update", controllers.BulkUpdateSpks(c))
	app.Delete("/bulk-delete", controllers.BulkDeleteSpks(c))

	app.Get("/:id", controllers.GetSpkByID(c))
	app.Put("/:id", controllers.UpdateSpks(c))
	app.Delete("/:id", controllers.DeleteSpks(c))
}
