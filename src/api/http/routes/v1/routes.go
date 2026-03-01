package routes

import (
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App, c *container.AppContainer) {
	api := app.Group("/api/v1")

	api.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"status":  "ok",
			"version": "v1",
			"message": "JalanKerja API",
		})
	})

	AuthRoutes(api, c)
	DepartmentRoutes(api, c)
	DivisionRoutes(api, c)
	LevelRoutes(api, c)
	PermissionRoutes(api, c)
	PositionRoutes(api, c)
	RoleRoutes(api, c)
	SopJobRoutes(api, c)
	SopRoutes(api, c)
	SpkJobRoutes(api, c)
	SpkRoutes(api, c)
	TitleRoutes(api, c)
	UserRoutes(api, c)
	GraphRoutes(api, c)
	DatabaseNodeRoutes(api, c)
}
