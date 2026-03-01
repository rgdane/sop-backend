package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/api/http/middleware"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func DepartmentRoutes(router fiber.Router, c *container.AppContainer) {
	app := router.Group("/departments", middleware.JWTMiddleware())

	app.Get("/", controllers.GetDepartments(c))
	app.Post("/", controllers.CreateDepartments(c))
	app.Post("/bulk-create", controllers.BulkCreateDepartments(c))
	app.Put("/bulk-update", controllers.BulkUpdateDepartments(c))
	app.Delete("/bulk-delete", controllers.BulkDeleteDepartments(c))

	app.Get("/:id", controllers.GetDepartmentByID(c))
	app.Put("/:id", controllers.UpdateDepartments(c))
	app.Delete("/:id", controllers.DeleteDepartments(c))
}
