package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/api/http/middleware"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func UserRoutes(router fiber.Router, c *container.AppContainer) {
	app := router.Group("users", middleware.JWTMiddleware())

	app.Get("/", controllers.GetUsers(c))
	app.Get("/:id", controllers.GetUserByID(c))
	app.Post("/bulk-create", controllers.BulkCreateUsers(c))
	app.Put("/bulk-update", controllers.BulkUpdateUsers(c))
	app.Delete("/bulk-delete", controllers.BulkDeleteUsers(c))
	app.Post("/", controllers.CreateUsers(c))
	app.Put("/:id", controllers.UpdateUsers(c))
	app.Delete("/:id", controllers.DeleteUsers(c))
}
