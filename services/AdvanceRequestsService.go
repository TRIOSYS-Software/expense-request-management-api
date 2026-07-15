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
	if err := callSQLACCAPIForAdvance(advance, advance.PaymentMethods.DESCRIPTION); err != nil {
		return err
	}
	return s.AdvanceRequestsRepo.UpdateSendToSQLACCStatus(advance.ID, true)
}

func callSQLACCAPIForAdvance(advance *models.AdvanceRequests, paymentMethod string) error {
	var docKey string
	if strings.Contains(strings.ToLower(paymentMethod), "bank") {
		docKey = fmt.Sprintf("APP-B-ADV-%d", advance.ID)
	} else {
		docKey = fmt.Sprintf("APP-C-ADV-%d", advance.ID)
	}
	data := map[string]any{
		"DOCNO":         docKey,
		"DOCTYPE":       "PV",
		"DESCRIPTION":   advance.Description,
		"PAYMENTMETHOD": advance.PaymentMethod,
		"PROJECT":       advance.Project,
		"DETAILS": []map[string]any{
			{
				"CODE":           advance.GLAccounts.CODE,
				"DESCRIPTION":    advance.Description,
				"PROJECT":        advance.Project,
				"AMOUNT":         advance.Amount,
				"LOCALAMOUNT":    advance.Amount,
				"CURRENCYAMOUNT": advance.Amount,
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
