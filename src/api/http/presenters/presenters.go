package presenters

import "github.com/gofiber/fiber/v2"

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    interface{} `json:"data"`
	Cursor  *int64      `json:"cursor,omitempty"`
	Total   *int64      `json:"total,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"error message"`
}

// MessageResponse represents a response with a message
type MessageResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"operation successful"`
	Data    interface{} `json:"data,omitempty"`
}

func SendSuccessResponse(c *fiber.Ctx, data any, total ...int64) error {
	resp := fiber.Map{
		"success": true,
		"data":    data,
	}

	if len(total) > 0 && total[0] > 0 {
		resp["total"] = total[0]
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func SendSuccessFlatResponse(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(data)
}

func SendSuccessCreatedResponse(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

func SendErrorResponse(c *fiber.Ctx, status int, err error) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error":   err.Error(),
	})
}

func SendSuccessResponseWithMessage(c *fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func SendErrorResponseWithMessage(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"error":   message,
	})
}

func SendCursorSuccessResponse(c *fiber.Ctx, data any, cursor int64, total ...int64) error {
	resp := fiber.Map{
		"success": true,
		"data":    data,
		"cursor":  cursor,
	}

	if len(total) > 0 {
		resp["total"] = total[0]
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func SendCursorPaginationResponse(c *fiber.Ctx, data any, next int64, prev int64, total int64) error {
	resp := fiber.Map{
		"success": true,
		"data":    data,
		"next":    next,
		"prev":    prev,
		"total":   total,
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// func SuccessLogin(c *fiber.Ctx, data any, token string) error {
// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"success": true,
// 		"data":   data,
// 		"token": token,
// 	})
// }
