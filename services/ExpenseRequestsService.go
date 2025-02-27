package services

import (
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
)

type ExpenseRequestsService struct {
	ExpenseRequestsRepo *repositories.ExpenseRequestsRepo
}

func NewExpenseRequestsService(expenseRequestsRepo *repositories.ExpenseRequestsRepo) *ExpenseRequestsService {
	return &ExpenseRequestsService{ExpenseRequestsRepo: expenseRequestsRepo}
}

func (s *ExpenseRequestsService) GetExpenseRequests() []models.ExpenseRequests {
	return s.ExpenseRequestsRepo.GetExpenseRequests()
}

func (s *ExpenseRequestsService) GetExpenseRequestByID(id uint) (*models.ExpenseRequests, error) {
	return s.ExpenseRequestsRepo.GetExpenseRequestByID(id)
}

func (s *ExpenseRequestsService) GetExpenseRequestsByUserID(id uint) []models.ExpenseRequests {
	return s.ExpenseRequestsRepo.GetExpenseRequestsByUserID(id)
}

func (s *ExpenseRequestsService) GetExpenseRequestsSummary(filters map[string]any) (map[string]any, error) {
	return s.ExpenseRequestsRepo.GetExpenseRequestsSummary(filters)
}

func (s *ExpenseRequestsService) CreateExpenseRequest(expenseRequest *models.ExpenseRequests) error {
	return s.ExpenseRequestsRepo.CreateExpenseRequest(expenseRequest)
}

func (s *ExpenseRequestsService) GetExpenseRequestByApproverID(id uint) []models.ExpenseRequests {
	return s.ExpenseRequestsRepo.GetExpenseRequestByApproverID(id)
}
