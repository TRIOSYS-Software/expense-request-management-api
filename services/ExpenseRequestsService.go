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

func (s *ExpenseRequestsService) UpdateExpenseRequest(expenseRequest *models.ExpenseRequests) error {
	return s.ExpenseRequestsRepo.UpdateExpenseRequest(expenseRequest)
}

func (s *ExpenseRequestsService) SendExpenseRequestToSQLACC(expenseRequestDTO *dtos.ApprovedExpenseRequestsDTO) error {
	expenseRequest, err := s.GetExpenseRequestByID(expenseRequestDTO.ExpenseID)
	if err != nil {
		return err
	}
	if expenseRequest.Status == "approved" {
		if err := callSQLACCAPI(expenseRequest, expenseRequestDTO.PaymentMethod); err != nil {
			return err
		}
		expenseRequest.IsSendToSQLACC = true
		if err := s.UpdateExpenseRequest(expenseRequest); err != nil {
			return err
		}
	}
	return nil
}

func callSQLACCAPI(expenseRequest *models.ExpenseRequests, paymentMethod string) error {
	data := map[string]any{
		"DOCNO":         fmt.Sprintf("APP-PV-%d", expenseRequest.ID),
		"DOCTYPE":       "PV",
		"DESCRIPTION":   expenseRequest.Description,
		"PAYMENTMETHOD": paymentMethod,
		"PROJECT":       *expenseRequest.Project,
		"DETAILS": []map[string]any{
			{
				"CODE":           paymentMethod,
				"DESCRIPTION":    expenseRequest.Description,
				"PROJECT":        *expenseRequest.Project,
				"AMOUNT":         expenseRequest.Amount,
				"LOCALAMOUNT":    expenseRequest.Amount,
				"CURRENCYAMOUNT": expenseRequest.Amount,
			},
		},
	}
	fmt.Println(data)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "http://localhost:1323/api/v1/payments", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("ShweTaik", getToken())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API call failed with status code: %d", resp.StatusCode)
	}

	return nil
}

func getToken() string {
	token, err := helper.Encrypt(configs.Envs.SQLACC_API_PASSWORD, configs.Envs.SQLACC_API_KEY)
	if err != nil {
		panic(err)
	}
	return token
}
