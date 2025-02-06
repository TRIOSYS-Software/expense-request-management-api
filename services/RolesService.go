package services

import (
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
)

type RolesService struct {
	RolesRepo *repositories.RolesRepo
}

func NewRolesService(RoleRepo *repositories.RolesRepo) *RolesService {
	return &RolesService{RolesRepo: RoleRepo}
}

func (r *RolesService) GetAll() []models.Roles {
	return r.RolesRepo.GetRoles()
}

func (r *RolesService) CreateRole(role *models.Roles) error {
	return r.RolesRepo.CreateRole(role)
}

func (r *RolesService) GetRoleByID(id uint) (*models.Roles, error) {
	return r.RolesRepo.GetRoleByID(id)
}

func (r *RolesService) UpdateRole(role *models.Roles) error {
	return r.RolesRepo.UpdateRole(role)
}

func (r *RolesService) DeleteRole(id uint) error {
	return r.RolesRepo.DeleteRole(id)
}
