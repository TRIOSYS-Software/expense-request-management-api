package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	helper "shwetaik-expense-management-api/Helper"
	"shwetaik-expense-management-api/configs"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
)

type PaymentMethodService struct {
	repo *repositories.PaymentMethodRepo
}

func NewPaymentMethodService(repo *repositories.PaymentMethodRepo) *PaymentMethodService {
	return &PaymentMethodService{repo: repo}
}

func (s *PaymentMethodService) GetPaymentMethods() ([]models.PaymentMethod, error) {
	return s.repo.GetPaymentMethods()
}

func FetchAllPaymentMethods() ([]models.PaymentMethod, error) {
	api := fmt.Sprintf("%s/%s", configs.Envs.SQLACC_API_URL, "payment-methods")
	req, err := http.NewRequest("GET", api, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("ShweTaik", helper.GetToken(configs.Envs.SQLACC_API_PASSWORD, configs.Envs.SQLACC_API_KEY))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch payment methods: %s", resp.Status)
	}

	var paymentMethods []models.PaymentMethod
	err = json.NewDecoder(resp.Body).Decode(&paymentMethods)
	if err != nil {
		return nil, err
	}
	return paymentMethods, nil
}

func (s *PaymentMethodService) SyncPaymentMethods() error {
	paymentMethods, err := FetchAllPaymentMethods()
	if err != nil {
		return err
	}
	return s.repo.SavePaymentMethods(paymentMethods)
}
