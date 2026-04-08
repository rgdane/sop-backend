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

// GetSpkJobs godoc
// @Summary Get all SPK Jobs (Hybrid)
// @Description Get all SPK Jobs from both SQL and Graph database
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param spk_id query int64 false "SPK ID"
// @Param sop_id query int64 false "SOP ID"
// @Param title_id query int64 false "Title ID"
// @Param name query string false "Filter by name"
// @Param show_deleted query bool false "Show deleted records"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs [get]
func GetSpkJobs(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		spkID, _ := helper.ParseQueryInt64(c, "spk_id")
		sopID, _ := helper.ParseQueryInt64(c, "sop_id")
		titleID, _ := helper.ParseQueryInt64(c, "title_id")
		name := c.Query("name", "")

		filter := dto.SpkJobFilterDto{
			Preload:     c.Query("preload", "false") == "true",
			SpkID:       spkID,
			SopID:       sopID,
			TitleID:     titleID,
			Name:        name,
			ShowDeleted: c.Query("show_deleted", "false") == "true",
		}

		data, total, err := cn.SpkJobHandler.GetAllSpkJobsHandler(filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data, total)
	}
}

// GetGraphSpkJobs godoc
// @Summary Get all SPK Jobs (Graph only)
// @Description Get all SPK Jobs from Graph database only
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param spk_id query int64 false "SPK ID"
// @Param sop_id query int64 false "SOP ID"
// @Param title_id query int64 false "Title ID"
// @Param name query string false "Filter by name"
// @Param show_deleted query bool false "Show deleted records"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs/graph/ [get]
func GetGraphSpkJobs(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		spkID, _ := helper.ParseQueryInt64(c, "spk_id")
		sopID, _ := helper.ParseQueryInt64(c, "sop_id")
		titleID, _ := helper.ParseQueryInt64(c, "title_id")
		name := c.Query("name", "")
		deleted := c.Query("show_deleted", "false") == "true"

		filter := dto.SpkJobFilterDto{
			Preload:     c.Query("preload", "false") == "true",
			SpkID:       spkID,
			SopID:       sopID,
			TitleID:     titleID,
			Name:        name,
			ShowDeleted: deleted,
		}

		data, total, err := cn.SpkJobHandler.GetAllSpkJobsGraphHandler(filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data, total)
	}
}

// GetSqlSpkJobs godoc
// @Summary Get all SPK Jobs (SQL only)
// @Description Get all SPK Jobs from SQL database only
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param spk_id query int64 false "SPK ID"
// @Param sop_id query int64 false "SOP ID"
// @Param title_id query int64 false "Title ID"
// @Param name query string false "Filter by name"
// @Param show_deleted query bool false "Show deleted records"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs/sql/ [get]
func GetSqlSpkJobs(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		spkID, _ := helper.ParseQueryInt64(c, "spk_id")
		sopID, _ := helper.ParseQueryInt64(c, "sop_id")
		titleID, _ := helper.ParseQueryInt64(c, "title_id")
		name := c.Query("name", "")
		deleted := c.Query("show_deleted", "false") == "true"

		filter := dto.SpkJobFilterDto{
			Preload:     c.Query("preload", "false") == "true",
			SpkID:       spkID,
			SopID:       sopID,
			TitleID:     titleID,
			Name:        name,
			ShowDeleted: deleted,
		}

		data, total, err := cn.SpkJobHandler.GetAllSpkJobsHandler(filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data, total)
	}
}

