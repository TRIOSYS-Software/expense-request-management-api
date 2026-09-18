package services

import (
	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
)

type ApproversService struct {
	ApproversRepo *repositories.ApproversRepo
}

func NewApproversService(approversRepo *repositories.ApproversRepo) *ApproversService {
	return &ApproversService{ApproversRepo: approversRepo}
}

func (s *ApproversService) GetApprovers() ([]models.Users, error) {
	return s.ApproversRepo.GetApprovers()
}

func (s *ApproversService) GetApproverByID(id uint) (*models.Users, error) {
	return s.ApproversRepo.GetApproverByID(id)
}

func (s *ApproversService) GetApproverActions(approverID, viewerRoleID, viewerID uint) (dtos.ApproverActionsResult, error) {
	viewerIsAdmin := s.ApproversRepo.IsRoleAdmin(viewerRoleID)
	return s.ApproversRepo.GetApproverActionsForViewer(approverID, viewerID, viewerIsAdmin)
}
