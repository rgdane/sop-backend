package controllers

import (
	"errors"
	"jk-api/api/http/controllers/v1/dto"
	"jk-api/api/http/controllers/v1/handlers"
	"jk-api/api/http/presenters"
	"jk-api/internal/container"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type AuthController struct {
	Handler *handlers.AuthHandler
}

func NewAuthController(h *handlers.AuthHandler) *AuthController {
	return &AuthController{Handler: h}
}

// GetProfile godoc
//
//	@Summary		Get user profile
//	@Description	Get current user profile
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		401	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/auth/profile [get]
func GetProfile(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return presenters.SendErrorResponse(c, fiber.StatusUnauthorized, errors.New("missing authorization header"))
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return presenters.SendErrorResponse(c, fiber.StatusUnauthorized, errors.New("invalid authorization header format"))
		}

		tokenString := tokenParts[1]
		data, err := cn.AuthHandler.GetProfileHandler(tokenString)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		return presenters.SendSuccessResponse(c, data)
	}
}

// Login godoc
//
//	@Summary		User login
//	@Description	Authenticate user and return token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.LoginRequest	true	"Login credentials"
//	@example		request
//
//	{
//	  "email": "user@example.com",
//	  "password": "password123"
//	}
//
//	@Success		200	{object}	presenters.SuccessResponse
//	@Failure		400	{object}	presenters.ErrorResponse
//	@Failure		500	{object}	presenters.ErrorResponse
//	@Router			/auth/login [post]
func Login(cn *container.AppContainer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dto.LoginRequest
		if err := c.BodyParser(&input); err != nil {
			return presenters.SendErrorResponseWithMessage(c, fiber.StatusBadRequest, "Invalid request")
		}

		dto, _, err := cn.AuthHandler.Login(&input)
		if err != nil {
			return presenters.SendErrorResponse(c, fiber.StatusInternalServerError, err)
		}

		return presenters.SendSuccessResponse(c, dto)
	}
}
