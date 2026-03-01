package routes

import (
	"jk-api/api/http/controllers/v1"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func GraphRoutes(router fiber.Router, c *container.AppContainer) {
	app := router.Group("graphs")

	// GENERAL
	app.Get("/sops", controllers.GetSOPGraphs())
	app.Get("/props", controllers.GetGraphByProps())
	app.Post("/", controllers.CreateGraph())
	app.Put("/", controllers.UpdateMultipleGraph())
	app.Put("/merge", controllers.MergeGraph())
	app.Delete("/", controllers.DeleteGraph())
	app.Get("/label/:label", controllers.GetGraphByLabel())

	// DOCUMENT GRAPHS
	app.Get("/document/:id", controllers.GetDocumentGraph())

	// TABLE GRAPHS
	app.Post("/table/:elementId", controllers.CreateTableGraph())
	app.Put("/table/:elementId", controllers.UpdateTableGraph())

	// TEXT GRAPHS
	app.Post("/text/:elementId", controllers.CreateTextGraph())
	// app.Put("/text/:elementId", controllers.UpdateTextGraph())

	// COMMENT GRAPHS
	app.Get("/comment/:elementId", controllers.GetCommentGraph())
	app.Post("/comment/:elementId", controllers.CreateCommentGraph())

	// REVIEW GRAPHS
	app.Post("/review/:elementId", controllers.CreateReviewGraph())
	app.Get("/:id", controllers.GetGraphById())
	app.Put("/:id", controllers.UpdateGraph())
}
