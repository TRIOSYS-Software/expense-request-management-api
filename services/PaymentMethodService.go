package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
	"shwetaik-expense-management-api/sqlacc"
)

const (
	paymentMethodsPath       = "/payment-methods"
	paymentMethodSyncTimeout = 5 * time.Minute
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

type paymentMethodDTO struct {
	Code         string `json:"code"`
	Journal      string `json:"journal"`
	CurrencyCode string `json:"currency_code"`
	Description  string `json:"description"`
}

type paymentMethodListEnvelope struct {
	Status  string             `json:"status"`
	Message string             `json:"message"`
	Data    []paymentMethodDTO `json:"data"`
}

func FetchAllPaymentMethods(ctx context.Context) ([]models.PaymentMethod, error) {
	const tag = "[sync:payment-methods]"
	client := sqlacc.Default()

	listStart := time.Now()
	resp, err := client.Get(ctx, paymentMethodsPath, nil)
	if err != nil {
		return nil, err
	}
	var env paymentMethodListEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		resp.Body.Close()
		return nil, fmt.Errorf("decode payment methods: %w", err)
	}
	resp.Body.Close()

	out := make([]models.PaymentMethod, len(env.Data))
	for i, dto := range env.Data {
		out[i] = models.PaymentMethod{
			CODE:         dto.Code,
			JOURNAL:      dto.Journal,
			CURRENCYCODE: dto.CurrencyCode,
			DESCRIPTION:  dto.Description,
		}
	}
	log.Printf("%s list: %d in %s", tag, len(out), time.Since(listStart).Round(time.Millisecond))
	return out, nil
}

func (s *PaymentMethodService) SyncPaymentMethods() error {
	const tag = "[sync:payment-methods]"
	start := time.Now()
	log.Printf("%s start", tag)

	ctx, cancel := context.WithTimeout(context.Background(), paymentMethodSyncTimeout)
	defer cancel()

	paymentMethods, err := FetchAllPaymentMethods(ctx)
	if err != nil {
		log.Printf("%s failed during fetch after %s: %v", tag, time.Since(start).Round(time.Millisecond), err)
		return err
	}

	saveStart := time.Now()
	counts, err := s.repo.SavePaymentMethods(paymentMethods)
	if err != nil {
		log.Printf("%s failed during save after %s: %v", tag, time.Since(start).Round(time.Millisecond), err)
		return err
	}
	log.Printf("%s save: upserted=%d deleted=%d in %s", tag, counts.Upserted, counts.Deleted, time.Since(saveStart).Round(time.Millisecond))
	log.Printf("%s done in %s", tag, time.Since(start).Round(time.Millisecond))
	return nil
}
