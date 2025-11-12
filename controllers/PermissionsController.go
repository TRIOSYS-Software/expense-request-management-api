package controllers

import (
	"net/http"
	"shwetaik-expense-management-api/services"

	"github.com/labstack/echo/v4"
)

type PermissionsController struct {
	Service *services.PermissionsService
}

func NewPermissionsController(s *services.PermissionsService) *PermissionsController {
	return &PermissionsController{Service: s}
}

func (p *PermissionsController) GetPermissions(c echo.Context) error {
	perms, err := p.Service.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, perms)
}
