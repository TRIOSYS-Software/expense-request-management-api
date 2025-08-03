package repositories

import (
	"fmt"
	"shwetaik-expense-management-api/dtos"
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
	err := a.db.Preload("PolicyUsers", func(db *gorm.DB) *gorm.DB { return db.Order("level ASC") }).
		Preload("PolicyUsers.Approver", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).
		Preload("Departments").
		Preload("Projects").
		Find(&approvalPolicies).Error
	return approvalPolicies, err
}

func (a *ApprovalPoliciesRepo) GetApprovalPolicyByID(id uint) (*models.ApprovalPolicies, error) {
	var approvalPolicy models.ApprovalPolicies
	err := a.db.Preload("PolicyUsers", func(db *gorm.DB) *gorm.DB { return db.Order("level ASC") }).Preload("PolicyUsers.Approver", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).Preload("Departments").First(&approvalPolicy, id).Error
	return &approvalPolicy, err
}

func (a *ApprovalPoliciesRepo) CreateApprovalPolicy(approvalPolicyDTO *dtos.ApprovalPolicyRequestDTO) error {
	tx := a.db.Begin()

	approvalPolicy := models.ApprovalPolicies{
		MinAmount:    approvalPolicyDTO.MinAmount,
		MaxAmount:    approvalPolicyDTO.MaxAmount,
		Project:      approvalPolicyDTO.Project,
		DepartmentID: approvalPolicyDTO.DepartmentID,
	}

	if IsAmountRangeOverlapping(tx, approvalPolicy.Project, approvalPolicy.MinAmount, approvalPolicy.MaxAmount, approvalPolicy.DepartmentID) {
		tx.Rollback()
		return fmt.Errorf("amount range overlapping")
	}

	if err := tx.Create(&approvalPolicy).Error; err != nil {
		tx.Rollback()
		return err
	}

	var PolicyUser []models.ApprovalPoliciesUsers
	for _, user := range approvalPolicyDTO.Approvers {
		PolicyUser = append(PolicyUser, models.ApprovalPoliciesUsers{
			ApprovalPolicyID: approvalPolicy.ID,
			UserID:           user.ApproverID,
			Level:            user.Level,
		})
	}

	if len(PolicyUser) > 0 {
		if err := tx.Create(&PolicyUser).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func IsAmountRangeOverlapping(db *gorm.DB, project string, minAmount float64, maxAmount float64, deprtmentID *uint) bool {
	var count int64

	if deprtmentID != nil {
		if err := db.Model(&models.ApprovalPolicies{}).
			Where("project = ?", project).
			Where("NOT (max_amount < ? OR min_amount > ?)", minAmount, maxAmount).
			Where("department_id = ?", deprtmentID).
			Count(&count).Error; err != nil {
			return false
		}
	} else {
		if err := db.Model(&models.ApprovalPolicies{}).
			Where("project = ?", project).
			Where("NOT (max_amount < ? OR min_amount > ?)", minAmount, maxAmount).
			Where("department_id IS NULL").
			Count(&count).Error; err != nil {
			return false
		}
	}

	return count > 0
}

func (a *ApprovalPoliciesRepo) UpdateApprovalPolicy(id uint, approvalPolicyDTO *dtos.ApprovalPolicyRequestDTO) error {
	tx := a.db.Begin()

	var approvalPoliciesToUpdate models.ApprovalPolicies
	if err := tx.Find(&approvalPoliciesToUpdate, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	approvalPoliciesToUpdate.MinAmount = approvalPolicyDTO.MinAmount
	approvalPoliciesToUpdate.MaxAmount = approvalPolicyDTO.MaxAmount
	approvalPoliciesToUpdate.Project = approvalPolicyDTO.Project
	approvalPoliciesToUpdate.DepartmentID = approvalPolicyDTO.DepartmentID

	// if IsAmountRangeOverlapping(tx, approvalPoliciesToUpdate.Project, approvalPoliciesToUpdate.MinAmount, approvalPoliciesToUpdate.MaxAmount, approvalPoliciesToUpdate.DepartmentID) {
	// 	tx.Rollback()
	// 	return fmt.Errorf("amount range overlapping")
	// }

	if err := tx.Save(&approvalPoliciesToUpdate).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("approval_policy_id = ?", id).Delete(&models.ApprovalPoliciesUsers{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	var PolicyUser []models.ApprovalPoliciesUsers
	for _, user := range approvalPolicyDTO.Approvers {
		PolicyUser = append(PolicyUser, models.ApprovalPoliciesUsers{
			ApprovalPolicyID: approvalPoliciesToUpdate.ID,
			UserID:           user.ApproverID,
			Level:            user.Level,
		})
	}

	if len(PolicyUser) > 0 {
		if err := tx.Create(&PolicyUser).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (a *ApprovalPoliciesRepo) DeleteApprovalPolicy(id uint) error {
	tx := a.db.Begin()
	var approvalPolicy models.ApprovalPolicies
	if err := tx.First(&approvalPolicy, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("approval_policy_id = ?", id).Delete(&models.ApprovalPoliciesUsers{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&approvalPolicy).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
