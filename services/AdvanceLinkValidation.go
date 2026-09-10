package services

import (
	"fmt"

	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"

	"gorm.io/gorm"
)

const balanceEpsilon = 1e-6

func ValidateAdvanceLink(repo *repositories.ExpenseRequestsRepo, request *models.ExpenseRequests, excludeID *uint) error {
	if request.Amount < 0 {
		return fmt.Errorf("Expense amount cannot be negative")
	}

	if request.AdvanceRequestID == nil {
		request.AdvanceUsedAmount = nil
		request.ReturnedAmount = nil
		return nil
	}

	if request.AdvanceUsedAmount == nil || *request.AdvanceUsedAmount <= 0 {
		return fmt.Errorf("Advance used amount is required and must be greater than zero when an advance request is linked")
	}
	used := *request.AdvanceUsedAmount

	returned := 0.0
	if request.ReturnedAmount != nil {
		returned = *request.ReturnedAmount
		if returned <= 0 {
			return fmt.Errorf("Returned amount, when provided, must be greater than zero")
		}
	}

	advance, err := repo.FindAdvanceRequest(*request.AdvanceRequestID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("Linked advance request not found")
		}
		return fmt.Errorf("Failed to load linked advance request: %w", err)
	}
	if advance.UserID != request.UserID {
		return fmt.Errorf("You may only link an advance request created by yourself")
	}
	if advance.Status != "approved" {
		return fmt.Errorf("Only an approved advance request may be linked")
	}

	remaining, err := repo.AdvanceRemainingFor(advance, excludeID)
	if err != nil {
		return fmt.Errorf("Failed to compute advance request balance: %w", err)
	}
	if used > remaining+balanceEpsilon {
		return fmt.Errorf("Advance used amount (%.2f) exceeds the advance request's remaining balance (%.2f)", used, remaining)
	}

	leftover := used - request.Amount
	if leftover < 0 {
		leftover = 0
	}
	if returned > leftover+balanceEpsilon {
		return fmt.Errorf("Returned amount (%.2f) cannot exceed the unused portion of the advance used (%.2f)", returned, leftover)
	}

	return nil
}
