package services

import (
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
)

type PermissionsService struct {
	Repo *repositories.PermissionsRepo
}

func NewPermissionsService(repo *repositories.PermissionsRepo) *PermissionsService {
	return &PermissionsService{Repo: repo}
}

func (s *PermissionsService) GetAll() ([]models.Permissions, error) {
	return s.Repo.GetAll()
}
