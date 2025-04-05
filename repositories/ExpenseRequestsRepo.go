package repositories

import (
	"fmt"
	"os"
	"path/filepath"
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

type ExpenseRequestsRepo struct {
	db *gorm.DB
}

func NewExpenseRequestsRepo(db *gorm.DB) *ExpenseRequestsRepo {
	return &ExpenseRequestsRepo{db: db}
}

func (r *ExpenseRequestsRepo) GetExpenseRequests() []models.ExpenseRequests {
	var expenseRequests []models.ExpenseRequests
	r.db.Preload("Projects").Preload("GLAccounts").
		Preload("PaymentMethods", func(db *gorm.DB) *gorm.DB { return db.Select("CODE, DESCRIPTION") }).
		Preload("Approvals").Preload("Approvals.Users", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, email, role_id, department_id")
	}).
		Preload("Approvals.Users.Roles").Preload("Approvals.Users.Departments").
		Preload("Category").
		Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).
		Order("expense_requests.created_at DESC").
		Find(&expenseRequests)
	return expenseRequests
}

func (r *ExpenseRequestsRepo) GetExpenseRequestByID(id uint) (*models.ExpenseRequests, error) {
	var expenseRequest models.ExpenseRequests
	err := r.db.Preload("Projects").
		Preload("GLAccounts").
		Preload("PaymentMethods", func(db *gorm.DB) *gorm.DB { return db.Select("CODE, DESCRIPTION") }).
		Preload("Approvals").
		Preload("Approvals.Users", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).
		Preload("Category").
		Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).
		First(&expenseRequest, id).Error
	return &expenseRequest, err
}

func (r *ExpenseRequestsRepo) GetExpenseRequestsByUserID(id uint) []models.ExpenseRequests {
	var expenseRequests []models.ExpenseRequests
	r.db.Where("user_id = ?", id).Preload("Approvals.Users", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, email")
	}).Preload("Category").Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).Order("expense_requests.created_at DESC").Find(&expenseRequests)
	return expenseRequests
}

func (r *ExpenseRequestsRepo) GetExpenseRequestsSummary(filters map[string]any) (map[string]any, error) {
	var expenseRequests []models.ExpenseRequests
	var summary = make(map[string]any)

	db := r.db.Model(&models.ExpenseRequests{}).Preload("Approvals")
	if filters["user_id"] != nil {
		db = db.Where("user_id = ?", filters["user_id"])
	}
	if filters["status"] != nil {
		db = db.Where("expense_requests.status = ?", filters["status"].(string))
	}
	if filters["category_id"] != nil {
		db = db.Where("category_id = ?", filters["category_id"])
	}

	if filters["start_date"] != nil && filters["end_date"] != nil {
		db = db.Where("date_submitted BETWEEN ? AND ?", filters["start_date"], filters["end_date"])
		summary["daily_totals"] = make(map[string]float64)
	}

	if filters["amount"] != nil {
		db = db.Where("amount = ?", filters["amount"])
	}

	if filters["approver_id"] != nil {
		db = db.Joins("JOIN expense_approvals ON expense_approvals.request_id = expense_requests.id").
			Where("expense_approvals.approver_id = ?", filters["approver_id"])
	}

	db.Find(&expenseRequests)

	summary["total"] = len(expenseRequests)
	summary["pending"] = 0
	summary["approved"] = 0
	summary["rejected"] = 0
	summary["total_amount"] = 0.00

	for _, expenseRequest := range expenseRequests {
		summary["total_amount"] = summary["total_amount"].(float64) + expenseRequest.Amount
		if expenseRequest.Status == "pending" {
			summary["pending"] = summary["pending"].(int) + 1
		} else if expenseRequest.Status == "approved" {
			summary["approved"] = summary["approved"].(int) + 1
		} else if expenseRequest.Status == "rejected" {
			summary["rejected"] = summary["rejected"].(int) + 1
		}

		if filters["start_date"] != nil && filters["end_date"] != nil {
			date := expenseRequest.DateSubmitted.Format("2006-01-02")
			summary["daily_totals"].(map[string]float64)[date] = summary["daily_totals"].(map[string]float64)[date] + expenseRequest.Amount
		}
	}
	return summary, nil
}

