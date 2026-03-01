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

// GetDepartments godoc
//
//	@Summary		Get all departments
//	@Description	Get list of departments with optional filters
//	@Tags			departments
//	@Accept			json
//	@Produce		json
//	@Param			sort			query		string	false	"Sort field"
//	@Param			order			query		string	false	"Sort order"
//	@Param			cursor			query		int64	false	"Cursor for pagination"
//	@Param			limit			query		int64	false	"Limit for pagination"
//	@Param			name			query		string	false	"Name filter"
//	@Param			show_deleted	query		bool	false	"Show deleted departments"
//	@Param			preload			query		bool	false	"Preload relations"
//	@Success		200				{object}	presenters.SuccessResponse
//	@Failure		500				{object}	presenters.ErrorResponse
//	@Router			/departments [get]
func GetDepartments(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sort := c.Query("sort")
		order := c.Query("order")
		cursor, _ := helper.ParseQueryInt64(c, "cursor")
		limit, _ := helper.ParseQueryInt64(c, "limit")
		name := c.Query("name")

		filter := dto.DepartmentFilterDto{
			Preload:     c.Query("preload", "false") == "true",
			Sort:        sort,
			Order:       order,
			Limit:       limit,
			Cursor:      cursor,
			Name:        name,
			ShowDeleted: c.Query("show_deleted", "false") == "true",
		}

		data, total, err := cn.DepartmentHandler.GetAllDepartmentsHandler(filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data, total)
	}
}

// GetDepartmentByID godoc
//
//	@Summary		Get department by ID
//	@Description	Get a specific department by its ID
//	@Tags			departments
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int64	true	"Department ID"
//	@Param			preload	query		bool	false	"Preload relations"
//	@Success		200		{object}	presenters.SuccessResponse
//	@Failure		400		{object}	presenters.ErrorResponse
//	@Failure		500		{object}	presenters.ErrorResponse
//	@Router			/departments/{id} [get]
func GetDepartmentByID(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filter := dto.DepartmentFilterDto{
			Preload: c.Query("preload", "false") == "true",
		}

		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		data, err := cn.DepartmentHandler.GetDepartmentByIDHandler(id, filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}

// CreateDepartments godoc
//
//	@Summary		Create a new department
//	@Description	Create a new department entry
//	@Tags			departments
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.CreateDepartmentDto	true	"Department data"
//	@example		request
//
//	{
//	  "name": "Engineering",
//	  "code": "ENG"
//	}
//
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/departments [post]
func CreateDepartments(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.CreateDepartmentDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request")
		}

		result, err := cn.DepartmentHandler.CreateDepartmentHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, result)
	}
}

// UpdateDepartments godoc
//
//	@Summary		Update an existing department
//	@Description	Update department by ID
//	@Tags			departments
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int64					true	"Department ID"
//	@Param			request	body	dto.UpdateDepartmentDto	true	"Updated department data"
//	@example		request
//
//	{
//	  "name": "Updated Engineering",
//	  "code": "ENG_UPDATED"
//	}
//
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/departments/{id} [put]
func UpdateDepartments(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		var input dto.UpdateDepartmentDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid input")
		}

		updated, err := cn.DepartmentHandler.UpdateDepartmentHandler(id, &input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, updated)
	}
}

// DeleteDepartments godoc
//
//	@Summary		Delete a department
//	@Description	Delete department by ID
//	@Tags			departments
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int64	true	"Department ID"
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/departments/{id} [delete]
func DeleteDepartments(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		if err := cn.DepartmentHandler.DeleteDepartmentHandler(id); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponseWithMessage(c, "Department deleted successfully", nil)
	}
}

// BulkCreateDepartments godoc
//
//	@Summary		Bulk create departments
//	@Description	Create multiple departments at once
//	@Tags			departments
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.BulkCreateDepartments	true	"Bulk department data"
//	@example		request
//
//	{
//	  "data": [
//	    {
//	      "name": "Engineering",
//	      "code": "ENG"
//	    },
//	    {
//	      "name": "Marketing",
//	      "code": "MKT"
//	    }
//	  ]
//	}
//
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/departments/bulk-create [post]
func BulkCreateDepartments(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.BulkCreateDepartments
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request body for bulk create")
		}
		if len(input.Data) == 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No data provided")
		}

		createdDepartments, err := cn.DepartmentHandler.BulkCreateDepartmentsHandler(&input, c)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, createdDepartments)
	}
}

// BulkUpdateDepartments godoc
//
//	@Summary		Bulk update departments
//	@Description	Update multiple departments at once
//	@Tags			departments
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.BulkUpdateDepartmentDto	true	"Bulk update data"
//	@example		request
//
//	{
//	  "ids": [1, 2],
//	  "data": {
//	    "name": "Updated Department"
//	  }
//	}
//
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/departments/bulk-update [put]
func BulkUpdateDepartments(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.BulkUpdateDepartmentDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request body for bulk update")
		}
		if len(input.IDs) == 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No department IDs provided")
		}
		if input.Data == nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No update data provided")
		}

		updatedDepartments, err := cn.DepartmentHandler.BulkUpdateHandler(&input, c)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, updatedDepartments)
	}
}

// BulkDeleteDepartments godoc
//
//	@Summary		Bulk delete departments
//	@Description	Delete multiple departments at once
//	@Tags			departments
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.BulkDeleteDepartmentDto	true	"Bulk delete data"
//	@example		request
//
//	{
//	  "ids": [1, 2, 3]
//	}
//
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/departments/bulk-delete [delete]
func BulkDeleteDepartments(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.BulkDeleteDepartmentDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request body for bulk delete")
		}

		if len(input.IDs) == 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No department IDs provided")
		}

		err := cn.DepartmentHandler.BulkDeleteHandler(&input, c)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		return presenters.SendSuccessResponseWithMessage(c, fmt.Sprintf("Successfully deleted %d departments", len(input.IDs)), nil)
	}
}
