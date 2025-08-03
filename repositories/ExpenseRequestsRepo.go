package repositories

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/utilities"
	"time"

	"gorm.io/gorm"
)

type ExpenseRequestsRepo struct {
	db               *gorm.DB
	notificationRepo *NotificationRepo
}

func NewExpenseRequestsRepo(db *gorm.DB) *ExpenseRequestsRepo {
	return &ExpenseRequestsRepo{
		db:               db,
		notificationRepo: NewNotificationRepo(db),
	}
}

func (r *ExpenseRequestsRepo) GetExpenseRequests() []models.ExpenseRequests {
	var expenseRequests []models.ExpenseRequests
	r.db.Preload("Projects").Preload("GLAccounts").
		Preload("PaymentMethods", func(db *gorm.DB) *gorm.DB { return db.Select("CODE, DESCRIPTION") }).
		Preload("Approvals").Preload("Approvals.Users", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, email, role_id, department_id")
	}).
		Preload("Approvals.Users.Roles").Preload("Approvals.Users.Departments").
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
		Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).
		First(&expenseRequest, id).Error
	return &expenseRequest, err
}

func (r *ExpenseRequestsRepo) GetExpenseRequestsByUserID(id uint) []models.ExpenseRequests {
	var expenseRequests []models.ExpenseRequests
	r.db.Where("user_id = ?", id).Preload("Approvals.Users", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, email")
	}).
		Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).
		Preload("Projects").
		Preload("GLAccounts").
		Preload("PaymentMethods", func(db *gorm.DB) *gorm.DB { return db.Select("CODE, DESCRIPTION") }).
		Order("expense_requests.created_at DESC").Find(&expenseRequests)
	return expenseRequests
}

func (r *ExpenseRequestsRepo) GetExpenseRequestsSummary(filters map[string]any) (dtos.ExpenseRequestSummary, error) {
	var expenseRequests []models.ExpenseRequests
	var summary dtos.ExpenseRequestSummary

	db := r.db.Model(&models.ExpenseRequests{}).Preload("Approvals")
	if filters["user_id"] != nil {
		db = db.Where("user_id = ?", filters["user_id"])
	}
	if filters["status"] != nil {
		db = db.Where("expense_requests.status = ?", filters["status"].(string))
	}

	if filters["start_date"] != nil && filters["end_date"] != nil {
		db = db.Where("date_submitted BETWEEN ? AND ?", filters["start_date"], filters["end_date"])
		summary.DailyTotal = make(map[string]float64)
	}

	if filters["amount"] != nil {
		db = db.Where("amount = ?", filters["amount"])
	}

	if filters["approver_id"] != nil {
		db = db.Joins("JOIN expense_approvals ON expense_approvals.request_id = expense_requests.id").
			Where("expense_approvals.approver_id = ?", filters["approver_id"])
	}

	db.Find(&expenseRequests)

	for _, expenseRequest := range expenseRequests {
		summary.TotalAmount = summary.TotalAmount + expenseRequest.Amount
		if expenseRequest.Status == "pending" {
			summary.Pending = summary.Pending + 1
		} else if expenseRequest.Status == "approved" {
			summary.Approved = summary.Approved + 1
		} else if expenseRequest.Status == "rejected" {
			summary.Rejected = summary.Rejected + 1
		}

		if filters["start_date"] != nil && filters["end_date"] != nil {
			date := expenseRequest.DateSubmitted.Format("2006-01-02")
			summary.DailyTotal[date] = summary.DailyTotal[date] + expenseRequest.Amount
		}

	}
	return summary, nil
}

func (r *ExpenseRequestsRepo) CreateExpenseRequest(expenseRequest *models.ExpenseRequests) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("PANIC recovered in CreateExpenseRequest: %v", r)
		}
	}()

	if err := tx.Create(expenseRequest).Error; err != nil {
		tx.Rollback()
		return err
	}

	var requestUser models.Users
	err := tx.Preload("Roles").Preload("Departments").Where("id = ?", expenseRequest.UserID).First(&requestUser).Error
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("requesting user with ID %d not found", expenseRequest.UserID)
		}
		return fmt.Errorf("failed to retrieve requesting user: %w", err)
	}

	if requestUser.DepartmentID == nil {
		tx.Rollback()
		return fmt.Errorf("requesting user (ID %d - %s) has no department assigned", requestUser.ID, requestUser.Name)
	}

	// Safely get user's role name, as Roles is a *Roles
	var userRoleName string
	if requestUser.Roles != nil {
		userRoleName = requestUser.Roles.Name
	} else {
		userRoleName = "Unknown Role"
		log.Printf("WARN: User %d (%s) has no role assigned or role not found for role_id: %d", requestUser.ID, requestUser.Name, requestUser.RoleID)
	}

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

		if i == 0 {
			message := fmt.Sprintf(
				"%s (%s) has created a new expense request (#%d) for your approval. Amount: $%.2f",
				requestUser.Name,
				userRoleName,
				expenseRequest.ID,
				expenseRequest.Amount,
			)
			notificationType := "new_request"

			notification := &models.Notification{
				UserID:    approverPolicyUser.UserID,
				ExpenseID: expenseRequest.ID,
				Message:   message,
				Type:      notificationType,
				IsRead:    false,
			}
			if err := r.notificationRepo.CreateNotification(notification); err != nil {
				log.Printf("Error saving notification to DB for user %d: %v", approverPolicyUser.UserID, err)
			}

			go utilities.SendWebSocketMessage(
				approverPolicyUser.UserID,
				utilities.WebSocketMessagePayload{
					ID:        notification.ID,
					Message:   message,
					Type:      notificationType,
					ExpenseID: expenseRequest.ID,
					IsRead:    false,
					CreatedAt: notification.CreatedAt.Format(time.RFC3339),
				},
			)
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
		Preload("Projects").
		Preload("GLAccounts").
		Preload("PaymentMethods", func(db *gorm.DB) *gorm.DB { return db.Select("CODE, DESCRIPTION") }).
		Preload("Approvals").
		Preload("Approvals.Users", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, email") // Select specific fields for Users
		}).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, email") // Select specific fields for User
		}).
		Order("expense_requests.created_at DESC").
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

	if expenseRequest.Attachment != nil && *expenseRequest.Attachment != "" {
		if old_expenseRequest.Attachment != nil {
			oldFilePath := filepath.Join("uploads", *old_expenseRequest.Attachment)
			if _, err := os.Stat(oldFilePath); err == nil {
				os.Remove(oldFilePath)
			}
		}
		old_expenseRequest.Attachment = expenseRequest.Attachment
	}

	if old_expenseRequest.Project != expenseRequest.Project || old_expenseRequest.Amount != expenseRequest.Amount {

		old_expenseRequest.Project = expenseRequest.Project
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
