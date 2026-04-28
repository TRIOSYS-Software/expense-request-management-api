package controllers

import (
	"net/http"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/services"
	"strconv"

	"github.com/labstack/echo/v4"
)

type ExpenseCategoriesController struct {
	ExpenseCategoriesService *services.ExpenseCategoriesService
}

func NewExpenseCategoriesController(expenseCategoriesService *services.ExpenseCategoriesService) *ExpenseCategoriesController {
	return &ExpenseCategoriesController{ExpenseCategoriesService: expenseCategoriesService}
}

func (ec *ExpenseCategoriesController) GetExpenseCategories(c echo.Context) error {
	ExpenseCategories, err := ec.ExpenseCategoriesService.GetExpenseCategories()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, ExpenseCategories)
}

func (ec *ExpenseCategoriesController) GetExpenseCategoryByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	ExpenseCategory, err := ec.ExpenseCategoriesService.GetExpenseCategoryByID(uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, ExpenseCategory)
}

func (ec *ExpenseCategoriesController) CreateExpenseCategory(c echo.Context) error {
	var expenseCategory models.ExpenseCategories
	if err := c.Bind(&expenseCategory); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	if err := ec.ExpenseCategoriesService.CreateExpenseCategory(&expenseCategory); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, expenseCategory)
}

func (ec *ExpenseCategoriesController) UpdateExpenseCategory(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	var expenseCategory models.ExpenseCategories
	if err := c.Bind(&expenseCategory); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	expenseCategory.ID = uint(id)
	if err := ec.ExpenseCategoriesService.UpdateExpenseCategory(&expenseCategory); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, expenseCategory)
}

func (ec *ExpenseCategoriesController) DeleteExpenseCategory(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error()})
	}
	if err := ec.ExpenseCategoriesService.DeleteExpenseCategory(uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": err.Error()})
	}
	return c.JSON(http.StatusOK, "Expense category deleted successfully")
}
