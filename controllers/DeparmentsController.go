package controllers

import (
	"net/http"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/services"
	"strconv"

	"github.com/labstack/echo/v4"
)

type DepartmentsController struct {
	DepartmentsService *services.DepartmentsService
}

func NewDepartmentsController(departmentsService *services.DepartmentsService) *DepartmentsController {
	return &DepartmentsController{DepartmentsService: departmentsService}
}

// GetDepartments get all departments
// @Summary Get all departments
// @Description Get all departments
// @Tags Departments
// @Accept json
// @Produce json
// @Success 200 {object} []models.Departments
// @Failure 404 {object} string
// @Router /departments [get]
// @Security JWT Token
func (d *DepartmentsController) GetDepartments(c echo.Context) error {
	departments, err := d.DepartmentsService.GetDepartments()
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, departments)
}

// CreateDepartment create a new department
// @Summary Create a new department
// @Description Create a new department
// @Tags Departments
// @Accept json
// @Produce json
// @Param department body models.Departments true "Department"
// @Success 200 {object} models.Departments
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Router /departments [post]
// @Security JWT Token
func (d *DepartmentsController) CreateDepartment(c echo.Context) error {
	department := new(models.Departments)
	if err := c.Bind(department); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	if err := d.DepartmentsService.CreateDepartment(department); err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, department)
}

// GetDepartmentByID get a department by id
// @Summary Get a department by id
// @Description Get a department by id
// @Tags Departments
// @Accept json
// @Produce json
// @Param id path int true "Department ID"
// @Success 200 {object} models.Departments
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Router /departments/{id} [get]
// @Security JWT Token
func (d *DepartmentsController) GetDepartmentByID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	department, err := d.DepartmentsService.GetDepartmentByID(uint(i))
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, department)
}

// UpdateDepartment update a department
// @Summary Update a department
// @Description Update a department
// @Tags Departments
// @Accept json
// @Produce json
// @Param id path int true "Department ID"
// @Param department body models.Departments true "Department"
// @Success 200 {object} models.Departments
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Router /departments/{id} [put]
// @Security JWT Token
func (d *DepartmentsController) UpdateDepartment(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	department := new(models.Departments)
	if err := c.Bind(department); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	if err := d.DepartmentsService.UpdateDepartment(uint(id), department); err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, department)
}

// DeleteDepartment delete a department
// @Summary Delete a department
// @Description Delete a department
// @Tags Departments
// @Accept json
// @Produce json
// @Param id path int true "Department ID"
// @Success 200 {object} string
// @Failure 400 {object} string
// @Failure 404 {object} string
// @Router /departments/{id} [delete]
// @Security JWT Token
func (d *DepartmentsController) DeleteDepartment(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	if err := d.DepartmentsService.DeleteDepartment(uint(i)); err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, "Department deleted successfully")
}
