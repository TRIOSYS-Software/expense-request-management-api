package controllers

import (
	"net/http"
	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/services"
	"strconv"

	"github.com/labstack/echo/v4"
)

type DeviceTokenController struct {
	service *services.DeviceTokenService
}

func NewDeviceTokenController(service *services.DeviceTokenService) *DeviceTokenController {
	return &DeviceTokenController{service: service}
}

// GET /users/:id/device-tokens
func (c *DeviceTokenController) GetTokensByUserID(ctx echo.Context) error {
	userIDParam := ctx.Param("id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	tokens, err := c.service.GetTokensByUserID(uint(userID))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	response := dtos.DeviceTokensResponse{
		UserID: uint(userID),
		Tokens: tokens,
	}
	return ctx.JSON(http.StatusOK, response)
}

// POST /users/:id/device-tokens
func (c *DeviceTokenController) CreateTokenByUserID(ctx echo.Context) error {
	userIDParam := ctx.Param("id")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	var req dtos.DeviceTokenRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	req.UserID = uint(userID)

	deviceToken, err := c.service.CreateTokenByUserID(req.UserID, req.Token, req.DeviceOS)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusCreated, deviceToken)
}

func (c *DeviceTokenController) DeleteToken(ctx echo.Context) error {
	token := ctx.Param("token")
	if token == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Missing token"})
	}

	if err := c.service.DeleteToken(token); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.NoContent(http.StatusNoContent)
}