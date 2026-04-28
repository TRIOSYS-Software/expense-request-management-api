package repositories

import (
	"fmt"
	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/models"
	"strconv"

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
		Preload("GLAccounts").
		Find(&approvalPolicies).Error
	return approvalPolicies, err
}

func (a *ApprovalPoliciesRepo) GetApprovalPolicyByID(id uint) (*models.ApprovalPolicies, error) {
	var approvalPolicy models.ApprovalPolicies
	err := a.db.Preload("PolicyUsers", func(db *gorm.DB) *gorm.DB { return db.Order("level ASC") }).
		Preload("PolicyUsers.Approver", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).
		Preload("Departments").
		Preload("GLAccounts").
		First(&approvalPolicy, id).Error
	return &approvalPolicy, err
}

func glAccsFromIDs(ids []string) []models.GLAcc {
	var glAccs []models.GLAcc
	for _, idStr := range ids {
		dockey, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		glAccs = append(glAccs, models.GLAcc{DOCKEY: dockey})
	}
	return glAccs
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
		return fmt.Errorf("You cannot create a policy for an existing amount range.")
	}

	if err := tx.Create(&approvalPolicy).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(approvalPolicyDTO.GLAccountIDs) > 0 {
		glAccs := glAccsFromIDs(approvalPolicyDTO.GLAccountIDs)
		if err := tx.Model(&approvalPolicy).Association("GLAccounts").Replace(glAccs); err != nil {
			tx.Rollback()
			return err
		}
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

func IsAmountRangeOverlapping(db *gorm.DB, project string, minAmount float64, maxAmount float64, departmentID *uint) bool {
	var count int64

	query := db.Model(&models.ApprovalPolicies{}).
		Where("project = ?", project).
		Where("NOT (max_amount < ? OR min_amount > ?)", minAmount, maxAmount)

	if departmentID != nil {
		query = query.Where("department_id = ?", departmentID)
	} else {
		query = query.Where("department_id IS NULL")
	}

	if err := query.Count(&count).Error; err != nil {
		return false
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

	if err := tx.Save(&approvalPoliciesToUpdate).Error; err != nil {
		tx.Rollback()
		return err
	}

	glAccs := glAccsFromIDs(approvalPolicyDTO.GLAccountIDs)
	if err := tx.Model(&approvalPoliciesToUpdate).Association("GLAccounts").Replace(glAccs); err != nil {
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

	if err := tx.Where("approval_policy_id = ?", id).Delete(&models.ApprovalPolicyGLAccount{}).Error; err != nil {
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
