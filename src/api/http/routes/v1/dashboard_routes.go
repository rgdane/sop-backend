package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func DashboardRoutes(router fiber.Router, c *container.AppContainer) {
	app := router.Group("/dashboard")

	app.Get("/counts", controllers.GetDashboardCounts(c))
}
