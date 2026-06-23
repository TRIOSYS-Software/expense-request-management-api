package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
	"shwetaik-expense-management-api/sqlacc"

	"golang.org/x/sync/errgroup"
)

const (
	paymentMethodsPath        = "/pmmethod"
	paymentMethodSyncTimeout  = 15 * time.Minute
	paymentMethodDetailWorker = 5
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

// paymentMethodDTO is the shape returned by both /pmmethod (list) and
// /pmmethod/<code> (detail). The list response omits `description`, so we
// must call the detail endpoint per code to populate it.
type paymentMethodDTO struct {
	Code         string `json:"code"`
	Journal      string `json:"journal"`
	CurrencyCode string `json:"currencycode"`
	Description  string `json:"description"`
}

type paymentMethodListEnvelope struct {
	Data  []paymentMethodDTO `json:"data"`
	Count int                `json:"count"`
}

func FetchAllPaymentMethods(ctx context.Context) ([]models.PaymentMethod, error) {
	const tag = "[sync:payment-methods]"
	client := sqlacc.Default()

	// 1. List all payment-method codes (paginated).
	listStart := time.Now()
	var listed []paymentMethodDTO
	offset := 0
	for {
		q := url.Values{}
		q.Set("offset", strconv.Itoa(offset))
		q.Set("limit", strconv.Itoa(pageSizeHint))

		resp, err := client.Get(ctx, paymentMethodsPath, q)
		if err != nil {
			return nil, err
		}
		var env paymentMethodListEnvelope
		if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decode payment methods offset=%d: %w", offset, err)
		}
		resp.Body.Close()

		if len(env.Data) == 0 {
			break
		}
		listed = append(listed, env.Data...)
		offset += len(env.Data)
	}
	log.Printf("%s list: %d codes in %s", tag, len(listed), time.Since(listStart).Round(time.Millisecond))

	// 2. For each listed code, fetch the detail to get `description`.
	// Run with bounded concurrency: detail calls are independent and the cloud
	// round-trip dominates wall time, so sequential N+1 over a flaky link
	// would burn the sync's whole context budget.
	detailStart := time.Now()
	out := make([]models.PaymentMethod, len(listed))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(paymentMethodDetailWorker)

	for i, item := range listed {
		i, code := i, item.Code
		g.Go(func() error {
			detail, err := fetchPaymentMethodDetail(gctx, client, code)
			if err != nil {
				return err
			}
			out[i] = models.PaymentMethod{
				CODE:         detail.Code,
				JOURNAL:      detail.Journal,
				CURRENCYCODE: detail.CurrencyCode,
				DESCRIPTION:  detail.Description,
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	log.Printf("%s details: %d in %s (concurrency=%d)", tag, len(out), time.Since(detailStart).Round(time.Millisecond), paymentMethodDetailWorker)
	return out, nil
}

func fetchPaymentMethodDetail(ctx context.Context, client *sqlacc.SQLAccClient, code string) (paymentMethodDTO, error) {
	resp, err := client.Get(ctx, paymentMethodsPath+"/"+url.PathEscape(code), nil)
	if err != nil {
		return paymentMethodDTO{}, fmt.Errorf("fetch payment-method detail %s: %w", code, err)
	}
	defer resp.Body.Close()

	var env paymentMethodListEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return paymentMethodDTO{}, fmt.Errorf("decode payment-method detail %s: %w", code, err)
	}
	if len(env.Data) == 0 {
		return paymentMethodDTO{}, fmt.Errorf("payment-method detail %s: empty data", code)
	}
	return env.Data[0], nil
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
