package services

import (
	"context"
	"fmt"
	"time"

	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
)

type ExpenseRequestsService struct {
	ExpenseRequestsRepo *repositories.ExpenseRequestsRepo
}

func NewExpenseRequestsService(expenseRequestsRepo *repositories.ExpenseRequestsRepo) *ExpenseRequestsService {
	return &ExpenseRequestsService{ExpenseRequestsRepo: expenseRequestsRepo}
}

func (s *ExpenseRequestsService) GetExpenseRequests(approverID uint, filter *dtos.ExpenseRequestFilterDTO) ([]models.ExpenseRequests, int64) {
	return s.ExpenseRequestsRepo.GetExpenseRequests(approverID, filter)
}

func (s *ExpenseRequestsService) GetExpenseRequestByID(id uint) (*models.ExpenseRequests, error) {
	return s.ExpenseRequestsRepo.GetExpenseRequestByID(id)
}

func (s *ExpenseRequestsService) GetExpenseRequestsByUserID(id uint, filter *dtos.ExpenseRequestFilterDTO) ([]models.ExpenseRequests, int64) {
	return s.ExpenseRequestsRepo.GetExpenseRequestsByUserID(id, filter)
}

func (s *ExpenseRequestsService) GetExpenseRequestsSummary(filters map[string]any) (dtos.ExpenseRequestSummary, error) {
	return s.ExpenseRequestsRepo.GetExpenseRequestsSummary(filters)
}

func (s *ExpenseRequestsService) GetAnalytics(filters map[string]any) (dtos.AnalyticsResponse, error) {
	return s.ExpenseRequestsRepo.GetAnalytics(filters)
}

// CreateExpenseRequest validates the advance link and creates the request in a
// single transaction, then dispatches notifications once it has committed.
//
// The validation runs inside the transaction on purpose: it reads the advance's
// remaining balance, and a check made outside would let two concurrent expenses
// both pass and both commit, overdrawing the advance.
func (s *ExpenseRequestsService) CreateExpenseRequest(expenseRequest *models.ExpenseRequests) error {
	var pending []repositories.PendingNotification

	err := s.ExpenseRequestsRepo.WithTx(func(tx *repositories.ExpenseRequestsRepo) error {
		if err := ValidateAdvanceLink(tx, expenseRequest, nil); err != nil {
			return err
		}
		var createErr error
		pending, createErr = createExpenseWithChain(tx, expenseRequest)
		return createErr
	})
	if err != nil {
		return err
	}

	s.ExpenseRequestsRepo.NotifyNewRequest(expenseRequest.ID, pending)
	return nil
}

func (s *ExpenseRequestsService) GetExpenseRequestByApproverID(id uint, filter *dtos.ExpenseRequestFilterDTO) ([]models.ExpenseRequests, int64) {
	return s.ExpenseRequestsRepo.GetExpenseRequestByApproverID(id, filter)
}

// UpdateExpenseRequest validates the advance link and applies the update in one
// transaction, so the balance check and the write cannot disagree.
func (s *ExpenseRequestsService) UpdateExpenseRequest(id uint, expenseRequest *models.ExpenseRequests) error {
	return s.ExpenseRequestsRepo.WithTx(func(tx *repositories.ExpenseRequestsRepo) error {
		existing, err := tx.GetExpenseRequestForUpdate(id)
		if err != nil {
			return err
		}

		// The update form may omit user_id; carry it forward so ownership checks
		// validate against the row's real owner. Excluding the row from its own
		// balance means raising an expense that already consumed the advance still
		// validates.
		expenseRequest.UserID = existing.UserID
		if err := ValidateAdvanceLink(tx, expenseRequest, &existing.ID); err != nil {
			return err
		}

		return tx.UpdateExpenseRequestTx(id, expenseRequest)
	})
}

func (s *ExpenseRequestsService) SendExpenseRequestToSQLACC(id uint) error {
	expenseRequest, err := s.GetExpenseRequestByID(id)
	if err != nil {
		return err
	}
	if expenseRequest.IsSendToSQLACC {
		return fmt.Errorf("expense request already sent to SQLACC")
	}
	switch expenseRequest.Status {
	case "approved", "completed":
	default:
		return fmt.Errorf("expense request must be approved or completed to sync")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	voucher := voucherRequest{
		ID:                       expenseRequest.ID,
		Description:              expenseRequest.Description,
		Project:                  expenseRequest.Project,
		Amount:                   expenseRequest.Amount,
		PaymentMethod:            expenseRequest.PaymentMethod,
		PaymentMethodDescription: expenseRequest.PaymentMethods.DESCRIPTION,
		GLAccountCode:            expenseRequest.GLAccounts.CODE,
	}
	if err := sendVoucher(ctx, voucher, docTypeExpense, "ER"); err != nil {
		return err
	}
	return s.ExpenseRequestsRepo.UpdateSendToSQLACCStatus(expenseRequest.ID, true)
}

func (s *ExpenseRequestsService) CompleteExpenseRequest(id uint, actorUserID uint, comment *string) error {
	return s.ExpenseRequestsRepo.CompleteExpenseRequest(id, actorUserID, comment)
}

func (s *ExpenseRequestsService) DeleteExpenseRequest(id uint) error {
	return s.ExpenseRequestsRepo.DeleteExpenseRequest(id)
}

func (s *ExpenseRequestsService) SoftDeleteExpenseRequest(id uint) error {
	return s.ExpenseRequestsRepo.SoftDeleteExpenseRequest(id)
}
