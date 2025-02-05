package repositories

import (
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
		// approvalPolicy, err := r.FindHighestPolicy(*expenseRequest)
		// if err != nil {
		// 	tx.Rollback()
		// 	return err
		// }
		// approverRoles := strings.Split(approvalPolicy.Approvers, ",")

		// for i, approverRole := range approverRoles {
		// 	var approverUser models.Users
		// 	tx.Where("role_id = ? AND department_id = ?", approverRole, requestUser.DepartmentID).First(&approverUser)
		// 	expenseApprovals := models.ExpenseApprovals{
		// 		RequestID:  expenseRequest.ID,
		// 		ApproverID: approverUser.ID,
		// 		Level:      uint(i) + 1,
		// 		Status:     "pending",
		// 	}
		// 	if err := tx.Create(&expenseApprovals).Error; err != nil {
		// 		tx.Rollback()
		// 		return err
		// 	}
		// }
	}

	return tx.Commit().Error
}

func (r *ExpenseRequestsRepo) FindHighestPolicy(request models.ExpenseRequests) (*models.ApprovalPolicies, error) {
	var approvalPolicies []models.ApprovalPolicies
	err := r.db.Where("department_id = ? ", request.UserID).Find(&approvalPolicies).Error
	if err != nil {
		return nil, err
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
			if uint(conditionValue) == *request.CategoryID {
				return &approvalPolicy, nil
			}
		case "amount":
			conditionValues := strings.Split(approvalPolicy.ConditionValue, " ")
			condition, value := conditionValues[0], conditionValues[1]
			amount, _ := strconv.Atoi(value)
			if condition == ">" && request.Amount > float64(amount) {
				return &approvalPolicy, nil
			} else if condition == "<" && request.Amount < float64(amount) {
				return &approvalPolicy, nil
			}
		}
	}
	return nil, nil
}
