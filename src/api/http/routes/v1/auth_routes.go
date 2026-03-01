package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/api/http/middleware"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(router fiber.Router, c *container.AppContainer) {
	app := router.Group("auth")

	app.Get("/profile", controllers.GetProfile(c), middleware.JWTMiddleware())
	app.Post("/login", controllers.Login(c))
}
