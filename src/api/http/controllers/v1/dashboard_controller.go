package controllers

import (
	"jk-api/api/http/presenters"
	"jk-api/internal/container"

	"github.com/gofiber/fiber/v2"
)

func GetDashboardCounts(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		data, err := cn.DashboardHandler.GetDashboardCountsHandler()
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}
