package controllers

import (
	"fmt"
	"strconv"

	"jk-api/api/http/controllers/v1/dto"
	"jk-api/api/http/presenters"
	"jk-api/internal/container"
	"jk-api/internal/shared/helper"

	"github.com/gofiber/fiber/v2"
)

// GetSops godoc
//
//	@Summary		Get all SOPs (Hybrid)
//	@Description	Get list of SOPs from both SQL and Graph database
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			title_id		query	int64	false	"Title ID filter"
//	@Param			division_id	query	int64	false	"Division ID filter"
//	@Param			cursor		query	int64	false	"Cursor for pagination"
//	@Param			limit		query	int64	false	"Limit for pagination"
//	@Param			name		query	string	false	"Name filter"
//	@Param			show_deleted	query	bool	false	"Show deleted SOPs"
//	@Param			preload		query	bool	false	"Preload relations"
//	@Security		BearerAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops [get]
func GetSops(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		titleId, _ := helper.ParseQueryInt64(c, "title_id")
		divisionId, _ := helper.ParseQueryInt64(c, "division_id")
		divisionIds, _ := helper.ParseQueryInt64Array(c, "division_ids")
		cursor, _ := helper.ParseQueryInt64(c, "cursor")
		limit, _ := helper.ParseQueryInt64(c, "limit")
		excludeId, _ := helper.ParseQueryInt64(c, "exclude_id")
		code := c.Query("code")
		name := c.Query("name")
		deleted := c.Query("show_deleted", "false") == "true"

		filter := dto.SopFilterDto{
			TitleID:     titleId,
			DivisionID:  divisionId,
			DivisionIDs: divisionIds,
			Preload:     c.Query("preload", "false") == "true",
			Cursor:      cursor,
			ShowDeleted: deleted,
			Limit:       limit,
			Code:        &code,
			Name:        name,
			ExcludeID:   excludeId,
		}

		data, total, err := cn.SopHandler.GetAllSopsHandler(filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data, total)
	}
}

// GetSqlSops godoc
//
//	@Summary		Get all SOPs (SQL only)
//	@Description	Get list of SOPs from SQL database only
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			title_id		query	int64	false	"Title ID filter"
//	@Param			division_id	query	int64	false	"Division ID filter"
//	@Param			cursor		query	int64	false	"Cursor for pagination"
//	@Param			limit		query	int64	false	"Limit for pagination"
//	@Param			name		query	string	false	"Name filter"
//	@Param			show_deleted	query	bool	false	"Show deleted SOPs"
//	@Param			preload		query	bool	false	"Preload relations"
//	@Security		BearerAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops/sql [get]
func GetSqlSops(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		titleId, _ := helper.ParseQueryInt64(c, "title_id")
		divisionId, _ := helper.ParseQueryInt64(c, "division_id")
		divisionIds, _ := helper.ParseQueryInt64Array(c, "division_ids")
		cursor, _ := helper.ParseQueryInt64(c, "cursor")
		limit, _ := helper.ParseQueryInt64(c, "limit")
		excludeId, _ := helper.ParseQueryInt64(c, "exclude_id")
		code := c.Query("code")
		name := c.Query("name")
		deleted := c.Query("show_deleted", "false") == "true"

		filter := dto.SopFilterDto{
			TitleID:     titleId,
			DivisionID:  divisionId,
			DivisionIDs: divisionIds,
			Preload:     c.Query("preload", "false") == "true",
			Cursor:      cursor,
			ShowDeleted: deleted,
			Limit:       limit,
			Code:        &code,
			Name:        name,
			ExcludeID:   excludeId,
		}

		data, total, err := cn.SopHandler.GetAllSopsHandler(filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data, total)
	}
}

