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

type ProjectService struct {
	repo *repositories.ProjectRepo
}

func NewProjectService(repo *repositories.ProjectRepo) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) GetProjects() ([]models.Project, error) {
	return s.repo.GetProjects()
}

func FetchAllProjects() ([]models.Project, error) {
	api := fmt.Sprintf("%s/%s", configs.Envs.SQLACC_API_URL, "projects")
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
		return nil, fmt.Errorf("failed to fetch project: %s", resp.Status)
	}

	var projects []models.Project
	err = json.NewDecoder(resp.Body).Decode(&projects)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func (s *ProjectService) SyncProjects() error {
	projects, err := FetchAllProjects()
	if err != nil {
		return err
	}
	return s.repo.SaveProjects(projects)
}
