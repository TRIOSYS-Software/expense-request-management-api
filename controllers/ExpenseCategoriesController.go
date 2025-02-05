package controllers

import (
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/services"
)

type ExpenseCategoriesController struct {
	ExpenseCategoriesService *services.ExpenseCategoriesService
}

func NewExpenseCategoriesController(expenseCategoriesService *services.ExpenseCategoriesService) *ExpenseCategoriesController {
	return &ExpenseCategoriesController{ExpenseCategoriesService: expenseCategoriesService}
}

func (ec *ExpenseCategoriesController) GetExpenseCategories() []models.ExpenseCategories {
	return ec.ExpenseCategoriesService.GetExpenseCategories()
}

func (ec *ExpenseCategoriesController) GetExpenseCategoryByID(id uint) (models.ExpenseCategories, error) {
	return ec.ExpenseCategoriesService.GetExpenseCategoryByID(id)
}

func (ec *ExpenseCategoriesController) CreateExpenseCategory(expenseCategory *models.ExpenseCategories) error {
	return ec.ExpenseCategoriesService.CreateExpenseCategory(expenseCategory)
}

func (ec *ExpenseCategoriesController) UpdateExpenseCategory(expenseCategory *models.ExpenseCategories) error {
	return ec.ExpenseCategoriesService.UpdateExpenseCategory(expenseCategory)
}

func (ec *ExpenseCategoriesController) DeleteExpenseCategory(id uint) error {
	return ec.ExpenseCategoriesService.DeleteExpenseCategory(id)
}
