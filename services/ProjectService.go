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
	projectsPath       = "/projects"
	projectSyncTimeout = 5 * time.Minute
	pageSizeHint       = 200
)

type pageMeta struct {
	Limit   int    `json:"limit"`
	After   string `json:"after"`
	HasMore bool   `json:"has_more"`
}

type ProjectService struct {
	repo *repositories.ProjectRepo
}

func NewProjectService(repo *repositories.ProjectRepo) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) GetProjects() ([]models.Project, error) {
	return s.repo.GetProjects()
}

type projectDTO struct {
	Code         string   `json:"code"`
	Description  string   `json:"description"`
	Description2 string   `json:"description2"`
	ProjectValue *float64 `json:"project_value"`
	ProjectCost  *float64 `json:"project_cost"`
	IsActive     *bool    `json:"is_active"`
}

type projectListEnvelope struct {
	Status     string       `json:"status"`
	Message    string       `json:"message"`
	Data       []projectDTO `json:"data"`
	Pagination pageMeta     `json:"pagination"`
}

func FetchAllProjects(ctx context.Context) ([]models.Project, error) {
	client := sqlacc.Default()
	var all []models.Project
	after := ""

	for {
		q := url.Values{}
		q.Set("limit", strconv.Itoa(pageSizeHint))
		if after != "" {
			q.Set("after", after)
		}

		resp, err := client.Get(ctx, projectsPath, q)
		if err != nil {
			return nil, err
		}
		var env projectListEnvelope
		if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decode projects after=%q: %w", after, err)
		}
		resp.Body.Close()

		for _, dto := range env.Data {
			all = append(all, projectFromDTO(dto))
		}
		if !env.Pagination.HasMore || env.Pagination.After == "" {
			break
		}
		after = env.Pagination.After
	}
	return all, nil
}

func projectFromDTO(d projectDTO) models.Project {
	p := models.Project{CODE: d.Code}
	if d.Description != "" {
		desc := d.Description
		p.DESCRIPTION = &desc
	}
	if d.Description2 != "" {
		d2 := d.Description2
		p.DESCRIPTION2 = &d2
	}
	p.PROJECTVALUE = d.ProjectValue
	p.PROJECTCOST = d.ProjectCost
	if d.IsActive != nil {
		active := *d.IsActive
		p.ISACTIVE = &active
	}
	return p
}

func (s *ProjectService) SyncProjects() error {
	const tag = "[sync:projects]"
	start := time.Now()
	log.Printf("%s start", tag)

	ctx, cancel := context.WithTimeout(context.Background(), projectSyncTimeout)
	defer cancel()

	fetchStart := time.Now()
	projects, err := FetchAllProjects(ctx)
	if err != nil {
		log.Printf("%s failed during fetch after %s: %v", tag, time.Since(start).Round(time.Millisecond), err)
		return err
	}
	log.Printf("%s fetch: %d rows in %s", tag, len(projects), time.Since(fetchStart).Round(time.Millisecond))

	saveStart := time.Now()
	counts, err := s.repo.SaveProjects(projects)
	if err != nil {
		log.Printf("%s failed during save after %s: %v", tag, time.Since(start).Round(time.Millisecond), err)
		return err
	}
	log.Printf("%s save: upserted=%d deleted=%d in %s", tag, counts.Upserted, counts.Deleted, time.Since(saveStart).Round(time.Millisecond))
	log.Printf("%s done in %s", tag, time.Since(start).Round(time.Millisecond))
	return nil
}
