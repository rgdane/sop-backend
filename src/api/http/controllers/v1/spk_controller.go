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

func GetSpks(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		titleId, _ := helper.ParseQueryInt64(c, "title_id")
		cursor, _ := helper.ParseQueryInt64(c, "cursor")
		limit, _ := helper.ParseQueryInt64(c, "limit")
		name := c.Query("name", "")

		filter := dto.SpkFilterDto{
			TitleIDs:    titleId,
			Preload:     c.Query("preload", "false") == "true",
			Cursor:      cursor,
			Limit:       limit,
			Name:        name,
			ShowDeleted: c.Query("show_deleted", "false") == "true",
		}

		data, total, err := cn.SpkHandler.GetAllSpksHandler(filter)

		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		return presenters.SendSuccessResponse(c, data, total)
	}
}

func GetSpkByID(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filter := dto.SpkFilterDto{
			Preload: c.Query("preload", "false") == "true",
		}

		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		data, err := cn.SpkHandler.GetSpkByIDHandler(id, filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}

func CreateSpks(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.CreateSpkDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request")
		}

		result, err := cn.SpkHandler.CreateSpkHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, result)
	}
}

func UpdateSpks(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		var input dto.UpdateSpkDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid input")
		}

		updated, err := cn.SpkHandler.UpdateSpkHandler(id, &input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, updated)
	}
}

func DeleteSpks(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		if err := cn.SpkHandler.DeleteSpkHandler(id); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponseWithMessage(c, "Spk berhasil dihapus", nil)
	}
}

func BulkCreateSpks(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.BulkCreateSpksDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request body for bulk create")
		}
		if len(input.Data) == 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No data provided")
		}

		createdSpks, err := cn.SpkHandler.BulkCreateSpksHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, createdSpks)
	}
}

func BulkUpdateSpks(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.BulkUpdateSpkDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request body for bulk update")
		}
		if len(input.IDs) == 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No SPK IDs provided")
		}
		if input.Data == nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No update data provided")
		}

		updatedSpks, err := cn.SpkHandler.BulkUpdateSpksHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, updatedSpks)
	}
}

func BulkDeleteSpks(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.BulkDeleteSpkDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request body for bulk delete")
		}

		if len(input.IDs) == 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No Spk IDs provided")
		}

		err := cn.SpkHandler.BulkDeleteHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		return presenters.SendSuccessResponseWithMessage(c, fmt.Sprintf("Successfully deleted %d Spks", len(input.IDs)), nil)
	}
}
