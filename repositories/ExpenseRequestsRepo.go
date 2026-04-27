package repositories

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"shwetaik-expense-management-api/configs"
	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/utilities"
	"sort"
	"strconv"
	"time"

	firebase "firebase.google.com/go/v4"

	"gorm.io/gorm"
)

type ExpenseRequestsRepo struct {
	db               *gorm.DB
	notificationRepo *NotificationRepo
	deviceTokenRepo  *DeviceTokenRepo
	uploadDir        string
}

func NewExpenseRequestsRepo(db *gorm.DB, firebaseApp *firebase.App) *ExpenseRequestsRepo {
	return &ExpenseRequestsRepo{
		db:               db,
		notificationRepo: NewNotificationRepo(db, firebaseApp),
		deviceTokenRepo:  NewDeviceTokenRepo(db),
		uploadDir:        configs.Envs.UploadDir,
	}
}

func applyFilters(db *gorm.DB, filter *dtos.ExpenseRequestFilterDTO) *gorm.DB {
	if filter == nil {
		return db
	}
	if filter.StartDate != "" && filter.EndDate != "" {
		db = db.Where("DATE(expense_requests.date_submitted) BETWEEN ? AND ?", filter.StartDate, filter.EndDate)
	} else if filter.StartDate != "" {
		db = db.Where("DATE(expense_requests.date_submitted) >= ?", filter.StartDate)
	} else if filter.EndDate != "" {
		db = db.Where("DATE(expense_requests.date_submitted) <= ?", filter.EndDate)
	}
	if filter.Search != "" {
		db = db.Joins("LEFT JOIN users search_users ON search_users.id = expense_requests.user_id").
			Joins("LEFT JOIN projects search_projects ON search_projects.CODE = expense_requests.project")
		searchPattern := "%" + filter.Search + "%"
		if idVal, err := strconv.Atoi(filter.Search); err == nil {
			db = db.Where("(expense_requests.id = ? OR search_users.name LIKE ? OR search_projects.CODE LIKE ? OR search_projects.DESCRIPTION LIKE ?)", idVal, searchPattern, searchPattern, searchPattern)
		} else {
			db = db.Where("(search_users.name LIKE ? OR search_projects.CODE LIKE ? OR search_projects.DESCRIPTION LIKE ?)", searchPattern, searchPattern, searchPattern)
		}
	}
	if filter.MinAmount != nil {
		db = db.Where("expense_requests.amount >= ?", *filter.MinAmount)
	}
	if filter.MaxAmount != nil {
		db = db.Where("expense_requests.amount <= ?", *filter.MaxAmount)
	}
	return db
}

func (r *ExpenseRequestsRepo) GetExpenseRequests(approverID uint, filter *dtos.ExpenseRequestFilterDTO) ([]models.ExpenseRequests, int64) {
	var expenseRequests []models.ExpenseRequests
	var total int64

	db := r.db.Model(&models.ExpenseRequests{})

	if filter != nil && filter.NeedMyApproval {
		db = db.Joins("JOIN expense_approvals ON expense_approvals.request_id = expense_requests.id").
			Where("expense_approvals.approver_id = ?", approverID).
			Where("expense_requests.status = 'pending'").
			Where("expense_approvals.level = expense_requests.current_approver_level")
	} else {
		if filter != nil && filter.IncludedAsApprover {
			db = db.Joins("JOIN expense_approvals ON expense_approvals.request_id = expense_requests.id").
				Where("expense_approvals.approver_id = ?", approverID)
		}

		if filter != nil && filter.Status != "" {
			db = db.Where("expense_requests.status = ?", filter.Status)
		} else if filter == nil || !filter.IncludedAsApprover {
			// Admin All: every approved + every rejected + all pending
			db = db.Where(`(
					expense_requests.status = 'approved'
					OR expense_requests.status = 'rejected'
					OR (expense_requests.status = 'pending')
				)`)
		}
	}

	db = applyFilters(db, filter)
	db.Session(&gorm.Session{}).Count(&total)

	db.Session(&gorm.Session{}).
		Preload("Projects").Preload("GLAccounts").
		Preload("PaymentMethods", func(db *gorm.DB) *gorm.DB { return db.Select("CODE, DESCRIPTION") }).
		Preload("Approvals").Preload("Approvals.Users", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, email, role_id, department_id")
	}).
		Preload("Approvals.Users.Roles").Preload("Approvals.Users.Departments").
		Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).
		Preload("Attachments").
		Order("expense_requests.created_at DESC").
		Offset(filter.Offset()).Limit(filter.Limit()).
		Find(&expenseRequests)

	return expenseRequests, total
}

