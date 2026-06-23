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
)

const (
	glAccountsPath   = "/account"
	glAccSyncTimeout = 10 * time.Minute
)

type GLAccService struct {
	Repo *repositories.GLAccRepo
}

func NewGLAccService(repo *repositories.GLAccRepo) *GLAccService {
	return &GLAccService{Repo: repo}
}

func (s *GLAccService) GetGLAcc() ([]models.GLAcc, error) {
	return s.Repo.GetGLAcc()
}

type glAccDTO struct {
	DocKey         int    `json:"dockey"`
	Parent         int    `json:"parent"`
	Code           string `json:"code"`
	Description    string `json:"description"`
	Description2   string `json:"description2"`
	AccType        string `json:"acctype"`
	SpecialAccType string `json:"specialacctype"`
	Tax            string `json:"tax"`
	CashflowType   int    `json:"cashflowtype"`
	SIC            string `json:"sic"`
}

type glAccListEnvelope struct {
	Data  []glAccDTO `json:"data"`
	Count int        `json:"count"`
}

// FetchAllGLAcc pulls the entire chart of accounts from SQL Acc by paginating
// /account with no code filter.
func FetchAllGLAcc(ctx context.Context) ([]models.GLAcc, error) {
	client := sqlacc.Default()
	var all []models.GLAcc
	offset := 0

	for {
		q := url.Values{}
		q.Set("offset", strconv.Itoa(offset))
		q.Set("limit", strconv.Itoa(pageSizeHint))

		resp, err := client.Get(ctx, glAccountsPath, q)
		if err != nil {
			return nil, err
		}
		var env glAccListEnvelope
		if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decode gl-acc offset=%d: %w", offset, err)
		}
		resp.Body.Close()

		if len(env.Data) == 0 {
			break
		}
		for _, dto := range env.Data {
			all = append(all, glAccFromDTO(dto))
		}
		offset += len(env.Data)
	}
	return all, nil
}

func glAccFromDTO(d glAccDTO) models.GLAcc {
	return models.GLAcc{
		DOCKEY:         d.DocKey,
		PARENT:         d.Parent,
		CODE:           d.Code,
		DESCRIPTION:    d.Description,
		DESCRIPTION2:   d.Description2,
		ACCTYPE:        d.AccType,
		SPECIALACCTYPE: d.SpecialAccType,
		TAX:            d.Tax,
		CASHFLOWTYPE:   d.CashflowType,
		SIC:            d.SIC,
	}
}

func (s *GLAccService) SyncGLAcc() error {
	const tag = "[sync:gl-accounts]"
	start := time.Now()
	log.Printf("%s start", tag)

	ctx, cancel := context.WithTimeout(context.Background(), glAccSyncTimeout)
	defer cancel()

	fetchStart := time.Now()
	glAccs, err := FetchAllGLAcc(ctx)
	if err != nil {
		log.Printf("%s failed during fetch after %s: %v", tag, time.Since(start).Round(time.Millisecond), err)
		return err
	}
	log.Printf("%s fetch: %d rows in %s", tag, len(glAccs), time.Since(fetchStart).Round(time.Millisecond))

	saveStart := time.Now()
	counts, err := s.Repo.ReplaceGLAcc(glAccs)
	if err != nil {
		log.Printf("%s failed during save after %s: %v", tag, time.Since(start).Round(time.Millisecond), err)
		return err
	}
	log.Printf("%s save: upserted=%d deleted=%d in %s", tag, counts.Upserted, counts.Deleted, time.Since(saveStart).Round(time.Millisecond))
	log.Printf("%s done in %s", tag, time.Since(start).Round(time.Millisecond))
	return nil
}
