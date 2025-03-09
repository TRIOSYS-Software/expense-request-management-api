package services

import (
	"shwetaik-expense-management-api/dtos"
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

func (s *ApprovalPoliciesService) CreateApprovalPolicy(approvalPolicyDTO *dtos.ApprovalPolicyRequestDTO) error {
	return s.repo.CreateApprovalPolicy(approvalPolicyDTO)
}

func (s *ApprovalPoliciesService) UpdateApprovalPolicy(id uint, approvalPolicyDTO *dtos.ApprovalPolicyRequestDTO) error {
	return s.repo.UpdateApprovalPolicy(id, approvalPolicyDTO)
}

func (s *ApprovalPoliciesService) DeleteApprovalPolicy(id uint) error {
	return s.repo.DeleteApprovalPolicy(id)
}