// GetSpkJobByID godoc
// @Summary Get SPK Job by ID (Hybrid)
// @Description Get a single SPK Job by ID from both SQL and Graph database
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param id path int true "SPK Job ID"
// @Param spk_id query int64 false "SPK ID"
// @Param preload query bool false "Preload associations"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs/{id} [get]
func GetSpkJobByID(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		spkID, _ := helper.ParseQueryInt64(c, "spk_id")

		filter := dto.SpkJobFilterDto{
			Preload: c.Query("preload", "false") == "true",
			SpkID:   spkID,
		}

		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		data, err := cn.SpkJobHandler.GetSpkJobByIDHandler(id, filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}

// GetSqlSpkJobByID godoc
// @Summary Get SPK Job by ID (SQL only)
// @Description Get a single SPK Job by ID from SQL database only
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param id path int true "SPK Job ID"
// @Param spk_id query int64 false "SPK ID"
// @Param preload query bool false "Preload associations"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs/sql/{id} [get]
func GetSqlSpkJobByID(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		spkID, _ := helper.ParseQueryInt64(c, "spk_id")

		filter := dto.SpkJobFilterDto{
			Preload: c.Query("preload", "false") == "true",
			SpkID:   spkID,
		}

		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		data, err := cn.SpkJobHandler.GetSpkJobByIDHandler(id, filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}

// GetGraphSpkJobByID godoc
// @Summary Get SPK Job by ID (Graph only)
// @Description Get a single SPK Job by ID from Graph database only
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param id path int true "SPK Job ID"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs/graph/{id} [get]
func GetGraphSpkJobByID(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		data, err := cn.SpkJobHandler.GetSpkJobByIdGraphHandler(id)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}

// CreateSpkJob godoc
// @Summary Create SPK Job (Hybrid)
// @Description Create a new SPK Job in both SQL and Graph database
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param request body dto.CreateSpkJobDto true "SPK Job data"
// @Success 201 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs [post]
func CreateSpkJob(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.CreateSpkJobDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request")
		}

		result, err := cn.SpkJobHandler.CreateSpkJobHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, result)
	}
}

// CreateSqlSpkJob godoc
// @Summary Create SPK Job (SQL only)
// @Description Create a new SPK Job in SQL database only
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param request body dto.CreateSpkJobDto true "SPK Job data"
// @Success 201 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs/sql/ [post]
func CreateSqlSpkJob(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.CreateSpkJobDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request")
		}

		result, err := cn.SpkJobHandler.CreateSpkJobSqlHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, result)
	}
}

// CreateGraphSpkJob godoc
// @Summary Create SPK Job (Graph only)
// @Description Create a new SPK Job in Graph database only
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param request body dto.CreateSpkJobDto true "SPK Job data"
// @Success 201 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs/graph/ [post]
func CreateGraphSpkJob(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.CreateSpkJobDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request")
		}

		result, err := cn.SpkJobHandler.CreateSpkJobGraphHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, result)
	}
}

// UpdateSpkJobs godoc
// @Summary Update SPK Job (Hybrid)
// @Description Update an existing SPK Job in both SQL and Graph database
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param id path int true "SPK Job ID"
// @Param request body dto.UpdateSpkJobDto true "SPK Job update data"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs/{id} [put]
func UpdateSpkJobs(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		var input dto.UpdateSpkJobDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid input")
		}

		updated, err := cn.SpkJobHandler.UpdateSpkJobHandler(id, &input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, updated)
	}
}

// UpdateSqlSpkJobs godoc
// @Summary Update SPK Job (SQL only)
// @Description Update an existing SPK Job in SQL database only
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param id path int true "SPK Job ID"
// @Param request body dto.UpdateSpkJobDto true "SPK Job update data"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs/sql/{id} [put]
func UpdateSqlSpkJobs(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		var input dto.UpdateSpkJobDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid input")
		}

		updated, err := cn.SpkJobHandler.UpdateSpkJobSqlHandler(id, &input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, updated)
	}
}

// UpdateGraphSpkJobs godoc
// @Summary Update SPK Job (Graph only)
// @Description Update an existing SPK Job in Graph database only
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param id path int true "SPK Job ID"
// @Param request body dto.UpdateSpkJobDto true "SPK Job update data"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs/graph/{id} [put]
func UpdateGraphSpkJobs(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		var input dto.UpdateSpkJobDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid input")
		}

		updated, err := cn.SpkJobHandler.UpdateSpkJobGraphHandler(id, &input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, updated)
	}
}

// DeleteSpkJob godoc
// @Summary Delete SPK Job (Hybrid)
// @Description Delete a SPK Job from both SQL and Graph database
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param id path int true "SPK Job ID"
// @Param isPermanent query bool false "Permanent delete (default: false)"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs/{id} [delete]
func DeleteSpkJob(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		isPermanent := c.Query("isPermanent", "false") == "true"
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		if err := cn.SpkJobHandler.DeleteSpkJobHandler(id, isPermanent); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponseWithMessage(c, "Spk Job berhasil dihapus", nil)
	}
}

