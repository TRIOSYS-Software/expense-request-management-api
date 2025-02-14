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
	r.db.Find(&expenseRequests)
	return expenseRequests
}

func (r *ExpenseRequestsRepo) GetExpenseRequestByID(id uint) (*models.ExpenseRequests, error) {
	var expenseRequest models.ExpenseRequests
	err := r.db.First(&expenseRequest, id).Error
	return &expenseRequest, err
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
		approverRoles := strings.Split(approvalPolicy.ApproverRoles, ",")

		for i, approverRole := range approverRoles {
			var approverUser models.Users
			tx.Where("role_id = ?", approverRole).First(&approverUser)
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
	}

	return tx.Commit().Error
}

func (r *ExpenseRequestsRepo) FindHighestPolicy(request *models.ExpenseRequests, departmentID uint) (*models.ApprovalPolicies, error) {
	var approvalPolicies []models.ApprovalPolicies
	err := r.db.Where("department_id = ? OR department_id IS NULL", departmentID).Find(&approvalPolicies).Error
	if err != nil {
		return nil, err
	}

	if len(approvalPolicies) == 0 {
		return nil, fmt.Errorf("no approval policies found")
	}
	for _, approvalPolicy := range approvalPolicies {
		switch approvalPolicy.ConditionType {
		case "project":
			if approvalPolicy.ConditionValue == *request.Project {
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

	if strings.HasPrefix(condition, ">=") || strings.HasPrefix(condition, "<=") {
		operator = condition[:2]
		value, _ = strconv.ParseFloat(condition[2:], 64)
	} else {
		operator = condition[:1]
		value, _ = strconv.ParseFloat(condition[1:], 64)
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
