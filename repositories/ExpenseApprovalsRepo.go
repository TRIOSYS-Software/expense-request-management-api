package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

type ExpenseApprovalsRepo struct {
	db *gorm.DB
}

func NewExpenseApprovalsRepo(db *gorm.DB) *ExpenseApprovalsRepo {
	return &ExpenseApprovalsRepo{db: db}
}

func (r *ExpenseApprovalsRepo) GetExpenseApprovals() []models.ExpenseApprovals {
	var expenseApprovals []models.ExpenseApprovals
	r.db.Preload("Users", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, email")
	}).Find(&expenseApprovals)
	return expenseApprovals
}

func (r *ExpenseApprovalsRepo) GetExpenseApprovalsByApproverID(approverID uint) []models.ExpenseApprovals {
	var expenseApprovals []models.ExpenseApprovals
	r.db.Where("approver_id = ?", approverID).Preload("Users", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, email")
	}).Find(&expenseApprovals)
	return expenseApprovals
}

func (r *ExpenseApprovalsRepo) UpdateExpenseApproval(id uint, expenseApproval *models.ExpenseApprovals) error {
	tx := r.db.Begin()
	var expenseApprovalToUpdate models.ExpenseApprovals
	if err := tx.Where("id = ?", id).First(&expenseApprovalToUpdate).Error; err != nil {
		return err
	}

	var expenseRequest models.ExpenseRequests
	tx.Preload("Approvals").Where("id = ?", expenseApprovalToUpdate.RequestID).First(&expenseRequest)
	if expenseApproval.Status == "rejected" {
		expenseRequest.Status = "rejected"
		if err := tx.Save(&expenseRequest).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if expenseApproval.Status == "approved" {
		expenseRequest.CurrentApproverLevel += 1
		// Get the maximum approver level from all approvals
		var maxLevel uint
		tx.Model(&models.ExpenseApprovals{}).Where("request_id = ?", expenseRequest.ID).Select("MAX(level)").Scan(&maxLevel)

		// If we've reached or exceeded the maximum level, mark as approved
		if expenseRequest.CurrentApproverLevel > maxLevel {
			expenseRequest.Status = "approved"
		}

		if err := tx.Save(&expenseRequest).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	expenseApprovalToUpdate.Status = expenseApproval.Status
	expenseApprovalToUpdate.Comments = expenseApproval.Comments
	expenseApprovalToUpdate.ApprovalDate = expenseApproval.ApprovalDate

	if err := tx.Save(&expenseApprovalToUpdate).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