func (r *ExpenseRequestsRepo) CreateExpenseRequest(expenseRequest *models.ExpenseRequests) error {
	tx := r.db.Begin()
	if err := tx.Create(expenseRequest).Error; err != nil {
		tx.Rollback()
		return err
	}
	var requestUser models.Users
	tx.Where("id = ?", expenseRequest.UserID).First(&requestUser)

	approvalPolicy, err := r.FindHighestPolicy(expenseRequest, *requestUser.DepartmentID)
	if err != nil {
		tx.Rollback()
		return err
	}

	var approvalPoliciesUsers []models.ApprovalPoliciesUsers
	tx.Preload("Approver").Where("approval_policy_id = ?", approvalPolicy.ID).Order("level ASC").Find(&approvalPoliciesUsers)

	if len(approvalPoliciesUsers) == 0 {
		tx.Rollback()
		return fmt.Errorf("no approver users found")
	}

	for i, approverPolicyUser := range approvalPoliciesUsers {
		expenseApprovals := models.ExpenseApprovals{
			RequestID:  expenseRequest.ID,
			ApproverID: approverPolicyUser.UserID,
			Level:      approverPolicyUser.Level,
			Status:     "pending",
			IsFinal:    i == len(approvalPoliciesUsers)-1,
		}
		if err := tx.Create(&expenseApprovals).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *ExpenseRequestsRepo) FindHighestPolicy(request *models.ExpenseRequests, departmentID uint) (*models.ApprovalPolicies, error) {
	var approvalPolicy models.ApprovalPolicies
	err := r.db.Where("(department_id = ? OR department_id IS NULL) AND project = ? AND ? BETWEEN min_amount AND max_amount", departmentID, request.Project, request.Amount).First(&approvalPolicy).Error
	if err != nil {
		return nil, fmt.Errorf("no approval policy found")
	}

	// for _, approvalPolicy := range approvalPolicies {
	// 	switch approvalPolicy.ConditionType {
	// 	case "project":
	// 		if approvalPolicy.ConditionValue == request.Project {
	// 			return &approvalPolicy, nil
	// 		}
	// 	case "user":
	// 		conditionValue, err := strconv.Atoi(approvalPolicy.ConditionValue)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		if uint(conditionValue) == request.UserID {
	// 			return &approvalPolicy, nil
	// 		}
	// 	case "category":
	// 		conditionValue, err := strconv.Atoi(approvalPolicy.ConditionValue)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		if request.CategoryID == uint(conditionValue) {
	// 			return &approvalPolicy, nil
	// 		}
	// 	case "amount":
	// 		if isAmountConditionMet(approvalPolicy.ConditionValue, request.Amount) {
	// 			return &approvalPolicy, nil
	// 		}
	// 	}
	// }
	return &approvalPolicy, nil
}

// func isAmountConditionMet(condition string, amount float64) bool {
// 	condition = strings.TrimSpace(condition) // Remove unnecessary spaces
// 	var operator string
// 	var value float64

// 	if strings.HasPrefix(condition, ">=") || strings.HasPrefix(condition, "<=") {
// 		operator = condition[:2]
// 		value, _ = strconv.ParseFloat(strings.TrimSpace(condition[2:]), 64)
// 	} else {
// 		operator = condition[:1]
// 		value, _ = strconv.ParseFloat(strings.TrimSpace(condition[1:]), 64)
// 	}

// 	switch operator {
// 	case ">":
// 		return amount > value
// 	case "<":
// 		return amount < value
// 	case "=":
// 		return amount == value
// 	case "<=":
// 		return amount <= value
// 	case ">=":
// 		return amount >= value
// 	}
// 	return false
// }

func (r *ExpenseRequestsRepo) GetExpenseRequestByApproverID(id uint) []models.ExpenseRequests {
	var expenseRequests []models.ExpenseRequests
	r.db.Joins("JOIN expense_approvals ON expense_approvals.request_id = expense_requests.id").
		Where("expense_approvals.approver_id = ?", id).
		Preload("Approvals").
		Preload("Approvals.Users", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, email") // Select specific fields for Users
		}).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, email") // Select specific fields for User
		}).
		Preload("Category").Order("expense_requests.created_at DESC").
		Find(&expenseRequests)
	return expenseRequests
}

