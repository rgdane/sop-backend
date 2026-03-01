package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/api/http/middleware"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func DivisionRoutes(router fiber.Router, c *container.AppContainer) {
	app := router.Group("/divisions", middleware.JWTMiddleware())

	app.Post("/bulk-create", controllers.BulkCreateDivisions(c))
	app.Put("/bulk-update", controllers.BulkUpdateDivisions(c))
	app.Delete("/bulk-delete", controllers.BulkDeleteDivisions(c))

	app.Get("/", controllers.GetDivisions(c))
	app.Get("/:id", controllers.GetDivisionByID(c))
	app.Post("/", controllers.CreateDivisions(c))
	app.Put("/:id", controllers.UpdateDivisions(c))
	app.Delete("/:id", controllers.DeleteDivisions(c))
}
