package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
	"shwetaik-expense-management-api/sqlacc"
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

func (s *ExpenseRequestsService) CreateExpenseRequest(expenseRequest *models.ExpenseRequests) error {
	return s.ExpenseRequestsRepo.CreateExpenseRequest(expenseRequest)
}

func (s *ExpenseRequestsService) GetExpenseRequestByApproverID(id uint, filter *dtos.ExpenseRequestFilterDTO) ([]models.ExpenseRequests, int64) {
	return s.ExpenseRequestsRepo.GetExpenseRequestByApproverID(id, filter)
}

func (s *ExpenseRequestsService) UpdateExpenseRequest(id uint, expenseRequest *models.ExpenseRequests) error {
	return s.ExpenseRequestsRepo.UpdateExpenseRequest(id, expenseRequest)
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

	if err := sendPaymentVoucher(ctx, expenseRequest); err != nil {
		return err
	}
	return s.ExpenseRequestsRepo.UpdateSendToSQLACCStatus(expenseRequest.ID, true)
}

func (s *ExpenseRequestsService) CompleteExpenseRequest(id uint, actorUserID uint, comment *string) error {
	return s.ExpenseRequestsRepo.CompleteExpenseRequest(id, actorUserID, comment)
}

func sendPaymentVoucher(ctx context.Context, er *models.ExpenseRequests) error {
	body, err := json.Marshal(buildPaymentVoucherPayload(er))
	if err != nil {
		return err
	}

	resp, err := sqlacc.Default().Post(ctx, "/payment-vouchers", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := readBodySnippet(resp.Body, 1024)
		log.Printf("[send-to-sqlacc] ER id=%d payment-vouchers POST -> %d body=%s payload=%s",
			er.ID, resp.StatusCode, snippet, string(body))
		return fmt.Errorf("payment-vouchers POST failed: status %d body=%s", resp.StatusCode, snippet)
	}
	return nil
}

func readBodySnippet(r io.Reader, max int) string {
	if r == nil {
		return ""
	}
	buf, err := io.ReadAll(io.LimitReader(r, int64(max)))
	if err != nil {
		return fmt.Sprintf("<read err: %v>", err)
	}
	s := strings.TrimSpace(string(buf))
	if s == "" {
		return "<empty>"
	}
	return s
}

func buildPaymentVoucherPayload(er *models.ExpenseRequests) map[string]any {
	return map[string]any{
		"docdate":       time.Now().Format("2006-01-02"),
		"paymentmethod": er.PaymentMethod,
		"description":   er.Description,
		"project":       er.Project,
		"docamt":        er.Amount,
		"sdsdocdetail": []map[string]any{
			{
				"code":        er.GLAccounts.CODE,
				"description": er.Description,
				"amount":      er.Amount,
				"project":     er.Project,
			},
		},
	}
}

func (s *ExpenseRequestsService) DeleteExpenseRequest(id uint) error {
	return s.ExpenseRequestsRepo.DeleteExpenseRequest(id)
}

func (s *ExpenseRequestsService) SoftDeleteExpenseRequest(id uint) error {
	return s.ExpenseRequestsRepo.SoftDeleteExpenseRequest(id)
}
