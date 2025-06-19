package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	helper "shwetaik-expense-management-api/Helper"
	"shwetaik-expense-management-api/configs"
	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
	"strings"
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

func (s *ExpenseRequestsService) GetExpenseRequestsSummary(filters map[string]any) (dtos.ExpenseRequestSummary, error) {
	return s.ExpenseRequestsRepo.GetExpenseRequestsSummary(filters)
}

func (s *ExpenseRequestsService) CreateExpenseRequest(expenseRequest *models.ExpenseRequests) error {
	return s.ExpenseRequestsRepo.CreateExpenseRequest(expenseRequest)
}

func (s *ExpenseRequestsService) GetExpenseRequestByApproverID(id uint) []models.ExpenseRequests {
	return s.ExpenseRequestsRepo.GetExpenseRequestByApproverID(id)
}

func (s *ExpenseRequestsService) UpdateExpenseRequest(id uint, expenseRequest *models.ExpenseRequests) error {
	return s.ExpenseRequestsRepo.UpdateExpenseRequest(id, expenseRequest)
}

func (s *ExpenseRequestsService) SendExpenseRequestToSQLACC(id uint) error {
	expenseRequest, err := s.GetExpenseRequestByID(id)
	if err != nil {
		return err
	}
	if expenseRequest.Status == "approved" {
		if err := callSQLACCAPI(expenseRequest, expenseRequest.PaymentMethods.DESCRIPTION); err != nil {
			return err
		}
		expenseRequest.IsSendToSQLACC = true
		if err := s.UpdateExpenseRequest(expenseRequest.ID, expenseRequest); err != nil {
			return err
		}
	}
	return nil
}

func callSQLACCAPI(expenseRequest *models.ExpenseRequests, paymentMethod string) error {
	var docKey string
	if strings.Contains(strings.ToLower(paymentMethod), "bank") {
		docKey = fmt.Sprintf("APP-B-PV-%d", expenseRequest.ID)
	} else {
		docKey = fmt.Sprintf("APP-C-PV-%d", expenseRequest.ID)
	}
	data := map[string]any{
		"DOCNO":         docKey,
		"DOCTYPE":       "PV",
		"DESCRIPTION":   expenseRequest.Description,
		"PAYMENTMETHOD": expenseRequest.PaymentMethod,
		"PROJECT":       expenseRequest.Project,
		"DETAILS": []map[string]any{
			{
				"CODE":           expenseRequest.GLAccounts.CODE,
				"DESCRIPTION":    expenseRequest.Description,
				"PROJECT":        expenseRequest.Project,
				"AMOUNT":         expenseRequest.Amount,
				"LOCALAMOUNT":    expenseRequest.Amount,
				"CURRENCYAMOUNT": expenseRequest.Amount,
			},
		},
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	api := fmt.Sprintf("%s/%s", configs.Envs.SQLACC_API_URL, "payments")

	req, err := http.NewRequest("POST", api, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("ShweTaik", helper.GetToken(configs.Envs.SQLACC_API_PASSWORD, configs.Envs.SQLACC_API_KEY))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API call failed with status code: %d", resp.StatusCode)
	}

	return nil
}

func (s *ExpenseRequestsService) DeleteExpenseRequest(id uint) error {
	return s.ExpenseRequestsRepo.DeleteExpenseRequest(id)
}
