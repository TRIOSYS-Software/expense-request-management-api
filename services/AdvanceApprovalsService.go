package services

import (
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
)

type AdvanceApprovalsService struct {
	repo *repositories.AdvanceApprovalsRepo
}

func NewAdvanceApprovalsService(repo *repositories.AdvanceApprovalsRepo) *AdvanceApprovalsService {
	return &AdvanceApprovalsService{repo: repo}
}

func (s *AdvanceApprovalsService) GetAdvanceApprovals() []models.AdvanceApprovals {
	return s.repo.GetAdvanceApprovals()
}

func (s *AdvanceApprovalsService) GetAdvanceApprovalsByApproverID(approverID uint) []models.AdvanceApprovals {
	return s.repo.GetAdvanceApprovalsByApproverID(approverID)
}

func (s *AdvanceApprovalsService) UpdateAdvanceApproval(id uint, advanceApproval *models.AdvanceApprovals) error {
	return s.repo.UpdateAdvanceApproval(id, advanceApproval)
}

func (s *AdvanceApprovalsService) UpdateAdvanceApprovalComment(id uint, comments string) error {
	return s.repo.UpdateAdvanceApprovalComment(id, comments)
}
