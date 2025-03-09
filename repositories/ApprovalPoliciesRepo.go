package repositories

import (
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
	err := a.db.Preload("PolicyUsers", func(db *gorm.DB) *gorm.DB { return db.Order("level ASC") }).Preload("PolicyUsers.Approver", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).Preload("Departments").Find(&approvalPolicies).Error
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
		ConditionType:  approvalPolicyDTO.ConditionType,
		ConditionValue: approvalPolicyDTO.ConditionValue,
		Priority:       approvalPolicyDTO.Priority,
		DepartmentID:   approvalPolicyDTO.DepartmentID,
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

func (a *ApprovalPoliciesRepo) UpdateApprovalPolicy(id uint, approvalPolicyDTO *dtos.ApprovalPolicyRequestDTO) error {
	tx := a.db.Begin()

	var approvalPoliciesToUpdate models.ApprovalPolicies
	if err := tx.Find(&approvalPoliciesToUpdate, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	approvalPoliciesToUpdate.ConditionType = approvalPolicyDTO.ConditionType
	approvalPoliciesToUpdate.ConditionValue = approvalPolicyDTO.ConditionValue
	approvalPoliciesToUpdate.Priority = approvalPolicyDTO.Priority
	approvalPoliciesToUpdate.DepartmentID = approvalPolicyDTO.DepartmentID

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
	var approvalPolicy models.ApprovalPolicies
	if err := a.db.First(&approvalPolicy, id).Error; err != nil {
		return err
	}

	if err := a.db.Model(&approvalPolicy).Association("Approver").Clear(); err != nil {
		return err
	}

	return a.db.Delete(&approvalPolicy).Error
}