// GetGraphSops godoc
//
//	@Summary		Get all SOPs (Graph only)
//	@Description	Get list of SOPs from Graph database only
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			title_id		query	int64	false	"Title ID filter"
//	@Param			cursor		query	int64	false	"Cursor for pagination"
//	@Param			limit		query	int64	false	"Limit for pagination"
//	@Param			name		query	string	false	"Name filter"
//	@Param			show_deleted	query	bool	false	"Show deleted SOPs"
//	@Security		BearerAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops/graph [get]
func GetGraphSops(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		titleId, _ := helper.ParseQueryInt64(c, "title_id")
		divisionId, _ := helper.ParseQueryInt64(c, "division_id")
		divisionIds, _ := helper.ParseQueryInt64Array(c, "division_ids")
		cursor, _ := helper.ParseQueryInt64(c, "cursor")
		limit, _ := helper.ParseQueryInt64(c, "limit")
		excludeId, _ := helper.ParseQueryInt64(c, "exclude_id")
		code := c.Query("code")
		name := c.Query("name")
		deleted := c.Query("show_deleted", "false") == "true"

		filter := dto.SopFilterDto{
			TitleID:     titleId,
			DivisionID:  divisionId,
			DivisionIDs: divisionIds,
			Preload:     c.Query("preload", "false") == "true",
			Cursor:      cursor,
			ShowDeleted: deleted,
			Limit:       limit,
			Code:        &code,
			Name:        name,
			ExcludeID:   excludeId,
		}

		data, total, err := cn.SopHandler.GetAllSopsGraphHandler(filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data, total)
	}
}

// GetSopByID godoc
//
//	@Summary		Get SOP by ID (Hybrid)
//	@Description	Get a specific SOP by its ID from both SQL and Graph database
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int64	true	"SOP ID"
//	@Param			preload	query	bool	false	"Preload relations"
//	@Security		BearerAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops/{id} [get]
func GetSopByID(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filter := dto.SopFilterDto{
			Preload: c.Query("preload", "false") == "true",
		}

		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		data, err := cn.SopHandler.GetSopByIDHandler(id, filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}

// GetSqlSopByID godoc
//
//	@Summary		Get SOP by ID (SQL only)
//	@Description	Get a specific SOP by its ID from SQL database only
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int64	true	"SOP ID"
//	@Param			preload	query	bool	false	"Preload relations"
//	@Security		BearerAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops/sql/{id} [get]
func GetSqlSopByID(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		filter := dto.SopFilterDto{
			Preload: c.Query("preload", "false") == "true",
		}

		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		data, err := cn.SopHandler.GetSopByIDHandler(id, filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}

// GetGraphSopByID godoc
//
//	@Summary		Get SOP by ID (Graph only)
//	@Description	Get a specific SOP by its ID from Graph database only
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int64	true	"SOP ID"
//	@Security		BearerAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops/graph/{id} [get]
func GetGraphSopByID(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		data, err := cn.SopHandler.GetSopByIdGraphHandler(id)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}

// CountSops godoc
//
//	@Summary		Count SOPs
//	@Description	Get count of SOPs with optional filters
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			title_id	query	int64	false	"Title ID filter"
//	@Security		BearerAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops/count [get]
func CountSops(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		titleId, _ := helper.ParseQueryInt64(c, "title_id")

		filter := dto.SopFilterDto{
			TitleID: titleId,
		}

		count, err := cn.SopHandler.CountSopsHandler(filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, count)
	}
}

// CreateSop godoc
//
//	@Summary		Create SOP (Hybrid)
//	@Description	Create a new SOP in both SQL and Graph database
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.CreateSopDto	true	"SOP data"
//	@Security		BearerAuth
//	@Success		201	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops [post]
func CreateSop(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.CreateSopDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request")
		}

		result, err := cn.SopHandler.CreateSopHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, result)
	}
}

// CreateSqlSop godoc
//
//	@Summary		Create SOP (SQL only)
//	@Description	Create a new SOP in SQL database only
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.CreateSopDto	true	"SOP data"
//	@Security		BearerAuth
//	@Success		201	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops/sql [post]
func CreateSqlSop(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.CreateSopDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request")
		}

		result, err := cn.SopHandler.CreateSopSqlHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, result)
	}
}

// CreateGraphSop godoc
//
//	@Summary		Create SOP (Graph only)
//	@Description	Create a new SOP in Graph database only
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.CreateSopDto	true	"SOP data"
//	@Security		BearerAuth
//	@Success		201	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops/graph [post]
func CreateGraphSop(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.CreateSopDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request")
		}

		result, err := cn.SopHandler.CreateSopGraphHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, result)
	}
}

// UpdateSop godoc
//
//	@Summary		Update SOP (Hybrid)
//	@Description	Update an existing SOP in both SQL and Graph database
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int64				true	"SOP ID"
//	@Param			request	body	dto.UpdateSopDto	true	"Updated SOP data"
//	@Security		BearerAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops/{id} [put]
func UpdateSop(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.UpdateSopDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid input")
		}

		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		result, err := cn.SopHandler.UpdateSopHandler(id, &input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, result)
	}
}