func (r *ExpenseRequestsRepo) GetExpenseRequestByID(id uint) (*models.ExpenseRequests, error) {
	var expenseRequest models.ExpenseRequests
	err := r.db.Preload("Projects").
		Preload("GLAccounts").
		Preload("PaymentMethods", func(db *gorm.DB) *gorm.DB { return db.Select("CODE, DESCRIPTION") }).
		Preload("Approvals").
		Preload("Approvals.Users", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).
		Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).
		Preload("Attachments").
		First(&expenseRequest, id).Error
	return &expenseRequest, err
}

func (r *ExpenseRequestsRepo) GetExpenseRequestsByUserID(id uint, filter *dtos.ExpenseRequestFilterDTO) ([]models.ExpenseRequests, int64) {
	var expenseRequests []models.ExpenseRequests
	var total int64
	db := r.db.Model(&models.ExpenseRequests{}).Where("expense_requests.user_id = ?", id)

	// Requester status filter: based on expense_requests.status
	if filter != nil && filter.Status != "" {
		db = db.Where("expense_requests.status = ?", filter.Status)
	}

	db = applyFilters(db, filter)
	db.Count(&total)

	db.Preload("Approvals.Users", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, email")
	}).
		Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).
		Preload("Projects").
		Preload("GLAccounts").
		Preload("PaymentMethods", func(db *gorm.DB) *gorm.DB { return db.Select("CODE, DESCRIPTION") }).
		Preload("Attachments").
		Order("expense_requests.created_at DESC").
		Offset(filter.Offset()).Limit(filter.Limit()).
		Find(&expenseRequests)
	return expenseRequests, total
}

func (r *ExpenseRequestsRepo) GetExpenseRequestsSummary(filters map[string]any) (dtos.ExpenseRequestSummary, error) {
	var expenseRequests []models.ExpenseRequests
	var summary dtos.ExpenseRequestSummary

	db := r.db.Model(&models.ExpenseRequests{}).Preload("Approvals")
	if filters["user_id"] != nil && filters["approver_id"] != nil {
		db = db.Joins("LEFT JOIN expense_approvals ON expense_approvals.request_id = expense_requests.id").
			Where("(expense_requests.user_id = ? OR expense_approvals.approver_id = ?)", filters["user_id"], filters["approver_id"]).
			Group("expense_requests.id")
	} else if filters["user_id"] != nil {
		db = db.Where("expense_requests.user_id = ?", filters["user_id"])
	} else if filters["approver_id"] != nil {
		db = db.Joins("JOIN expense_approvals ON expense_approvals.request_id = expense_requests.id").
			Where("expense_approvals.approver_id = ?", filters["approver_id"]).
			Group("expense_requests.id")
	}

	if filters["status"] != nil {
		db = db.Where("expense_requests.status = ?", filters["status"].(string))
	}

	if filters["start_date"] != nil && filters["end_date"] != nil {
		db = db.Where("DATE(date_submitted) BETWEEN ? AND ?", filters["start_date"], filters["end_date"])
		summary.DailyTotal = make(map[string]dtos.DailyBreakdown)
	}

	if filters["amount"] != nil {
		db = db.Where("amount = ?", filters["amount"])
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
			entry := summary.DailyTotal[date]
			switch expenseRequest.Status {
			case "approved":
				entry.Approved += expenseRequest.Amount
			case "pending":
				entry.Pending += expenseRequest.Amount
			case "rejected":
				entry.Rejected += expenseRequest.Amount
			}
			summary.DailyTotal[date] = entry
		}

	}
	summary.Total = len(expenseRequests)
	return summary, nil
}

