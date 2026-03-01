// database_node_controller.go - Dibuat berdasarkan typography_category_controller.go

package controllers

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/api/http/presenters"
	"jk-api/internal/container"
	"jk-api/internal/shared/helper"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func GetDatabaseNodes(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		cursor, _ := helper.ParseQueryInt64(c, "cursor")
		limit, _ := helper.ParseQueryInt64(c, "limit")
		search := c.Query("search")

		filter := dto.DatabaseNodeFilter{
			Preload:     c.Query("preload", "false") == "true",
			Cursor:      cursor,
			Limit:       limit,
			Search:      search,
			ShowDeleted: c.Query("show_deleted", "false") == "true",
		}

		data, total, err := cn.DatabaseNodeHandler.GetAllDatabaseNodesHandler(filter)

		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		return presenters.SendSuccessResponse(c, data, total)
	}
}

func GetDatabaseNodeByID(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filter := dto.DatabaseNodeFilter{
			Preload: c.Query("preload", "false") == "true",
		}

		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		data, err := cn.DatabaseNodeHandler.GetDatabaseNodeByIDHandler(id, filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}

func CreateDatabaseNode(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.CreateDatabaseNodeDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request")
		}
		fmt.Printf("Creating database node with input: %+v\n", input)

		result, err := cn.DatabaseNodeHandler.CreateDatabaseNodeHandler(&input, c)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, result)
	}
}

func UpdateDatabaseNode(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		var input dto.UpdateDatabaseNodeDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid input")
		}

		updated, err := cn.DatabaseNodeHandler.UpdateDatabaseNodeHandler(id, &input, c)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, updated)
	}
}

func DeleteDatabaseNode(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		if err := cn.DatabaseNodeHandler.DeleteDatabaseNodeHandler(id, c); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponseWithMessage(c, "DatabaseNode deleted successfully", nil)
	}
}
