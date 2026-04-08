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

// GetSpks godoc
// @Summary Get all SPKs (Hybrid)
// @Description Get all SPKs from both SQL and Graph database
// @Tags SPK
// @Accept json
// @Produce json
// @Param title_id query int64 false "Title ID"
// @Param cursor query int64 false "Cursor for pagination"
// @Param limit query int64 false "Limit for pagination"
// @Param name query string false "Filter by name"
// @Param show_deleted query bool false "Show deleted records"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spks [get]
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

// GetGraphSpks godoc
// @Summary Get all SPKs (Graph only)
// @Description Get all SPKs from Graph database only
// @Tags SPK
// @Accept json
// @Produce json
// @Param title_id query int64 false "Title ID"
// @Param cursor query int64 false "Cursor for pagination"
// @Param limit query int64 false "Limit for pagination"
// @Param name query string false "Filter by name"
// @Param show_deleted query bool false "Show deleted records"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spks/graph/ [get]
func GetGraphSpks(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		titleId, _ := helper.ParseQueryInt64(c, "title_id")
		cursor, _ := helper.ParseQueryInt64(c, "cursor")
		limit, _ := helper.ParseQueryInt64(c, "limit")
		name := c.Query("name", "")
		deleted := c.Query("show_deleted", "false") == "true"

		filter := dto.SpkFilterDto{
			TitleIDs:    titleId,
			Preload:     c.Query("preload", "false") == "true",
			Cursor:      cursor,
			Limit:       limit,
			Name:        name,
			ShowDeleted: deleted,
		}

		data, total, err := cn.SpkHandler.GetAllSpksGraphHandler(filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data, total)
	}
}

// GetSqlSpks godoc
// @Summary Get all SPKs (SQL only)
// @Description Get all SPKs from SQL database only
// @Tags SPK
// @Accept json
// @Produce json
// @Param title_id query int64 false "Title ID"
// @Param cursor query int64 false "Cursor for pagination"
// @Param limit query int64 false "Limit for pagination"
// @Param name query string false "Filter by name"
// @Param show_deleted query bool false "Show deleted records"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spks/sql/ [get]
func GetSqlSpks(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		titleId, _ := helper.ParseQueryInt64(c, "title_id")
		cursor, _ := helper.ParseQueryInt64(c, "cursor")
		limit, _ := helper.ParseQueryInt64(c, "limit")
		name := c.Query("name", "")
		deleted := c.Query("show_deleted", "false") == "true"

		filter := dto.SpkFilterDto{
			TitleIDs:    titleId,
			Preload:     c.Query("preload", "false") == "true",
			Cursor:      cursor,
			Limit:       limit,
			Name:        name,
			ShowDeleted: deleted,
		}

		data, total, err := cn.SpkHandler.GetAllSpksHandler(filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data, total)
	}
}

// GetSpkByID godoc
// @Summary Get SPK by ID (Hybrid)
// @Description Get a single SPK by ID from both SQL and Graph database
// @Tags SPK
// @Accept json
// @Produce json
// @Param id path int true "SPK ID"
// @Param preload query bool false "Preload associations"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spks/{id} [get]
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

// GetGraphSpkByID godoc
// @Summary Get SPK by ID (Graph only)
// @Description Get a single SPK by ID from Graph database only
// @Tags SPK
// @Accept json
// @Produce json
// @Param id path int true "SPK ID"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spks/graph/{id} [get]
func GetGraphSpkByID(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		data, err := cn.SpkHandler.GetSpkByIdGraphHandler(id)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}

// CreateSpk godoc
// @Summary Create SPK (Hybrid)
// @Description Create a new SPK in both SQL and Graph database
// @Tags SPK
// @Accept json
// @Produce json
// @Param request body dto.CreateSpkDto true "SPK data"
// @Success 201 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spks [post]
func CreateSpk(cn *container.AppContainer) fiber.Handler {
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

// CreateSqlSpk godoc
// @Summary Create SPK (SQL only)
// @Description Create a new SPK in SQL database only
// @Tags SPK
// @Accept json
// @Produce json
// @Param request body dto.CreateSpkDto true "SPK data"
// @Success 201 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spks/sql/ [post]
func CreateSqlSpk(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.CreateSpkDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request")
		}

		result, err := cn.SpkHandler.CreateSpkSqlHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, result)
	}
}

// CreateGraphSpk godoc
// @Summary Create SPK (Graph only)
// @Description Create a new SPK in Graph database only
// @Tags SPK
// @Accept json
// @Produce json
// @Param request body dto.CreateSpkDto true "SPK data"
// @Success 201 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spks/graph/ [post]
func CreateGraphSpk(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.CreateSpkDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request")
		}

		result, err := cn.SpkHandler.CreateSpkGraphHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, result)
	}
}

