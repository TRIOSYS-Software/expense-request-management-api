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

type GLAccService struct {
	Repo *repositories.GLAccRepo
}

func NewGLAccService(repo *repositories.GLAccRepo) *GLAccService {
	return &GLAccService{Repo: repo}
}

func (s *GLAccService) GetGLAcc() ([]models.GLAcc, error) {
	return s.Repo.GetGLAcc()
}

func FetchAllGLAcc() ([]models.GLAcc, error) {
	uri := fmt.Sprintf("gl-accounts/codes?codes=%v", configs.Envs.FILTER_GL_CODE)
	api := fmt.Sprintf("%s/%s", configs.Envs.SQLACC_API_URL, uri)
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

	var GLAccs []models.GLAcc
	err = json.NewDecoder(resp.Body).Decode(&GLAccs)
	if err != nil {
		return nil, err
	}
	return GLAccs, nil
}

func (s *GLAccService) DeleteGLAccs(glAccs []models.GLAcc) error {
	oldGLAccs, err := s.Repo.GetGLAcc()
	if err != nil {
		return err
	}
	glAccMap := make(map[int]models.GLAcc, len(glAccs))
	for _, newGLAcc := range glAccs {
		glAccMap[newGLAcc.DOCKEY] = newGLAcc
	}
	for _, oldGLAcc := range oldGLAccs {
		if _, ok := glAccMap[oldGLAcc.DOCKEY]; !ok {
			if err := s.Repo.DeleteGLAcc(oldGLAcc); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *GLAccService) SyncGLAcc() error {
	GLAccs, err := FetchAllGLAcc()
	if err != nil {
		return err
	}
	if err := s.DeleteGLAccs(GLAccs); err != nil {
		return err
	}
	return s.Repo.SaveGLAcc(GLAccs)
}
