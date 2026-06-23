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
	projectsPath       = "/project"
	projectSyncTimeout = 5 * time.Minute
	pageSizeHint       = 200
)

type ProjectService struct {
	repo *repositories.ProjectRepo
}

func NewProjectService(repo *repositories.ProjectRepo) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) GetProjects() ([]models.Project, error) {
	return s.repo.GetProjects()
}

// projectDTO mirrors the official SQL Acc /project response field names.
// Numeric fields come back as strings (e.g. "0.00"); parsed during mapping.
type projectDTO struct {
	Code         string `json:"code"`
	Description  string `json:"description"`
	Description2 string `json:"description2"`
	ProjectValue string `json:"projectvalue"`
	ProjectCost  string `json:"projectcost"`
	IsActive     bool   `json:"isactive"`
}

type projectListEnvelope struct {
	Data  []projectDTO `json:"data"`
	Count int          `json:"count"`
}

func FetchAllProjects(ctx context.Context) ([]models.Project, error) {
	client := sqlacc.Default()
	var all []models.Project
	offset := 0

	for {
		q := url.Values{}
		q.Set("offset", strconv.Itoa(offset))
		q.Set("limit", strconv.Itoa(pageSizeHint))

		resp, err := client.Get(ctx, projectsPath, q)
		if err != nil {
			return nil, err
		}
		var env projectListEnvelope
		if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decode projects offset=%d: %w", offset, err)
		}
		resp.Body.Close()

		if len(env.Data) == 0 {
			break
		}
		for _, dto := range env.Data {
			all = append(all, projectFromDTO(dto))
		}
		offset += len(env.Data)
	}
	return all, nil
}

func projectFromDTO(d projectDTO) models.Project {
	p := models.Project{
		CODE: d.Code,
	}
	if d.Description != "" {
		desc := d.Description
		p.DESCRIPTION = &desc
	}
	if d.Description2 != "" {
		d2 := d.Description2
		p.DESCRIPTION2 = &d2
	}
	if v, err := strconv.ParseFloat(d.ProjectValue, 64); err == nil {
		p.PROJECTVALUE = &v
	}
	if v, err := strconv.ParseFloat(d.ProjectCost, 64); err == nil {
		p.PROJECTCOST = &v
	}
	active := d.IsActive
	p.ISACTIVE = &active
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
