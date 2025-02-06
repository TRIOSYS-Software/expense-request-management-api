package controllers

import (
	"net/http"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/services"
	"strconv"

	"github.com/labstack/echo/v4"
)

type RolesController struct {
	RolesService *services.RolesService
}

func NewRolesController(roleService *services.RolesService) *RolesController {
	return &RolesController{RolesService: roleService}
}

func (r *RolesController) GetRoles(c echo.Context) error {
	roles := r.RolesService.GetAll()
	if roles == nil {
		return c.JSON(http.StatusNotFound, "Roles not found")
	}
	return c.JSON(http.StatusOK, roles)
}

func (r *RolesController) CreateRole(c echo.Context) error {
	role := new(models.Roles)
	if err := c.Bind(role); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	if err := r.RolesService.CreateRole(role); err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, role)
}

func (r *RolesController) GetRoleByID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	role, err := r.RolesService.GetRoleByID(uint(i))
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, role)
}

func (r *RolesController) UpdateRole(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	role := new(models.Roles)
	if err := c.Bind(role); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	role.ID = uint(id)
	if err := r.RolesService.UpdateRole(role); err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, role)
}

func (r *RolesController) DeleteRole(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	if err := r.RolesService.DeleteRole(uint(i)); err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, "Role deleted successfully")
}
