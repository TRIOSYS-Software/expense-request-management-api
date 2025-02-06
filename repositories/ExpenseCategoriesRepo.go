package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

type ExpenseCategoriesRepo struct {
	db *gorm.DB
}

func NewExpenseCategoriesRepo(db *gorm.DB) *ExpenseCategoriesRepo {
	return &ExpenseCategoriesRepo{db: db}
}

func (ecr *ExpenseCategoriesRepo) GetExpenseCategories() ([]models.ExpenseCategories, error) {
	var expenseCategories []models.ExpenseCategories
	err := ecr.db.Find(&expenseCategories).Error
	return expenseCategories, err
}

func (ecr *ExpenseCategoriesRepo) GetExpenseCategoryByID(id uint) (*models.ExpenseCategories, error) {
	var expenseCategory models.ExpenseCategories
	err := ecr.db.First(&expenseCategory, id).Error
	return &expenseCategory, err
}

func (ecr *ExpenseCategoriesRepo) CreateExpenseCategory(expenseCategory *models.ExpenseCategories) error {
	return ecr.db.Create(expenseCategory).Error
}

func (ecr *ExpenseCategoriesRepo) UpdateExpenseCategory(expenseCategory *models.ExpenseCategories) error {
	return ecr.db.Save(expenseCategory).Error
}

func (ecr *ExpenseCategoriesRepo) DeleteExpenseCategory(id uint) error {
	return ecr.db.Delete(&models.ExpenseCategories{}, id).Error
}
