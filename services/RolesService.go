package services

import (
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
)

type RolesService struct {
	RolesRepo *repositories.RolesRepo
	PermissionsRepo *repositories.PermissionsRepo
}

func NewRolesService(RoleRepo *repositories.RolesRepo, permRepo *repositories.PermissionsRepo) *RolesService {
	return &RolesService{
		RolesRepo: RoleRepo,
		PermissionsRepo: permRepo,
	}
}

func (s *RolesService) GetAll() ([]models.Roles, error) {
	return s.RolesRepo.GetRoles()
}

func (s *RolesService) GetRoleByID(id uint) (*models.Roles, error) {
	return s.RolesRepo.GetRoleByID(id)
}

func (s *RolesService) CreateRoleWithPermissions(role *models.Roles, permissionIDs []uint) error {
	if len(permissionIDs) > 0 {
		if _, err := s.PermissionsRepo.GetByIDs(permissionIDs); err != nil {
			return err
		}
	}
	return s.RolesRepo.CreateRoleWithPermissions(role, permissionIDs)
}

func (s *RolesService) UpdateRoleWithPermissions(role *models.Roles, permissionIDs []uint) error {
	if len(permissionIDs) > 0 {
		if _, err := s.PermissionsRepo.GetByIDs(permissionIDs); err != nil {
			return err
		}
	}
	return s.RolesRepo.UpdateRoleWithPermissions(role, permissionIDs)
}

func (s *RolesService) DeleteRole(id uint) error {
	return s.RolesRepo.DeleteRoleIfUnused(id)
}