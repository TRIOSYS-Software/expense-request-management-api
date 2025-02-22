package services

import (
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
)

type ExpenseApprovalsService struct {
	repo *repositories.ExpenseApprovalsRepo
}

func NewExpenseApprovalsService(repo *repositories.ExpenseApprovalsRepo) *ExpenseApprovalsService {
	return &ExpenseApprovalsService{repo: repo}
}

func (s *ExpenseApprovalsService) GetExpenseApprovals() []models.ExpenseApprovals {
	return s.repo.GetExpenseApprovals()
}

func (s *ExpenseApprovalsService) GetExpenseApprovalsByApproverID(approverID uint) []models.ExpenseApprovals {
	return s.repo.GetExpenseApprovalsByApproverID(approverID)
}

func (s *ExpenseApprovalsService) UpdateExpenseApproval(id uint, expenseApproval *models.ExpenseApprovals) error {
	return s.repo.UpdateExpenseApproval(id, expenseApproval)
}
