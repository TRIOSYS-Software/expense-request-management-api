package controllers

import (
	"net/http"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/services"
	"strconv"

	"github.com/labstack/echo/v4"
)

type ExpenseRequestsController struct {
	ExpenseRequestsService *services.ExpenseRequestsService
}

func NewExpenseRequestsController(expenseRequestsService *services.ExpenseRequestsService) *ExpenseRequestsController {
	return &ExpenseRequestsController{ExpenseRequestsService: expenseRequestsService}
}

func (ex *ExpenseRequestsController) GetExpenseRequests(c echo.Context) error {
	expenseRequests := ex.ExpenseRequestsService.GetExpenseRequests()
	return c.JSON(http.StatusOK, expenseRequests)
}

func (ex *ExpenseRequestsController) GetExpenseRequestByID(c echo.Context) error {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid expense request id")
	}
	expenseRequest, err := ex.ExpenseRequestsService.GetExpenseRequestByID(uint(i))
	if err != nil {
		return c.JSON(http.StatusNotFound, err)
	}
	return c.JSON(http.StatusOK, expenseRequest)
}

func (ex *ExpenseRequestsController) CreateExpenseRequest(c echo.Context) error {
	expenseRequest := new(models.ExpenseRequests)
	if err := c.Bind(expenseRequest); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	if err := ex.ExpenseRequestsService.CreateExpenseRequest(expenseRequest); err != nil {
		return c.JSON(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, expenseRequest)
}
