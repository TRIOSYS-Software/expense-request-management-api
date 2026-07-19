package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
	"shwetaik-expense-management-api/sqlacc"
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

func (s *AdvanceRequestsService) CreateAdvanceRequest(advanceRequest *models.AdvanceRequests) error {
	return s.AdvanceRequestsRepo.CreateAdvanceRequest(advanceRequest)
}

func (s *AdvanceRequestsService) UpdateAdvanceRequest(id uint, advanceRequest *models.AdvanceRequests) error {
	return s.AdvanceRequestsRepo.UpdateAdvanceRequest(id, advanceRequest)
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

func (s *AdvanceRequestsService) CloseAdvanceRequest(id uint, actorUserID uint, comment *string) error {
	return s.AdvanceRequestsRepo.CloseAdvanceRequest(id, actorUserID, comment)
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

	if err := sendAdvancePaymentVoucher(ctx, advance); err != nil {
		return err
	}
	return s.AdvanceRequestsRepo.UpdateSendToSQLACCStatus(advance.ID, true)
}

func sendAdvancePaymentVoucher(ctx context.Context, ar *models.AdvanceRequests) error {
	body, err := json.Marshal(buildAdvancePaymentVoucherPayload(ar))
	if err != nil {
		return err
	}

	resp, err := sqlacc.Default().Post(ctx, "/payment-vouchers/direct", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := readBodySnippet(resp.Body, 1024)
		log.Printf("[send-to-sqlacc] AR id=%d payment-vouchers POST -> %d body=%s payload=%s",
			ar.ID, resp.StatusCode, snippet, string(body))
		return fmt.Errorf("payment-vouchers POST failed: status %d body=%s", resp.StatusCode, snippet)
	}
	return nil
}

func buildAdvancePaymentVoucherPayload(ar *models.AdvanceRequests) map[string]any {
	return map[string]any{
		"docdate":       time.Now().Format("2006-01-02"),
		"paymentmethod": ar.PaymentMethod,
		"description":   ar.Description,
		"project":       ar.Project,
		"docamt":        ar.Amount,
		"sdsdocdetail": []map[string]any{
			{
				"code":        ar.GLAccounts.CODE,
				"description": ar.Description,
				"amount":      ar.Amount,
				"project":     ar.Project,
			},
		},
	}
}
