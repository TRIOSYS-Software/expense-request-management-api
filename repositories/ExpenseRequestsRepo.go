package repositories

import (
	"fmt"
	"shwetaik-expense-management-api/models"
	"strconv"
	"strings"

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
	r.db.Preload("Approvals.Users", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, email, role_id, department_id")
	}).Preload("Approvals.Users.Roles").Preload("Approvals.Users.Departments").Preload("Category").Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).Find(&expenseRequests)
	return expenseRequests
}

func (r *ExpenseRequestsRepo) GetExpenseRequestByID(id uint) (*models.ExpenseRequests, error) {
	var expenseRequest models.ExpenseRequests
	err := r.db.First(&expenseRequest, id).Error
	return &expenseRequest, err
}

func (r *ExpenseRequestsRepo) GetExpenseRequestsByUserID(id uint) []models.ExpenseRequests {
	var expenseRequests []models.ExpenseRequests
	r.db.Where("user_id = ?", id).Preload("Approvals.Users", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, email")
	}).Preload("Category").Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).Find(&expenseRequests)
	return expenseRequests
}

func (r *ExpenseRequestsRepo) CreateExpenseRequest(expenseRequest *models.ExpenseRequests) error {
	tx := r.db.Begin()
	if err := tx.Create(expenseRequest).Error; err != nil {
		tx.Rollback()
		return err
	}
	var requestUser models.Users
	tx.Where("id = ?", expenseRequest.UserID).First(&requestUser)

	if expenseRequest.Approvers != nil {
		approvers := strings.Split(*expenseRequest.Approvers, ",")
		for i, approver := range approvers {
			var approverUser models.Users
			tx.Where("id", approver).First(&approverUser)
			expenseApprovals := models.ExpenseApprovals{
				RequestID:  expenseRequest.ID,
				ApproverID: approverUser.ID,
				Level:      uint(i) + 1,
				Status:     "pending",
			}
			if err := tx.Create(&expenseApprovals).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	} else {
		approvalPolicy, err := r.FindHighestPolicy(expenseRequest, requestUser.DepartmentID)
		if err != nil {
			tx.Rollback()
			return err
		}

		fmt.Println(approvalPolicy)

		var approverUser []models.Users
		tx.Model(&approvalPolicy).Association("Approver").Find(&approverUser)

		if len(approverUser) == 0 {
			tx.Rollback()
			return fmt.Errorf("no approver users found")
		}
		for i, approver := range approverUser {
			expenseApprovals := models.ExpenseApprovals{
				RequestID:  expenseRequest.ID,
				ApproverID: approver.ID,
				Level:      uint(i) + 1,
				Status:     "pending",
			}
			if err := tx.Create(&expenseApprovals).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}
	return tx.Commit().Error
}

func (r *ExpenseRequestsRepo) FindHighestPolicy(request *models.ExpenseRequests, departmentID uint) (*models.ApprovalPolicies, error) {
	var approvalPolicies []models.ApprovalPolicies
	err := r.db.Where("department_id = ? OR department_id IS NULL", departmentID).Order("priority DESC").Find(&approvalPolicies).Error
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	if len(approvalPolicies) == 0 {
		return nil, fmt.Errorf("no approval policies found")
	}
	for _, approvalPolicy := range approvalPolicies {
		switch approvalPolicy.ConditionType {
		case "project":
			if request.Project != nil && approvalPolicy.ConditionValue == *request.Project {
				return &approvalPolicy, nil
			}
		case "user":
			conditionValue, err := strconv.Atoi(approvalPolicy.ConditionValue)
			if err != nil {
				return nil, err
			}
			if uint(conditionValue) == request.UserID {
				return &approvalPolicy, nil
			}
		case "category":
			conditionValue, err := strconv.Atoi(approvalPolicy.ConditionValue)
			if err != nil {
				return nil, err
			}
			if request.CategoryID != nil && *request.CategoryID == uint(conditionValue) {
				return &approvalPolicy, nil
			}
		case "amount":
			if isAmountConditionMet(approvalPolicy.ConditionValue, request.Amount) {
				return &approvalPolicy, nil
			}
		}
	}
	return nil, fmt.Errorf("no approval policies found")
}

func isAmountConditionMet(condition string, amount float64) bool {
	condition = strings.TrimSpace(condition) // Remove unnecessary spaces
	var operator string
	var value float64
	fmt.Println(condition)

	if strings.HasPrefix(condition, ">=") || strings.HasPrefix(condition, "<=") {
		operator = condition[:2]
		value, _ = strconv.ParseFloat(strings.TrimSpace(condition[2:]), 64)
		fmt.Println(operator, value)
	} else {
		operator = condition[:1]
		value, _ = strconv.ParseFloat(strings.TrimSpace(condition[1:]), 64)
		fmt.Println(operator, value)
	}

	switch operator {
	case ">":
		return amount > value
	case "<":
		return amount < value
	case "=":
		return amount == value
	case "<=":
		return amount <= value
	case ">=":
		return amount >= value
	}
	return false
}

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
		Preload("Category").
		Find(&expenseRequests)
	return expenseRequests
}
