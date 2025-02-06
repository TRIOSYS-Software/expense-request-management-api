package services

import (
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
)

type ExpenseCategoriesService struct {
	ExpenseCategoriesRepo repositories.ExpenseCategoriesRepo
}

func NewExpenseCategoriesService(expenseCategoriesRepo *repositories.ExpenseCategoriesRepo) *ExpenseCategoriesService {
	return &ExpenseCategoriesService{ExpenseCategoriesRepo: *expenseCategoriesRepo}
}

func (ec *ExpenseCategoriesService) GetExpenseCategories() ([]models.ExpenseCategories, error) {
	return ec.ExpenseCategoriesRepo.GetExpenseCategories()
}

func (ec *ExpenseCategoriesService) GetExpenseCategoryByID(id uint) (*models.ExpenseCategories, error) {
	return ec.ExpenseCategoriesRepo.GetExpenseCategoryByID(id)
}

func (ec *ExpenseCategoriesService) CreateExpenseCategory(expenseCategory *models.ExpenseCategories) error {
	return ec.ExpenseCategoriesRepo.CreateExpenseCategory(expenseCategory)
}

func (ec *ExpenseCategoriesService) UpdateExpenseCategory(expenseCategory *models.ExpenseCategories) error {
	return ec.ExpenseCategoriesRepo.UpdateExpenseCategory(expenseCategory)
}

func (ec *ExpenseCategoriesService) DeleteExpenseCategory(id uint) error {
	return ec.ExpenseCategoriesRepo.DeleteExpenseCategory(id)
}