func (r *ExpenseRequestsRepo) UpdateExpenseRequest(id uint, expenseRequest *models.ExpenseRequests) error {
	tx := r.db.Begin()

	var old_expenseRequest models.ExpenseRequests
	if err := tx.First(&old_expenseRequest, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	old_expenseRequest.IsSendToSQLACC = expenseRequest.IsSendToSQLACC
	old_expenseRequest.Description = expenseRequest.Description
	old_expenseRequest.PaymentMethod = expenseRequest.PaymentMethod
	old_expenseRequest.GLAccount = expenseRequest.GLAccount

	if old_expenseRequest.Attachment != expenseRequest.Attachment {
		if old_expenseRequest.Attachment != nil {
			oldFilePath := filepath.Join("uploads", *old_expenseRequest.Attachment)
			if _, err := os.Stat(oldFilePath); err == nil {
				os.Remove(oldFilePath)
			}
		}
		old_expenseRequest.Attachment = expenseRequest.Attachment
	}

	if old_expenseRequest.CategoryID != expenseRequest.CategoryID && old_expenseRequest.Project != expenseRequest.Project && old_expenseRequest.Amount != expenseRequest.Amount {

		old_expenseRequest.Project = expenseRequest.Project
		old_expenseRequest.CategoryID = expenseRequest.CategoryID
		old_expenseRequest.Amount = expenseRequest.Amount
		old_expenseRequest.Description = expenseRequest.Description
		old_expenseRequest.CurrentApproverLevel = 1

		if err := tx.Save(&old_expenseRequest).Error; err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Where("request_id = ?", old_expenseRequest.ID).Delete(&models.ExpenseApprovals{}).Error; err != nil {
			tx.Rollback()
			return err
		}

		var requestUser models.Users
		tx.Where("id = ?", expenseRequest.UserID).First(&requestUser)

		approvalPolicy, err := r.FindHighestPolicy(expenseRequest, *requestUser.DepartmentID)
		if err != nil {
			tx.Rollback()
			return err
		}

		var approvalPoliciesUsers []models.ApprovalPoliciesUsers
		tx.Preload("Approver").Where("approval_policy_id = ?", approvalPolicy.ID).Order("level ASC").Find(&approvalPoliciesUsers)

		if len(approvalPoliciesUsers) == 0 {
			tx.Rollback()
			return fmt.Errorf("no approver users found")
		}

		for i, approverPolicyUser := range approvalPoliciesUsers {
			fmt.Println("approver", approverPolicyUser, i)
			expenseApprovals := models.ExpenseApprovals{
				RequestID:  old_expenseRequest.ID,
				ApproverID: approverPolicyUser.UserID,
				Level:      approverPolicyUser.Level,
				Status:     "pending",
				IsFinal:    i == len(approvalPoliciesUsers)-1,
			}
			if err := tx.Create(&expenseApprovals).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	if err := tx.Save(&old_expenseRequest).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *ExpenseRequestsRepo) DeleteExpenseRequest(id uint) error {
	tx := r.db.Begin()
	if err := tx.Where("request_id = ?", id).Delete(&models.ExpenseApprovals{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("id = ?", id).Delete(&models.ExpenseRequests{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
