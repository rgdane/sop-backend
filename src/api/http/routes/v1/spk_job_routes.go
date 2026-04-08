package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/api/http/middleware"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func SpkJobRoutes(router fiber.Router, c *container.AppContainer) {
	app := router.Group("/spk-jobs", middleware.JWTMiddleware())

	app.Post("/bulk-create", controllers.BulkCreateSpkJobs(c))
	app.Put("/bulk-update", controllers.BulkUpdateSpkJobs(c))
	app.Delete("/bulk-delete", controllers.BulkDeleteSpkJobs(c))

	app.Get("/graph/", controllers.GetGraphSpkJobs(c))
	app.Get("/graph/:id", controllers.GetGraphSpkJobByID(c))
	app.Post("/graph/", controllers.CreateGraphSpkJob(c))
	app.Put("/graph/:id", controllers.UpdateGraphSpkJobs(c))
	app.Delete("/graph/:id", controllers.DeleteGraphSpkJob(c))

	app.Get("/sql/", controllers.GetSpkJobs(c))
	app.Get("/sql/:id", controllers.GetSpkJobByID(c))
	app.Post("/sql/", controllers.CreateSqlSpkJob(c))
	app.Put("/sql/:id", controllers.UpdateSqlSpkJobs(c))
	app.Delete("/sql/:id", controllers.DeleteSqlSpkJob(c))

	app.Get("/", controllers.GetSpkJobs(c))
	app.Post("/", controllers.CreateSpkJob(c))

	app.Get("/:id", controllers.GetSpkJobByID(c))
	app.Put("/:id", controllers.UpdateSpkJobs(c))
	app.Put("/:id/reorder", controllers.ReorderSpkJob(c))
	app.Delete("/:id", controllers.DeleteSpkJob(c))
}