// UpdateSpk godoc
// @Summary Update SPK (Hybrid)
// @Description Update an existing SPK in both SQL and Graph database
// @Tags SPK
// @Accept json
// @Produce json
// @Param id path int true "SPK ID"
// @Param request body dto.UpdateSpkDto true "SPK update data"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spks/{id} [put]
func UpdateSpk(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.UpdateSpkDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid input")
		}

		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		result, err := cn.SpkHandler.UpdateSpkHandler(id, &input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, result)
	}
}

// UpdateSqlSpk godoc
// @Summary Update SPK (SQL only)
// @Description Update an existing SPK in SQL database only
// @Tags SPK
// @Accept json
// @Produce json
// @Param id path int true "SPK ID"
// @Param request body dto.UpdateSpkDto true "SPK update data"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spks/sql/{id} [put]
func UpdateSqlSpk(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.UpdateSpkDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid input")
		}

		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		result, err := cn.SpkHandler.UpdateSpkSqlHandler(id, &input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, result)
	}
}

// UpdateGraphSpk godoc
// @Summary Update SPK (Graph only)
// @Description Update an existing SPK in Graph database only
// @Tags SPK
// @Accept json
// @Produce json
// @Param id path int true "SPK ID"
// @Param request body dto.UpdateSpkDto true "SPK update data"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spks/graph/{id} [put]
func UpdateGraphSpk(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.UpdateSpkDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid input")
		}

		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		result, err := cn.SpkHandler.UpdateSpkGraphHandler(id, &input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, result)
	}
}

// DeleteSpk godoc
// @Summary Delete SPK (Hybrid)
// @Description Delete an SPK from both SQL and Graph database
// @Tags SPK
// @Accept json
// @Produce json
// @Param id path int true "SPK ID"
// @Param isPermanent query bool false "Permanent delete (default: false)"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spks/{id} [delete]
func DeleteSpk(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		isPermanent := c.Query("isPermanent", "false") == "true"
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		err = cn.SpkHandler.DeleteSpkHandler(id, isPermanent)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, nil)
	}
}

// DeleteSqlSpk godoc
// @Summary Delete SPK (SQL only)
// @Description Delete an SPK from SQL database only
// @Tags SPK
// @Accept json
// @Produce json
// @Param id path int true "SPK ID"
// @Param isPermanent query bool false "Permanent delete (default: false)"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spks/sql/{id} [delete]
func DeleteSqlSpk(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		isPermanent := c.Query("isPermanent", "false") == "true"
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		err = cn.SpkHandler.DeleteSpkSqlHandler(id, isPermanent)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, nil)
	}
}

// DeleteGraphSpk godoc
// @Summary Delete SPK (Graph only)
// @Description Delete an SPK from Graph database only
// @Tags SPK
// @Accept json
// @Produce json
// @Param id path int true "SPK ID"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spks/graph/{id} [delete]
func DeleteGraphSpk(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		err = cn.SpkHandler.DeleteSpkGraphHandler(id)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponseWithMessage(c, "SPK deleted successfully from graph", nil)
	}
}

// BulkCreateSpks godoc
// @Summary Bulk create SPKs (Hybrid)
// @Description Create multiple SPKs in both SQL and Graph database
// @Tags SPK
// @Accept json
// @Produce json
// @Param request body dto.BulkCreateSpksDto true "Bulk SPK data"
// @Success 201 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spks/bulk-create [post]
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

// BulkUpdateSpks godoc
// @Summary Bulk update SPKs (Hybrid)
// @Description Update multiple SPKs in both SQL and Graph database
// @Tags SPK
// @Accept json
// @Produce json
// @Param request body dto.BulkUpdateSpkDto true "Bulk update data"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spks/bulk-update [put]
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

// BulkDeleteSpks godoc
// @Summary Bulk delete SPKs (Hybrid)
// @Description Delete multiple SPKs from both SQL and Graph database
// @Tags SPK
// @Accept json
// @Produce json
// @Param isPermanent query bool false "Permanent delete (default: false)"
// @Param request body dto.BulkDeleteSpkDto true "Bulk delete data"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spks/bulk-delete [delete]
func BulkDeleteSpks(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.BulkDeleteSpkDto
		isPermanent := c.Query("isPermanent", "false") == "true"
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request body for bulk delete")
		}

		if len(input.IDs) == 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No Spk IDs provided")
		}

		err := cn.SpkHandler.BulkDeleteSpksHandler(&input, isPermanent)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		return presenters.SendSuccessResponseWithMessage(c, fmt.Sprintf("Successfully deleted %d Spks", len(input.IDs)), nil)
	}
}
