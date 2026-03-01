package controllers

import (
	"fmt"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/api/http/controllers/v1/handlers"
	"jk-api/api/http/presenters"

	"github.com/gofiber/fiber/v2"
)

func GetGraphById() fiber.Handler {
	return func(c *fiber.Ctx) error {
		elementId := c.Params("id")
		filter := dto.GraphFilterDto{
			DocumentId: c.Query("documentId", ""),
			ProductId:  int64(c.QueryInt("productId", 0)),
			ProjectId:  int64(c.QueryInt("projectId", 0)),
			EpicId:     int64(c.QueryInt("epicId", 0)),
			FeatureId:  int64(c.QueryInt("featureId", 0)),
			FilterId:   string(c.Query("filterId", "")),
		}
		data, err := handlers.GetGraphByIdHandler(elementId, filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}

func GetSOPGraphs() fiber.Handler {
	return func(c *fiber.Ctx) error {
		data, err := handlers.GetSOPGraphHandler()
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}

func GetGraphByLabel() fiber.Handler {
	return func(c *fiber.Ctx) error {
		label := c.Params("label")
		props := dto.GraphFilterDto{
			CommentId: c.Query("commentId", ""),
			ProductId: int64(c.QueryInt("productId", 0)),
			ProjectId: int64(c.QueryInt("projectId", 0)),
			EpicId:    int64(c.QueryInt("epicId", 0)),
			FeatureId: int64(c.QueryInt("featureId", 0)),
			Type:      c.Query("type", ""),
			Name:      c.Query("name", ""),
		}
		data, error := handlers.GetGraphByLabelHandler(label, props)

		if error != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, error)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}

func GetGraphByProps() fiber.Handler {
	return func(c *fiber.Ctx) error {
		props := dto.GraphFilterDto{
			CommentId: c.Query("commentId", ""),
			ProductId: int64(c.QueryInt("productId", 0)),
			ProjectId: int64(c.QueryInt("project_id", 0)),
			EpicId:    int64(c.QueryInt("epicId", 0)),
			FeatureId: int64(c.QueryInt("featureId", 0)),
			Type:      c.Query("type", ""),
		}

		data, err := handlers.GetGraphByPropsHandler(props)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}

func CreateGraph() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var payload dto.GraphNode
		if err := c.BodyParser(&payload); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		record, err := handlers.CreateGraphHandler(payload)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		return presenters.SendSuccessResponseWithMessage(c, "Graph created successfully", record)
	}
}

func BulkCreateGraph() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var payload []dto.GraphNode
		if err := c.BodyParser(&payload); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		error := handlers.BulkCreateGraphHandler(payload)
		if error != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, error)
		}

		return presenters.SendSuccessResponseWithMessage(c, "Graph created successfully", payload)
	}
}

func UpdateGraph() fiber.Handler {
	return func(c *fiber.Ctx) error {
		elementId := c.Params("id")
		var payload dto.NodeData
		if err := c.BodyParser(&payload); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusBadRequest, err)
		}

		if err := handlers.UpdateGraphHandler(elementId, payload); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponseWithMessage(c, "Graph updated successfully", nil)
	}
}

func UpdateMultipleGraph() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var payload []dto.NodeData
		if err := c.BodyParser(&payload); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusBadRequest, err)
		}

		fmt.Println("PAYLOAD UPDATE GRAPH", payload)
		if err := handlers.UpdateMultipleGraphHandler(payload); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponseWithMessage(c, "Graph updated successfully", nil)
	}
}

func MergeGraph() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var payload dto.GraphNode
		if err := c.BodyParser(&payload); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusBadRequest, err)
		}

		if err := handlers.MergeGraphHandler(payload); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponseWithMessage(c, "Graph updated successfully", nil)
	}
}

func GetDocumentGraph() fiber.Handler {
	return func(c *fiber.Ctx) error {
		documentId := c.Params("id")
		filter := dto.GraphFilterDto{
			ProjectId:   int64(c.QueryInt("projectId", 0)),
			ProductId:   int64(c.QueryInt("productId", 0)),
			EpicId:      int64(c.QueryInt("epicId", 0)),
			FeatureId:   int64(c.QueryInt("featureId", 0)),
			DocumentId:  c.Query("documentId", ""),
			ProductName: c.Query("product_name", ""),
			Multiple:    c.Query("multiple", "false") == "true",
		}

		data, err := handlers.GetDocumentGraphHandler(documentId, filter)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}

func GetCommentGraph() fiber.Handler {
	return func(c *fiber.Ctx) error {
		commentId := c.Params("elementId")
		data, err := handlers.GetCommentGraphHandler(commentId)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}

func CreateTableGraph() fiber.Handler {
	return func(c *fiber.Ctx) error {
		elementId := c.Params("elementId")
		relation := c.Query("relation")

		var payload dto.ColumnDto
		if err := c.BodyParser(&payload); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusBadRequest, err)
		}

		if err := handlers.CreateTableGraphHandler(elementId, relation, payload); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		return presenters.SendSuccessResponseWithMessage(c, "Graph table created successfully", nil)
	}
}

func CreateTextGraph() fiber.Handler {
	return func(c *fiber.Ctx) error {
		elementId := c.Params("elementId")

		var payload dto.TextDto
		if err := c.BodyParser(&payload); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusBadRequest, err)
		}

		if err := handlers.CreateTextGraphHandler(elementId, payload); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		return presenters.SendSuccessResponseWithMessage(c, "Graph text created successfully", nil)
	}
}

func CreateCommentGraph() fiber.Handler {
	return func(c *fiber.Ctx) error {
		elementId := c.Params("elementId")

		var comment dto.CommentDto
		if err := c.BodyParser(&comment); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusBadRequest, err)
		}

		if err := handlers.CreateCommentGraphHandler(elementId, comment); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		return presenters.SendSuccessResponseWithMessage(c, "Graph comment created successfully", nil)
	}
}

func CreateReviewGraph() fiber.Handler {
	return func(c *fiber.Ctx) error {
		elementId := c.Params("elementId")

		var review dto.ReviewDto
		if err := c.BodyParser(&review); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusBadRequest, err)
		}

		if err := handlers.CreateReviewGraphHandler(elementId, review); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		return presenters.SendSuccessResponseWithMessage(c, "Graph review created successfully", nil)
	}
}

func UpdateTableGraph() fiber.Handler {
	return func(c *fiber.Ctx) error {
		elementId := c.Params("elementId")
		relation := c.Query("relation")
		var payload dto.ColumnDto
		if err := c.BodyParser(&payload); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusBadRequest, err)
		}

		if err := handlers.UpdateTableGraphHandler(elementId, relation, payload); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		return presenters.SendSuccessResponseWithMessage(c, "Graph table updated successfully", nil)
	}
}

func DeleteGraph() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var payload dto.DeleteDto
		if err := c.BodyParser(&payload); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusBadRequest, err)
		}

		if len(payload.ElementIds) == 0 {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "elementIds cannot be empty")
		}

		if err := handlers.DeleteGraphHandler(payload.ElementIds); err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		return presenters.SendSuccessResponseWithMessage(c, "Graph nodes deleted successfully", nil)
	}
}