// DeleteSqlSpkJob godoc
// @Summary Delete SPK Job (SQL only)
// @Description Delete a SPK Job from SQL database only
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param id path int true "SPK Job ID"
// @Param isPermanent query bool false "Permanent delete (default: false)"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs/sql/{id} [delete]
func DeleteSqlSpkJob(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		isPermanent := c.Query("isPermanent", "false") == "true"
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		if err := cn.SpkJobHandler.DeleteSpkJobSqlHandler(id, isPermanent); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponseWithMessage(c, "Spk Job berhasil dihapus", nil)
	}
}

// DeleteGraphSpkJob godoc
// @Summary Delete SPK Job (Graph only)
// @Description Delete a SPK Job from Graph database only
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param id path int true "SPK Job ID"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs/graph/{id} [delete]
func DeleteGraphSpkJob(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		if err := cn.SpkJobHandler.DeleteSpkJobGraphHandler(id); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponseWithMessage(c, "Spk Job deleted successfully from graph", nil)
	}
}

// BulkCreateSpkJobs godoc
// @Summary Bulk create SPK Jobs (Hybrid)
// @Description Create multiple SPK Jobs in both SQL and Graph database
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param request body dto.BulkCreateSpkJobsDto true "Bulk SPK Job data"
// @Success 201 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs/bulk-create [post]
func BulkCreateSpkJobs(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.BulkCreateSpkJobsDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request body for bulk create")
		}
		if len(input.Data) == 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No data provided")
		}

		createdSpkJobs, err := cn.SpkJobHandler.BulkCreateSpkJobsHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, createdSpkJobs)
	}
}

// BulkUpdateSpkJobs godoc
// @Summary Bulk update SPK Jobs (Hybrid)
// @Description Update multiple SPK Jobs in both SQL and Graph database
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param request body dto.BulkUpdateSpkJobDto true "Bulk update data"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs/bulk-update [put]
func BulkUpdateSpkJobs(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.BulkUpdateSpkJobDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request body for bulk update")
		}
		if len(input.IDs) == 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No level IDs provided")
		}
		if input.Data == nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No update data provided")
		}

		updatedSpkJobs, err := cn.SpkJobHandler.BulkUpdateSpkJobsHandler(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, updatedSpkJobs)
	}
}

// BulkDeleteSpkJobs godoc
// @Summary Bulk delete SPK Jobs (Hybrid)
// @Description Delete multiple SPK Jobs from both SQL and Graph database
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param isPermanent query bool false "Permanent delete (default: false)"
// @Param request body dto.BulkDeleteSpkJobDto true "Bulk delete data"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs/bulk-delete [delete]
func BulkDeleteSpkJobs(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.BulkDeleteSpkJobDto
		isPermanent := c.Query("isPermanent", "false") == "true"
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request body for bulk delete")
		}

		if len(input.IDs) == 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "No spk_job IDs provided")
		}

		err := cn.SpkJobHandler.BulkDeleteSpkJobsHandler(&input, isPermanent)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		return presenters.SendSuccessResponseWithMessage(c, fmt.Sprintf("Successfully deleted %d spk_jobs", len(input.IDs)), nil)
	}
}

// ReorderSpkJob godoc
// @Summary Reorder SPK Job
// @Description Reorder SPK Job index within an SPK
// @Tags SPK-Job
// @Accept json
// @Produce json
// @Param id path int true "SPK Job ID"
// @Param request body dto.ReorderSpkJobDto true "Reorder data"
// @Success 200 {object} presenters.SuccessResponse
// @Failure 400 {object} presenters.ErrorResponse
// @Failure 500 {object} presenters.ErrorResponse
// @Router /spk-jobs/{id}/reorder [post]
func ReorderSpkJob(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid ID")
		}

		var input dto.ReorderSpkJobDto
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid input")
		}

		err = cn.SpkJobHandler.ReorderSpkJobHandler(id, &input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		return presenters.SendSuccessResponseWithMessage(c, "Status reordered successfully", nil)
	}
}