// UpdateSqlSop godoc
//
//	@Summary		Update SOP (SQL only)
//	@Description	Update an existing SOP in SQL database only
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int64				true	"SOP ID"
//	@Param			request	body	dto.UpdateSopDto	true	"Updated SOP data"
//	@Security		BearerAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops/sql/{id} [put]
func UpdateSqlSop(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.UpdateSopDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid input")
		}

		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		result, err := cn.SopHandler.UpdateSopSqlHandler(id, &input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, result)
	}
}

// UpdateGraphSop godoc
//
//	@Summary		Update SOP (Graph only)
//	@Description	Update an existing SOP in Graph database only
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int64				true	"SOP ID"
//	@Param			request	body	dto.UpdateSopDto	true	"Updated SOP data"
//	@Security		BearerAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops/graph/{id} [put]
func UpdateGraphSop(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.UpdateSopDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid input")
		}

		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		result, err := cn.SopHandler.UpdateSopGraphHandler(id, &input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, result)
	}
}

// DeleteSop godoc
//
//	@Summary		Delete SOP (Hybrid)
//	@Description	Delete SOP from both SQL and Graph database
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			id			path	int64	true	"SOP ID"
//	@Param			isPermanent	query	bool	false	"Permanent delete"
//	@Security		BearerAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops/{id} [delete]
func DeleteSop(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		isPermanent := c.Query("isPermanent", "false") == "true"
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		err = cn.SopHandler.DeleteSopHandler(id, isPermanent)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, nil)
	}
}

// DeleteSqlSop godoc
//
//	@Summary		Delete SOP (SQL only)
//	@Description	Delete SOP from SQL database only
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			id			path	int64	true	"SOP ID"
//	@Param			isPermanent	query	bool	false	"Permanent delete"
//	@Security		BearerAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops/sql/{id} [delete]
func DeleteSqlSop(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		isPermanent := c.Query("isPermanent", "false") == "true"
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		err = cn.SopHandler.DeleteSopSqlHandler(id, isPermanent)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, nil)
	}
}

// DeleteGraphSop godoc
//
//	@Summary		Delete SOP (Graph only)
//	@Description	Delete SOP from Graph database only
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int64	true	"SOP ID"
//	@Security		BearerAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops/graph/{id} [delete]
func DeleteGraphSop(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		err = cn.SopHandler.DeleteSopGraphHandler(id)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponseWithMessage(c, "SOP deleted successfully from graph", nil)
	}
}

// BulkCreateSops godoc
//
//	@Summary		Bulk create SOPs (Hybrid)
//	@Description	Create multiple SOPs in both SQL and Graph database
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.BulkCreateSopsDto	true	"Bulk SOP data"
//	@Security		BearerAuth
//	@Success		201	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops/bulk-create [post]
func BulkCreateSops(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.BulkCreateSopsDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request body for bulk create")
		}
		if len(input.Data) == 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No data provided")
		}

		createdSops, err := cn.SopHandler.BulkCreateSopsHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, createdSops)
	}
}

// BulkUpdateSops godoc
//
//	@Summary		Bulk update SOPs (Hybrid)
//	@Description	Update multiple SOPs in both SQL and Graph database
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.BulkUpdateSopDto	true	"Bulk update data"
//	@Security		BearerAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops/bulk-update [put]
func BulkUpdateSops(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.BulkUpdateSopDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request body for bulk update")
		}
		if len(input.IDs) == 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No SOP IDs provided")
		}
		if input.Data == nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No update data provided")
		}

		updatedSops, err := cn.SopHandler.BulkUpdateSopsHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, updatedSops)
	}
}

// BulkDeleteSops godoc
//
//	@Summary		Bulk delete SOPs (Hybrid)
//	@Description	Delete multiple SOPs from both SQL and Graph database
//	@Tags			SOP
//	@Accept			json
//	@Produce		json
//	@Param			request		body	dto.BulkDeleteSopDto	true	"Bulk delete data"
//	@Param			isPermanent	query	bool					false	"Permanent delete"
//	@Security		BearerAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/sops/bulk-delete [delete]
func BulkDeleteSops(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.BulkDeleteSopDto
		isPermanent := c.Query("isPermanent", "false") == "true"
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request body for bulk delete")
		}

		if len(input.IDs) == 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No SOP IDs provided")
		}

		err := cn.SopHandler.BulkDeleteSopsHandler(&input, isPermanent)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		return presenters.SendSuccessResponseWithMessage(c, fmt.Sprintf("Successfully deleted %d SOPs", len(input.IDs)), nil)
	}
}
