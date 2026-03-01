package config

import (
	"jk-api/docs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

func NewFiberApp() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:         AppConfig.AppName,
		ServerHeader:    AppConfig.SwaggerHost,
		ReadBufferSize:  16 * 1024,
		WriteBufferSize: 16 * 1024,
		BodyLimit:       10 * 1024 * 1024,

		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			if c.Get("Upgrade") == "websocket" {
				log.Errorf("WebSocket Error [%s %s]: %v", c.Method(), c.Path(), err)
			} else {
				log.Errorf("HTTP Error [%s %s]: %v", c.Method(), c.Path(), err)
			}

			return c.Status(code).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		},
	})

	setupMiddleware(app)
	return app
}

func InitFiberApp() *fiber.App {
	docs.SwaggerInfo.Host = AppConfig.SwaggerHost

	app := NewFiberApp()
	return app
}

func setupMiddleware(app *fiber.App) {
	app.Use(recover.New())

	app.Use(logger.New(logger.Config{
		Format:     "[${time}] [${status}] ${method} ${path} - ${latency}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Asia/Jakarta",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,Omni-Token",
		AllowCredentials: false,
		MaxAge:           86400,
	}))

	app.Get("/swagger/*", fiberSwagger.WrapHandler)
}
