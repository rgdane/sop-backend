package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/api/http/middleware"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func DatabaseNodeRoutes(router fiber.Router, c *container.AppContainer) {
	app := router.Group("/database_nodes", middleware.JWTMiddleware())

	app.Get("/", controllers.GetDatabaseNodes(c))
	app.Get("/:id", controllers.GetDatabaseNodeByID(c))
	app.Post("/", controllers.CreateDatabaseNode(c))
	app.Put("/:id", controllers.UpdateDatabaseNode(c))
	app.Delete("/:id", controllers.DeleteDatabaseNode(c))
}
