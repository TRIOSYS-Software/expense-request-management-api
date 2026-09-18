package services

import (
	"context"
	"fmt"
	"time"

	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
)

type AdvanceRequestsService struct {
	AdvanceRequestsRepo *repositories.AdvanceRequestsRepo
}

func NewAdvanceRequestsService(repo *repositories.AdvanceRequestsRepo) *AdvanceRequestsService {
	return &AdvanceRequestsService{AdvanceRequestsRepo: repo}
}

func (s *AdvanceRequestsService) GetAdvanceRequests(approverID uint, filter *dtos.AdvanceRequestFilterDTO) ([]models.AdvanceRequests, int64) {
	return s.AdvanceRequestsRepo.GetAdvanceRequests(approverID, filter)
}

func (s *AdvanceRequestsService) GetAdvanceRequestByID(id uint) (*models.AdvanceRequests, error) {
	return s.AdvanceRequestsRepo.GetAdvanceRequestByID(id)
}

func (s *AdvanceRequestsService) GetAdvanceRequestsByUserID(id uint, filter *dtos.AdvanceRequestFilterDTO) ([]models.AdvanceRequests, int64) {
	return s.AdvanceRequestsRepo.GetAdvanceRequestsByUserID(id, filter)
}

func (s *AdvanceRequestsService) GetAdvanceRequestByApproverID(id uint, filter *dtos.AdvanceRequestFilterDTO) ([]models.AdvanceRequests, int64) {
	return s.AdvanceRequestsRepo.GetAdvanceRequestByApproverID(id, filter)
}

func (s *AdvanceRequestsService) GetAdvanceRequestsSummary(filters map[string]any) (dtos.AdvanceRequestSummary, error) {
	return s.AdvanceRequestsRepo.GetAdvanceRequestsSummary(filters)
}

func (s *AdvanceRequestsService) GetSelectableAdvanceRequests(userID uint) ([]models.AdvanceRequests, error) {
	return s.AdvanceRequestsRepo.GetSelectableAdvanceRequests(userID)
}

// CreateAdvanceRequest creates the request and its approval chain in one
// transaction, dispatching notifications only after it commits.
func (s *AdvanceRequestsService) CreateAdvanceRequest(advanceRequest *models.AdvanceRequests) error {
	var pending []repositories.PendingNotification

	err := s.AdvanceRequestsRepo.WithTx(func(tx *repositories.AdvanceRequestsRepo) error {
		var createErr error
		pending, createErr = createAdvanceWithChain(tx, advanceRequest)
		return createErr
	})
	if err != nil {
		return err
	}

	s.AdvanceRequestsRepo.NotifyNewAdvanceRequest(advanceRequest.ID, pending)
	return nil
}

func (s *AdvanceRequestsService) UpdateAdvanceRequest(id uint, advanceRequest *models.AdvanceRequests) error {
	return s.AdvanceRequestsRepo.WithTx(func(tx *repositories.AdvanceRequestsRepo) error {
		return tx.UpdateAdvanceRequestTx(id, advanceRequest)
	})
}

func (s *AdvanceRequestsService) DeleteAdvanceRequest(id uint) error {
	return s.AdvanceRequestsRepo.DeleteAdvanceRequest(id)
}

func (s *AdvanceRequestsService) SoftDeleteAdvanceRequest(id uint) error {
	return s.AdvanceRequestsRepo.SoftDeleteAdvanceRequest(id)
}

func (s *AdvanceRequestsService) CountLinkedExpenseRequests(id uint) (int64, error) {
	return s.AdvanceRequestsRepo.CountLinkedExpenseRequests(id)
}

// CloseAdvanceRequest closes the advance transactionally, then notifies the
// requester once the close has committed.
func (s *AdvanceRequestsService) CloseAdvanceRequest(id uint, actorUserID uint, comment *string) error {
	var pending *repositories.PendingNotification

	err := s.AdvanceRequestsRepo.WithTx(func(tx *repositories.AdvanceRequestsRepo) error {
		var closeErr error
		pending, closeErr = tx.CloseAdvanceRequestTx(id, actorUserID, comment)
		return closeErr
	})
	if err != nil {
		return err
	}

	s.AdvanceRequestsRepo.NotifyAdvanceClosed(id, pending)
	return nil
}

func (s *AdvanceRequestsService) SendAdvanceRequestToSQLACC(id uint) error {
	advance, err := s.GetAdvanceRequestByID(id)
	if err != nil {
		return err
	}
	if advance.IsSendToSQLACC {
		return fmt.Errorf("advance request already sent to SQLACC")
	}
	switch advance.Status {
	case "approved", "completed", "closed":
	default:
		return fmt.Errorf("advance request must be approved, completed, or closed to sync")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	voucher := voucherRequest{
		ID:                       advance.ID,
		Description:              advance.Description,
		Project:                  advance.Project,
		Amount:                   advance.Amount,
		PaymentMethod:            advance.PaymentMethod,
		PaymentMethodDescription: advance.PaymentMethods.DESCRIPTION,
		GLAccountCode:            advance.GLAccounts.CODE,
	}
	if err := sendVoucher(ctx, voucher, docTypeAdvance, "AR"); err != nil {
		return err
	}
	return s.AdvanceRequestsRepo.UpdateSendToSQLACCStatus(advance.ID, true)
}
