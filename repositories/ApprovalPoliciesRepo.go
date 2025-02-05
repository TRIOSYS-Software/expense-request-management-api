package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

type ApprovalPoliciesRepo struct {
	db *gorm.DB
}

func NewApprovalPoliciesRepo(db *gorm.DB) *ApprovalPoliciesRepo {
	return &ApprovalPoliciesRepo{db: db}
}

func (a *ApprovalPoliciesRepo) GetApprovalPolicies() ([]models.ApprovalPolicies, error) {
	var approvalPolicies []models.ApprovalPolicies
	err := a.db.Find(&approvalPolicies).Error
	return approvalPolicies, err
}

func (a *ApprovalPoliciesRepo) GetApprovalPolicyByID(id uint) (*models.ApprovalPolicies, error) {
	var approvalPolicy models.ApprovalPolicies
	err := a.db.First(&approvalPolicy, id).Error
	return &approvalPolicy, err
}

func (a *ApprovalPoliciesRepo) CreateApprovalPolicy(approvalPolicy *models.ApprovalPolicies) error {
	return a.db.Create(approvalPolicy).Error
}

func (a *ApprovalPoliciesRepo) UpdateApprovalPolicy(approvalPolicy *models.ApprovalPolicies) error {
	return a.db.Save(approvalPolicy).Error
}

func (a *ApprovalPoliciesRepo) DeleteApprovalPolicy(id uint) error {
	return a.db.Delete(&models.ApprovalPolicies{}, id).Error
}
