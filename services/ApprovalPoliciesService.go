package services

import (
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
)

type ApprovalPoliciesService struct {
	repo *repositories.ApprovalPoliciesRepo
}

func NewApprovalPoliciesService(repo *repositories.ApprovalPoliciesRepo) *ApprovalPoliciesService {
	return &ApprovalPoliciesService{repo: repo}
}

func (s *ApprovalPoliciesService) GetApprovalPolicies() ([]models.ApprovalPolicies, error) {
	return s.repo.GetApprovalPolicies()
}

func (s *ApprovalPoliciesService) GetApprovalPolicyByID(id uint) (*models.ApprovalPolicies, error) {
	return s.repo.GetApprovalPolicyByID(id)
}

func (s *ApprovalPoliciesService) CreateApprovalPolicy(approvalPolicy *models.ApprovalPolicies) error {
	return s.repo.CreateApprovalPolicy(approvalPolicy)
}

func (s *ApprovalPoliciesService) UpdateApprovalPolicy(id uint, approvalPolicy *models.ApprovalPolicies) error {
	return s.repo.UpdateApprovalPolicy(id, approvalPolicy)
}

func (s *ApprovalPoliciesService) DeleteApprovalPolicy(id uint) error {
	return s.repo.DeleteApprovalPolicy(id)
}
