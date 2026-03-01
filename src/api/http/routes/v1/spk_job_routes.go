package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/api/http/middleware"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func SpkJobRoutes(router fiber.Router, c *container.AppContainer) {
	app := router.Group("/spk-jobs", middleware.JWTMiddleware())

	app.Get("/", controllers.GetSpkJobs(c))
	app.Post("/bulk-create", controllers.BulkCreateSpkJobs(c))
	app.Post("/", controllers.CreateSpkJob(c))
	app.Delete("/bulk-delete", controllers.BulkDeleteSpkJobs(c))

	app.Get("/:id", controllers.GetSpkByID(c))
	app.Put("/:id", controllers.UpdateSpkJobs(c))
	app.Put("/:id/reorder", controllers.ReorderSpkJob(c))
	app.Delete("/:id", controllers.DeleteSpkJob(c))
}
