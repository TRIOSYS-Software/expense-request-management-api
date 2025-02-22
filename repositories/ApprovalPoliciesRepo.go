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
	err := a.db.Preload("Approver").Preload("Departments").Find(&approvalPolicies).Error
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

func (a *ApprovalPoliciesRepo) UpdateApprovalPolicy(id uint, approvalPolicy *models.ApprovalPolicies) error {
	tx := a.db.Begin()

	var approvalPoliciesToUpdate models.ApprovalPolicies
	if err := tx.Find(&approvalPoliciesToUpdate, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	approvalPolicy.ID = approvalPoliciesToUpdate.ID
	approvalPolicy.CreatedAt = approvalPoliciesToUpdate.CreatedAt

	if err := tx.Save(approvalPolicy).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&approvalPolicy).Association("Approver").Replace(approvalPolicy.Approver); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (a *ApprovalPoliciesRepo) DeleteApprovalPolicy(id uint) error {
	var approvalPolicy models.ApprovalPolicies
	if err := a.db.First(&approvalPolicy, id).Error; err != nil {
		return err
	}

	if err := a.db.Model(&approvalPolicy).Association("Approver").Clear(); err != nil {
		return err
	}

	return a.db.Delete(&approvalPolicy).Error
}
