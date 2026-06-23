package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
	if expenseRequest.Status != "approved" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := sendPaymentVoucher(ctx, expenseRequest); err != nil {
		return err
	}
	return s.ExpenseRequestsRepo.UpdateSendToSQLACCStatus(expenseRequest.ID, true)
}

func paymentVoucherDocNo(expenseRequest *models.ExpenseRequests) string {
	if strings.Contains(strings.ToLower(expenseRequest.PaymentMethods.DESCRIPTION), "bank") {
		return fmt.Sprintf("APP-B-PV-%d", expenseRequest.ID)
	}
	return fmt.Sprintf("APP-C-PV-%d", expenseRequest.ID)
}

// paymentVoucherExists returns true if SQL Acc already has a PV with this
// docno. The deterministic docno scheme (APP-{B|C}-PV-{id}) is our
// idempotency key, so a retry after a flaky-network success doesn't double-post.
func paymentVoucherExists(ctx context.Context, docno string) (bool, error) {
	q := url.Values{}
	q.Set("docno", docno)

	resp, err := sqlacc.Default().Get(ctx, "/paymentvoucher", q)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("paymentvoucher precheck: status %d", resp.StatusCode)
	}

	var env struct {
		Data  []map[string]any `json:"data"`
		Count int              `json:"count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return false, fmt.Errorf("paymentvoucher precheck decode: %w", err)
	}
	return len(env.Data) > 0, nil
}

func sendPaymentVoucher(ctx context.Context, er *models.ExpenseRequests) error {
	docno := paymentVoucherDocNo(er)

	exists, err := paymentVoucherExists(ctx, docno)
	if err != nil {
		return err
	}
	if exists {
		// Already in SQL Acc — treat as success so the local flag flips
		// without re-posting and creating a duplicate.
		return nil
	}

	body, err := json.Marshal(buildPaymentVoucherPayload(er, docno))
	if err != nil {
		return err
	}

	resp, err := sqlacc.Default().Post(ctx, "/paymentvoucher", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("paymentvoucher POST failed: status %d", resp.StatusCode)
	}
	return nil
}

// buildPaymentVoucherPayload maps an approved expense request to the SQL Acc
// /paymentvoucher schema. Field selection mirrors the postman collection; the
// numeric fields are sent as strings ("0.00") per the example body. Most
// optional fields (irbm_*, eiv_*, address*, …) are left empty.
//
// NOTE: pending vendor confirmation of required fields. Values here are based
// on the postman example and the local model.
func buildPaymentVoucherPayload(er *models.ExpenseRequests, docno string) map[string]any {
	today := time.Now().Format("2006-01-02")
	amount := formatMoney(er.Amount)

	detail := map[string]any{
		"seq":            0,
		"code":           er.GLAccounts.CODE,
		"project":        er.Project,
		"description":    er.Description,
		"amount":         amount,
		"localamount":    amount,
		"currencyamount": amount,
		"currencycode":   er.PaymentMethods.CURRENCYCODE,
		"taxinclusive":   false,
		"changed":        true,
	}

	return map[string]any{
		"docno":         docno,
		"doctype":       "PV",
		"docdate":       today,
		"postdate":      today,
		"description":   er.Description,
		"paymentmethod": er.PaymentMethod,
		"project":       er.Project,
		"currencycode":  er.PaymentMethods.CURRENCYCODE,
		"docamt":        amount,
		"localdocamt":   amount,
		"cancelled":     false,
		"changed":       true,
		"sdsdocdetail":  []map[string]any{detail},
	}
}

func formatMoney(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func (s *ExpenseRequestsService) DeleteExpenseRequest(id uint) error {
	return s.ExpenseRequestsRepo.DeleteExpenseRequest(id)
}
