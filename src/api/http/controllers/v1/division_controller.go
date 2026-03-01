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

// GetDivisions godoc
//
//	@Summary		Get all divisions
//	@Description	Get list of divisions with optional filters
//	@Tags			divisions
//	@Accept			json
//	@Produce		json
//	@Param			department_id	query	int64	false	"Department ID filter"
//	@Param			sop_id			query	int64	false	"SOP ID filter"
//	@Param			sort			query	string	false	"Sort field"
//	@Param			order			query	string	false	"Sort order"
//	@Param			cursor			query	int64	false	"Cursor for pagination"
//	@Param			limit			query	int64	false	"Limit for pagination"
//	@Param			name			query	string	false	"Name filter"
//	@Param			show_deleted	query	bool	false	"Show deleted divisions"
//	@Param			preload			query	bool	false	"Preload relations"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/divisions [get]
func GetDivisions(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		departmentId, _ := helper.ParseQueryInt64(c, "department_id")
		sopId, _ := helper.ParseQueryInt64(c, "sop_id")
		sort := c.Query("sort")
		order := c.Query("order")
		cursor, _ := helper.ParseQueryInt64(c, "cursor")
		limit, _ := helper.ParseQueryInt64(c, "limit")
		name := c.Query("name")

		filter := dto.DivisionFilterDto{
			DepartmentID: departmentId,
			SopId:        sopId,
			Preload:      c.Query("preload", "false") == "true",
			Sort:         sort,
			Order:        order,
			Cursor:       cursor,
			Limit:        limit,
			Name:         name,
			ShowDeleted:  c.Query("show_deleted", "false") == "true",
		}

		data, total, err := cn.DivisionHandler.GetAllDivisionsHandler(filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data, total)
	}
}

// GetDivisionByID godoc
//
//	@Summary		Get division by ID
//	@Description	Get a specific division by its ID
//	@Tags			divisions
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int64	true	"Division ID"
//	@Param			preload	query	bool	false	"Preload relations"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/divisions/{id} [get]
func GetDivisionByID(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filter := dto.DivisionFilterDto{
			Preload: c.Query("preload", "false") == "true",
		}
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		data, err := cn.DivisionHandler.GetDivisionByIDHandler(id, filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}

// CreateDivisions godoc
//
//	@Summary		Create a new division
//	@Description	Create a new division entry
//	@Tags			divisions
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.CreateDivisionDto	true	"Division data"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/divisions [post]
func CreateDivisions(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.CreateDivisionDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request")
		}

		result, err := cn.DivisionHandler.CreateDivisionHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, result)
	}
}

// UpdateDivisions godoc
//
//	@Summary		Update an existing division
//	@Description	Update division by ID
//	@Tags			divisions
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int64					true	"Division ID"
//	@Param			request	body	dto.UpdateDivisionDto	true	"Updated division data"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/divisions/{id} [put]
func UpdateDivisions(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		var input dto.UpdateDivisionDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid input")
		}

		updated, err := cn.DivisionHandler.UpdateDivisionHandler(id, &input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, updated)
	}
}

// DeleteDivisions godoc
//
//	@Summary		Delete a division
//	@Description	Delete division by ID
//	@Tags			divisions
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int64	true	"Division ID"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/divisions/{id} [delete]
func DeleteDivisions(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		if err := cn.DivisionHandler.DeleteDivisionHandler(id); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponseWithMessage(c, "Division deleted successfully", nil)
	}
}

// BulkCreateDivisions godoc
//
//	@Summary		Bulk create divisions
//	@Description	Create multiple divisions at once
//	@Tags			divisions
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.BulkCreateDivisionDto	true	"Bulk division data"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/divisions/bulk-create [post]
func BulkCreateDivisions(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.BulkCreateDivisionDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request body for bulk create")
		}
		if len(input.Data) == 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No data provided")
		}

		createdDivisions, err := cn.DivisionHandler.BulkCreateHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, createdDivisions)
	}
}

// BulkUpdateDivisions godoc
//
//	@Summary		Bulk update divisions
//	@Description	Update multiple divisions at once
//	@Tags			divisions
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.BulkUpdateDivisionDto	true	"Bulk update data"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/divisions/bulk-update [put]
func BulkUpdateDivisions(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.BulkUpdateDivisionDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request body for bulk update")
		}
		if len(input.IDs) == 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No Division IDs provided")
		}
		if input.Data == nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No update data provided")
		}

		updatedDivisions, err := cn.DivisionHandler.BulkUpdateHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, updatedDivisions)
	}
}

// BulkDeleteDivisions godoc
//
//	@Summary		Bulk delete divisions
//	@Description	Delete multiple divisions at once
//	@Tags			divisions
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.BulkDeleteDivisionDto	true	"Bulk delete data"
//	@Security		ApiKeyAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/divisions/bulk-delete [delete]
func BulkDeleteDivisions(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.BulkDeleteDivisionDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request body for bulk delete")
		}

		if len(input.IDs) == 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No Division IDs provided")
		}

		err := cn.DivisionHandler.BulkDeleteHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		return presenters.SendSuccessResponseWithMessage(c, fmt.Sprintf("Successfully deleted %d Divisions", len(input.IDs)), nil)
	}
}