func (r *ExpenseRequestsRepo) CreateExpenseRequest(expenseRequest *models.ExpenseRequests) error {
	tx := r.db.Begin()
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
			log.Printf("PANIC recovered in CreateExpenseRequest: %v", rec)
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

	approvalPolicy, err := r.findHighestPolicy(tx, expenseRequest, *requestUser.DepartmentID)
	if err != nil {
		tx.Rollback()
		return err
	}

	var approvalPoliciesUsers []models.ApprovalPoliciesUsers
	if err := tx.Preload("Approver").Where("approval_policy_id = ?", approvalPolicy.ID).Order("level ASC").Find(&approvalPoliciesUsers).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to retrieve approver users: %w", err)
	}

	if len(approvalPoliciesUsers) == 0 {
		tx.Rollback()
		return fmt.Errorf("no approver users found")
	}

	type pendingNotification struct {
		userID  uint
		message string
		nType   string
	}
	var notifications []pendingNotification

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

		if approverPolicyUser.Level == expenseRequest.CurrentApproverLevel {
			message := fmt.Sprintf(
				"%s (%s) has created a new expense request (#%d) for your approval. Amount: $%.2f",
				requestUser.Name,
				userRoleName,
				expenseRequest.ID,
				expenseRequest.Amount,
			)
			notifications = append(notifications, pendingNotification{
				userID:  approverPolicyUser.UserID,
				message: message,
				nType:   "new_request",
			})
		}
	}

	if len(notifications) == 0 {
		log.Printf("WARN: No approver matched CurrentApproverLevel %d for expense request %d", expenseRequest.CurrentApproverLevel, expenseRequest.ID)
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// Send notifications in background — don't block the HTTP response
	go func() {
		for _, n := range notifications {
			notification := &models.Notification{
				UserID:    n.userID,
				ExpenseID: expenseRequest.ID,
				Message:   n.message,
				Type:      n.nType,
				IsRead:    false,
			}

			if err := r.notificationRepo.CreateNotification(notification); err != nil {
				log.Printf("Error saving notification to DB for user %d: %v", n.userID, err)
			}

			tokens, err := r.deviceTokenRepo.GetTokensByUserID(n.userID)
			if err != nil {
				log.Printf("Error fetching device tokens for user %d: %v", n.userID, err)
			} else if len(tokens) > 0 {
				data := map[string]string{
					"expenseId": fmt.Sprintf("%d", expenseRequest.ID),
					"type":      n.nType,
				}
				r.notificationRepo.SendPushNotification(tokens, "New Expense Request", n.message, data)
			}

			utilities.SendWebSocketMessage(
				n.userID,
				utilities.WebSocketMessagePayload{
					ID:        notification.ID,
					Message:   n.message,
					Type:      n.nType,
					ExpenseID: expenseRequest.ID,
					IsRead:    false,
					CreatedAt: notification.CreatedAt.Format(time.RFC3339),
				},
			)
		}
	}()

	return nil
}

func (r *ExpenseRequestsRepo) findHighestPolicy(tx *gorm.DB, request *models.ExpenseRequests, departmentID uint) (*models.ApprovalPolicies, error) {
	var approvalPolicy models.ApprovalPolicies
	err := tx.Where(
		"(department_id = ? OR department_id IS NULL) AND project = ? AND ? BETWEEN min_amount AND max_amount AND (NOT EXISTS (SELECT 1 FROM approval_policy_gl_accounts WHERE approval_policy_id = approval_policies.id) OR EXISTS (SELECT 1 FROM approval_policy_gl_accounts WHERE approval_policy_id = approval_policies.id AND gl_account_dockey = CAST(? AS UNSIGNED)))",
		departmentID, request.Project, request.Amount, request.GLAccount,
	).Order("NOT EXISTS (SELECT 1 FROM approval_policy_gl_accounts WHERE approval_policy_id = approval_policies.id) ASC").First(&approvalPolicy).Error
	if err != nil {
		return nil, fmt.Errorf("no approval policy found")
	}
	return &approvalPolicy, nil
}

