package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/api/http/middleware"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func SopRoutes(router fiber.Router, c *container.AppContainer) {
	app := router.Group("/sops", middleware.JWTMiddleware())

	app.Get("/", controllers.GetSops(c))
	app.Post("/", controllers.CreateSop(c))
	app.Post("/bulk-create", controllers.BulkCreateSops(c))
	app.Put("/bulk-update", controllers.BulkUpdateSops(c))
	app.Delete("/bulk-delete", controllers.BulkDeleteSops(c))

	app.Get("/count", controllers.CountSops(c))
	app.Get("/:id", controllers.GetSopByID(c))
	app.Put("/:id", controllers.UpdateSop(c))
	app.Delete("/:id", controllers.DeleteSop(c))
}
