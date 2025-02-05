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

func (d *DepartmentsController) GetDepartments(c echo.Context) error {
	departments, err := d.DepartmentsService.GetDepartments()
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, departments)
}

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

func (d *DepartmentsController) UpdateDepartment(c echo.Context) error {
	department := new(models.Departments)
	if err := c.Bind(department); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	if err := d.DepartmentsService.UpdateDepartment(department); err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, department)
}

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
