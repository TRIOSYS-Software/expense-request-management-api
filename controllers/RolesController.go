package controllers

import (
	"net/http"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/services"
	"strconv"

	"github.com/labstack/echo/v4"
)
type CreateRoleRequest struct {
	Name          string `json:"name" validate:"required"`
	Description   string `json:"description"`
	IsAdmin       bool   `json:"is_admin"`
	PermissionIDs []uint `json:"permission_ids"`
}

type UpdateRoleRequest struct {
	Name          string `json:"name" validate:"required"`
	Description   string `json:"description"`
	IsAdmin       bool   `json:"is_admin"`
	PermissionIDs []uint `json:"permission_ids"`
}

type RolesController struct {
	RolesService *services.RolesService
}

func NewRolesController(roleService *services.RolesService) *RolesController {
	return &RolesController{RolesService: roleService}
}

func (r *RolesController) GetRoles(c echo.Context) error {
	roles, err := r.RolesService.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	if len(roles) == 0 {
		return c.JSON(http.StatusOK, []models.Roles{})
	}
	return c.JSON(http.StatusOK, roles)
}

func (r *RolesController) GetRoleByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid Role ID"})
	}
	role, err := r.RolesService.GetRoleByID(uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, role)
}

func (r *RolesController) CreateRole(c echo.Context) error {
	req := new(CreateRoleRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Name is required"})
	}

	role := &models.Roles{
		Name: req.Name,
		Description: req.Description,
		IsAdmin:     req.IsAdmin,
	}
	if err := r.RolesService.CreateRoleWithPermissions(role, req.PermissionIDs); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	created, _ := r.RolesService.GetRoleByID(role.ID)
	return c.JSON(http.StatusCreated, created)
}

func (r *RolesController) UpdateRole(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid Role ID"})
	}
	req := new(UpdateRoleRequest)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Name is required"})
	}

	role := &models.Roles{
		ID:   uint(id),
		Name: req.Name,
		Description: req.Description,
		IsAdmin:     req.IsAdmin,
	}
	
	if err := r.RolesService.UpdateRoleWithPermissions(role, req.PermissionIDs); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	updated, _ := r.RolesService.GetRoleByID(role.ID)
	return c.JSON(http.StatusOK, updated)
}

func (r *RolesController) DeleteRole(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid Role ID"})
	}
	if err := r.RolesService.DeleteRole(uint(id)); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"message": "Role deleted successfully"})
}