func (r *ExpenseRequestsRepo) GetExpenseRequestByApproverID(id uint, filter *dtos.ExpenseRequestFilterDTO) ([]models.ExpenseRequests, int64) {
	var expenseRequests []models.ExpenseRequests
	var total int64
	db := r.db.Model(&models.ExpenseRequests{}).
		Joins("JOIN expense_approvals ON expense_approvals.request_id = expense_requests.id").
		Where("expense_approvals.approver_id = ?", id)

	if filter != nil && filter.NeedMyApproval {
		db = db.Where("expense_requests.status = 'pending'").
			Where("expense_approvals.level = expense_requests.current_approver_level")
	} else if filter != nil && filter.Status != "" {
		db = db.Where("expense_requests.status = ?", filter.Status)
		if filter.Status == "rejected" {
			db = db.Where("expense_approvals.level <= expense_requests.current_approver_level")
		}
		// Pending: show all in chain, no level restriction
		// Approved: show all in chain, no level restriction
	} else {
		db = db.Where(`(
			(expense_requests.status = 'pending')
			OR (expense_requests.status = 'approved')
			OR (expense_requests.status = 'rejected' AND expense_approvals.level <= expense_requests.current_approver_level)
		)`)
	}

	db = applyFilters(db, filter)
	db.Session(&gorm.Session{}).Count(&total)

	db.Session(&gorm.Session{}).
		Preload("Projects").
		Preload("GLAccounts").
		Preload("PaymentMethods", func(db *gorm.DB) *gorm.DB { return db.Select("CODE, DESCRIPTION") }).
		Preload("Approvals").
		Preload("Approvals.Users", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, email")
		}).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, name, email")
		}).
		Preload("Attachments").
		Order("expense_requests.created_at DESC").
		Offset(filter.Offset()).Limit(filter.Limit()).
		Find(&expenseRequests)
	return expenseRequests, total
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

	// Handle Legacy Attachment Retention
	if old_expenseRequest.Attachment != nil {
		if !expenseRequest.KeepLegacyAttachment {
			// User wants to remove the legacy attachment
			oldFilePath := filepath.Join(r.uploadDir, *old_expenseRequest.Attachment)
			if _, err := os.Stat(oldFilePath); err == nil {
				os.Remove(oldFilePath)
			}
			old_expenseRequest.Attachment = nil
		}
	}

	var existingAttachments []models.ExpenseRequestAttachments
	if err := tx.Where("expense_request_id = ?", old_expenseRequest.ID).Find(&existingAttachments).Error; err != nil {
		tx.Rollback()
		return err
	}

	keptIDsMap := make(map[uint]bool)
	for _, id := range expenseRequest.KeptAttachmentIDs {
		keptIDsMap[id] = true
	}

	for _, att := range existingAttachments {
		if !keptIDsMap[att.ID] {
			// Delete from DB
			if err := tx.Delete(&att).Error; err != nil {
				tx.Rollback()
				return err
			}
			// Delete from Disk
			filePath := filepath.Join(r.uploadDir, att.FilePath)
			if _, err := os.Stat(filePath); err == nil {
				os.Remove(filePath)
			}
		}
	}

	// Handle New Attachments
	if len(expenseRequest.Attachments) > 0 {
		for _, att := range expenseRequest.Attachments {
			att.ExpenseRequestID = old_expenseRequest.ID
			if err := tx.Create(&att).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
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

		approvalPolicy, err := r.findHighestPolicy(tx, expenseRequest, *requestUser.DepartmentID)
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

func (r *ExpenseRequestsRepo) UpdateSendToSQLACCStatus(id uint, status bool) error {
	return r.db.Model(&models.ExpenseRequests{}).Where("id = ?", id).Update("is_send_to_sqlacc", status).Error
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

func (r *ExpenseRequestsRepo) GetAnalytics(filters map[string]any) (dtos.AnalyticsResponse, error) {
	var result dtos.AnalyticsResponse

	// Build base approved-request scope reusable across sub-queries.
	baseApproved := func() *gorm.DB {
		q := r.db.Table("expense_requests er").Where("er.status = 'approved'")
		if filters["start_date"] != nil && filters["end_date"] != nil {
			q = q.Where("DATE(er.date_submitted) BETWEEN ? AND ?", filters["start_date"], filters["end_date"])
		}
		if filters["user_id"] != nil {
			q = q.Where("er.user_id = ?", filters["user_id"])
		}
		if filters["approver_id"] != nil {
			q = q.Joins("JOIN expense_approvals _af ON _af.request_id = er.id AND _af.approver_id = ?", filters["approver_id"])
		}
		return q
	}

	// Spend by project.
	type spendRow struct {
		Code   string
		Name   string
		Amount float64
	}
	var projectRows []spendRow
	baseApproved().
		Select("er.project AS code, COALESCE(p.DESCRIPTION, er.project) AS name, SUM(er.amount) AS amount").
		Joins("LEFT JOIN projects p ON p.CODE = er.project").
		Group("er.project, p.DESCRIPTION").
		Order("amount DESC").
		Limit(10).
		Scan(&projectRows)
	for _, row := range projectRows {
		result.SpendByProject = append(result.SpendByProject, dtos.AnalyticsSpendItem{
			Name: row.Name, Code: row.Code, Amount: row.Amount,
		})
	}

	// Spend by GL account.
	var glRows []spendRow
	baseApproved().
		Select("er.gl_account AS code, COALESCE(ga.DESCRIPTION, er.gl_account) AS name, SUM(er.amount) AS amount").
		Joins("LEFT JOIN gl_accs ga ON ga.DOCKEY = er.gl_account").
		Group("er.gl_account, ga.DESCRIPTION").
		Order("amount DESC").
		Limit(10).
		Scan(&glRows)
	for _, row := range glRows {
		result.SpendByGL = append(result.SpendByGL, dtos.AnalyticsSpendItem{
			Name: row.Name, Code: row.Code, Amount: row.Amount,
		})
	}

	// Spend by payment method.
	var pmRows []spendRow
	baseApproved().
		Select("er.payment_method AS code, COALESCE(pm.DESCRIPTION, er.payment_method) AS name, SUM(er.amount) AS amount").
		Joins("LEFT JOIN payment_methods pm ON pm.CODE = er.payment_method").
		Group("er.payment_method, pm.DESCRIPTION").
		Order("amount DESC").
		Scan(&pmRows)
	for _, row := range pmRows {
		result.SpendByPayment = append(result.SpendByPayment, dtos.AnalyticsSpendItem{
			Name: row.Name, Code: row.Code, Amount: row.Amount,
		})
	}

	// Top submitters (by approved spend).
	type lbRow struct {
		ID     uint
		Name   string
		Team   string
		Amount float64
		Count  int
	}
	var submitterRows []lbRow
	baseApproved().
		Select("u.id AS id, u.name AS name, COALESCE(d.name, '') AS team, SUM(er.amount) AS amount, COUNT(er.id) AS count").
		Joins("JOIN users u ON u.id = er.user_id").
		Joins("LEFT JOIN departments d ON d.id = u.department_id").
		Group("u.id, u.name, d.name").
		Order("amount DESC").
		Limit(10).
		Scan(&submitterRows)
	for _, row := range submitterRows {
		result.TopSubmitters = append(result.TopSubmitters, dtos.AnalyticsLeaderboardItem{
			ID: row.ID, Name: row.Name, Team: row.Team, Amount: row.Amount, Count: row.Count,
		})
	}

	// Top approvers (by approved amount across their approvals).
	approversQ := r.db.Table("expense_approvals ea").
		Select("u.id AS id, u.name AS name, COALESCE(d.name, '') AS team, SUM(er.amount) AS amount, COUNT(ea.id) AS count").
		Joins("JOIN users u ON u.id = ea.approver_id").
		Joins("LEFT JOIN departments d ON d.id = u.department_id").
		Joins("JOIN expense_requests er ON er.id = ea.request_id").
		Where("ea.status = 'approved'")
	if filters["start_date"] != nil && filters["end_date"] != nil {
		approversQ = approversQ.Where("DATE(er.date_submitted) BETWEEN ? AND ?", filters["start_date"], filters["end_date"])
	}
	if filters["user_id"] != nil {
		approversQ = approversQ.Where("er.user_id = ?", filters["user_id"])
	}
	var approverRows []lbRow
	approversQ.
		Group("u.id, u.name, d.name").
		Order("amount DESC").
		Limit(10).
		Scan(&approverRows)
	for _, row := range approverRows {
		result.TopApprovers = append(result.TopApprovers, dtos.AnalyticsLeaderboardItem{
			ID: row.ID, Name: row.Name, Team: row.Team, Amount: row.Amount, Count: row.Count,
		})
	}

	// Recent activity: merge latest submissions + latest approval actions.
	type activityEvent struct {
		Kind   string
		Who    string
		Code   string
		Text   string
		Amount float64
		At     time.Time
	}
	var events []activityEvent

	// Recent expense request submissions.
	type submissionRow struct {
		ID            uint
		Amount        float64
		SubmitterName string
		DateSubmitted time.Time
		Status        string
	}
	var submissions []submissionRow
	submissionQ := r.db.Table("expense_requests er").
		Select("er.id, er.amount, u.name AS submitter_name, er.date_submitted, er.status").
		Joins("JOIN users u ON u.id = er.user_id").
		Order("er.date_submitted DESC").
		Limit(20)
	if filters["start_date"] != nil && filters["end_date"] != nil {
		submissionQ = submissionQ.Where("DATE(er.date_submitted) BETWEEN ? AND ?", filters["start_date"], filters["end_date"])
	}
	if filters["user_id"] != nil {
		submissionQ = submissionQ.Where("er.user_id = ?", filters["user_id"])
	}
	if filters["approver_id"] != nil {
		submissionQ = submissionQ.
			Joins("JOIN expense_approvals _sa ON _sa.request_id = er.id AND _sa.approver_id = ?", filters["approver_id"])
	}
	submissionQ.Scan(&submissions)
	for _, s := range submissions {
		kind := s.Status
		text := "submitted"
		if s.Status == "approved" {
			text = "submitted (approved)"
		} else if s.Status == "rejected" {
			text = "submitted (rejected)"
		}
		events = append(events, activityEvent{
			Kind:   kind,
			Who:    s.SubmitterName,
			Code:   fmt.Sprintf("EXP-%05d", s.ID),
			Text:   text,
			Amount: s.Amount,
			At:     s.DateSubmitted,
		})
	}

	// Recent approval actions.
	type approvalActionRow struct {
		RequestID    uint
		Amount       float64
		ApproverName string
		ApprovalDate time.Time
		Status       string
	}
	var approvalActions []approvalActionRow
	approvalActionsQ := r.db.Table("expense_approvals ea").
		Select("ea.request_id, er.amount, u.name AS approver_name, ea.approval_date, ea.status").
		Joins("JOIN users u ON u.id = ea.approver_id").
		Joins("JOIN expense_requests er ON er.id = ea.request_id").
		Where("ea.status IN ('approved', 'rejected') AND ea.approval_date IS NOT NULL").
		Order("ea.approval_date DESC").
		Limit(20)
	if filters["start_date"] != nil && filters["end_date"] != nil {
		approvalActionsQ = approvalActionsQ.Where("DATE(er.date_submitted) BETWEEN ? AND ?", filters["start_date"], filters["end_date"])
	}
	if filters["user_id"] != nil {
		approvalActionsQ = approvalActionsQ.Where("er.user_id = ?", filters["user_id"])
	}
	if filters["approver_id"] != nil {
		approvalActionsQ = approvalActionsQ.Where("ea.approver_id = ?", filters["approver_id"])
	}
	approvalActionsQ.Scan(&approvalActions)
	for _, a := range approvalActions {
		text := "approved"
		if a.Status == "rejected" {
			text = "rejected"
		}
		events = append(events, activityEvent{
			Kind:   a.Status,
			Who:    a.ApproverName,
			Code:   fmt.Sprintf("EXP-%05d", a.RequestID),
			Text:   text,
			Amount: a.Amount,
			At:     a.ApprovalDate,
		})
	}

	// Sort all events by time DESC, deduplicate by Code+Kind, take top 15.
	sort.Slice(events, func(i, j int) bool {
		return events[i].At.After(events[j].At)
	})
	seen := make(map[string]bool)
	for _, ev := range events {
		key := ev.Code + ev.Kind
		if seen[key] {
			continue
		}
		seen[key] = true
		result.RecentActivity = append(result.RecentActivity, dtos.AnalyticsActivityItem{
			Kind:   ev.Kind,
			Who:    ev.Who,
			Code:   ev.Code,
			Text:   ev.Text,
			Amount: ev.Amount,
			When:   humanizeTime(ev.At),
		})
		if len(result.RecentActivity) >= 15 {
			break
		}
	}

	return result, nil
}

func humanizeTime(t time.Time) string {
	diff := time.Since(t)
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d min ago", mins)
	case diff < 24*time.Hour:
		hrs := int(diff.Hours())
		if hrs == 1 {
			return "1 hr ago"
		}
		return fmt.Sprintf("%d hr ago", hrs)
	default:
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}
