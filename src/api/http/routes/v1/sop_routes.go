package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func SopRoutes(router fiber.Router, c *container.AppContainer) {
	app := router.Group("/sops")

	app.Post("/bulk-create", controllers.BulkCreateSops(c))
	app.Put("/bulk-update", controllers.BulkUpdateSops(c))
	app.Delete("/bulk-delete", controllers.BulkDeleteSops(c))

	app.Get("/graph/", controllers.GetGraphSops(c))
	app.Get("/graph/:id", controllers.GetGraphSopByID(c))
	app.Post("/graph/", controllers.CreateGraphSop(c))
	app.Put("/graph/:id", controllers.UpdateGraphSop(c))
	app.Delete("/graph/:id", controllers.DeleteGraphSop(c))

	app.Get("/sql/", controllers.GetSqlSops(c))
	app.Get("/sql/:id", controllers.GetSqlSopByID(c))
	app.Post("/sql/", controllers.CreateSqlSop(c))
	app.Put("/sql/:id", controllers.UpdateSqlSop(c))
	app.Delete("/sql/:id", controllers.DeleteSqlSop(c))

	app.Get("/", controllers.GetSops(c))
	app.Post("/", controllers.CreateSop(c))
	app.Get("/count", controllers.CountSops(c))

	app.Get("/:id", controllers.GetSopByID(c))
	app.Put("/:id", controllers.UpdateSop(c))
	app.Delete("/:id", controllers.DeleteSop(c))
}
