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
	"strconv"
	"time"

	firebase "firebase.google.com/go/v4"

	"gorm.io/gorm"
)

type AdvanceRequestsRepo struct {
	db               *gorm.DB
	notificationRepo *NotificationRepo
	deviceTokenRepo  *DeviceTokenRepo
	uploadDir        string
}

func NewAdvanceRequestsRepo(db *gorm.DB, firebaseApp *firebase.App) *AdvanceRequestsRepo {
	return &AdvanceRequestsRepo{
		db:               db,
		notificationRepo: NewNotificationRepo(db, firebaseApp),
		deviceTokenRepo:  NewDeviceTokenRepo(db),
		uploadDir:        configs.Envs.UploadDir,
	}
}

func applyAdvanceFilters(db *gorm.DB, filter *dtos.AdvanceRequestFilterDTO) *gorm.DB {
	if filter == nil {
		return db
	}
	if filter.StartDate != "" && filter.EndDate != "" {
		db = db.Where("DATE(advance_requests.date_submitted) BETWEEN ? AND ?", filter.StartDate, filter.EndDate)
	} else if filter.StartDate != "" {
		db = db.Where("DATE(advance_requests.date_submitted) >= ?", filter.StartDate)
	} else if filter.EndDate != "" {
		db = db.Where("DATE(advance_requests.date_submitted) <= ?", filter.EndDate)
	}
	if filter.Search != "" {
		db = db.Joins("LEFT JOIN users search_users ON search_users.id = advance_requests.user_id").
			Joins("LEFT JOIN projects search_projects ON search_projects.CODE = advance_requests.project")
		searchPattern := "%" + filter.Search + "%"
		if idVal, err := strconv.Atoi(filter.Search); err == nil {
			db = db.Where("(advance_requests.id = ? OR search_users.name LIKE ? OR search_projects.CODE LIKE ? OR search_projects.DESCRIPTION LIKE ?)", idVal, searchPattern, searchPattern, searchPattern)
		} else {
			db = db.Where("(search_users.name LIKE ? OR search_projects.CODE LIKE ? OR search_projects.DESCRIPTION LIKE ?)", searchPattern, searchPattern, searchPattern)
		}
	}
	if filter.MinAmount != nil {
		db = db.Where("advance_requests.amount >= ?", *filter.MinAmount)
	}
	if filter.MaxAmount != nil {
		db = db.Where("advance_requests.amount <= ?", *filter.MaxAmount)
	}
	return db
}

func (r *AdvanceRequestsRepo) GetAdvanceRequests(approverID uint, filter *dtos.AdvanceRequestFilterDTO) ([]models.AdvanceRequests, int64) {
	var advanceRequests []models.AdvanceRequests
	var total int64

	db := r.db.Model(&models.AdvanceRequests{})

	if filter != nil && filter.NeedMyApproval {
		db = db.Joins("JOIN advance_approvals ON advance_approvals.request_id = advance_requests.id").
			Where("advance_approvals.approver_id = ?", approverID).
			Where("advance_requests.status = 'pending'").
			Where("advance_approvals.level = advance_requests.current_approver_level")
	} else {
		if filter != nil && filter.IncludedAsApprover {
			db = db.Joins("JOIN advance_approvals ON advance_approvals.request_id = advance_requests.id").
				Where("advance_approvals.approver_id = ?", approverID)
		}

		if filter != nil && filter.Status != "" {
			db = db.Where("advance_requests.status = ?", filter.Status)
		}
	}

	db = applyAdvanceFilters(db, filter)
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
		Order("advance_requests.created_at DESC").
		Offset(filter.Offset()).Limit(filter.Limit()).
		Find(&advanceRequests)

	return advanceRequests, total
}

func (r *AdvanceRequestsRepo) GetAdvanceRequestByID(id uint) (*models.AdvanceRequests, error) {
	var advanceRequest models.AdvanceRequests
	err := r.db.Preload("Projects").
		Preload("GLAccounts").
		Preload("PaymentMethods", func(db *gorm.DB) *gorm.DB { return db.Select("CODE, DESCRIPTION") }).
		Preload("Approvals").
		Preload("Approvals.Users", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).
		Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).
		Preload("Attachments").
		Preload("ExpenseRequest").
		First(&advanceRequest, id).Error
	return &advanceRequest, err
}

func (r *AdvanceRequestsRepo) GetAdvanceRequestsByUserID(id uint, filter *dtos.AdvanceRequestFilterDTO) ([]models.AdvanceRequests, int64) {
	var advanceRequests []models.AdvanceRequests
	var total int64
	db := r.db.Model(&models.AdvanceRequests{}).Where("advance_requests.user_id = ?", id)

	if filter != nil && filter.Status != "" {
		db = db.Where("advance_requests.status = ?", filter.Status)
	}

	db = applyAdvanceFilters(db, filter)
	db.Count(&total)

	db.Preload("Approvals.Users", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, email")
	}).
		Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id, name, email") }).
		Preload("Projects").
		Preload("GLAccounts").
		Preload("PaymentMethods", func(db *gorm.DB) *gorm.DB { return db.Select("CODE, DESCRIPTION") }).
		Preload("Attachments").
		Order("advance_requests.created_at DESC").
		Offset(filter.Offset()).Limit(filter.Limit()).
		Find(&advanceRequests)
	return advanceRequests, total
}

func (r *AdvanceRequestsRepo) GetAdvanceRequestByApproverID(id uint, filter *dtos.AdvanceRequestFilterDTO) ([]models.AdvanceRequests, int64) {
	var advanceRequests []models.AdvanceRequests
	var total int64

	db := r.db.Model(&models.AdvanceRequests{})

	if filter != nil && filter.NeedMyApproval {
		// "Awaiting": I must be the active approver at the current level.
		db = db.Joins("JOIN advance_approvals ON advance_approvals.request_id = advance_requests.id").
			Where("advance_approvals.approver_id = ?", id).
			Where("advance_requests.status = 'pending'").
			Where("advance_approvals.level = advance_requests.current_approver_level")
	} else {
		// Default: show ARs where I'm the requester OR I appear in the approval chain.
		db = db.Where(
			"advance_requests.user_id = ? OR EXISTS (SELECT 1 FROM advance_approvals WHERE advance_approvals.request_id = advance_requests.id AND advance_approvals.approver_id = ?)",
			id, id,
		)

		if filter != nil && filter.Status != "" {
			db = db.Where("advance_requests.status = ?", filter.Status)
		}
	}

	db = applyAdvanceFilters(db, filter)
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
		Order("advance_requests.created_at DESC").
		Offset(filter.Offset()).Limit(filter.Limit()).
		Find(&advanceRequests)
	return advanceRequests, total
}

func (r *AdvanceRequestsRepo) GetAdvanceRequestsSummary(filters map[string]any) (dtos.AdvanceRequestSummary, error) {
	var advanceRequests []models.AdvanceRequests
	var summary dtos.AdvanceRequestSummary

	db := r.db.Model(&models.AdvanceRequests{}).Preload("Approvals")
	if filters["user_id"] != nil && filters["approver_id"] != nil {
		db = db.Joins("LEFT JOIN advance_approvals ON advance_approvals.request_id = advance_requests.id").
			Where("(advance_requests.user_id = ? OR advance_approvals.approver_id = ?)", filters["user_id"], filters["approver_id"]).
			Group("advance_requests.id")
	} else if filters["user_id"] != nil {
		db = db.Where("advance_requests.user_id = ?", filters["user_id"])
	} else if filters["approver_id"] != nil {
		db = db.Joins("JOIN advance_approvals ON advance_approvals.request_id = advance_requests.id").
			Where("advance_approvals.approver_id = ?", filters["approver_id"]).
			Group("advance_requests.id")
	}

	if filters["status"] != nil {
		db = db.Where("advance_requests.status = ?", filters["status"].(string))
	}

	if filters["start_date"] != nil && filters["end_date"] != nil {
		db = db.Where("DATE(date_submitted) BETWEEN ? AND ?", filters["start_date"], filters["end_date"])
		summary.DailyTotal = make(map[string]dtos.DailyBreakdown)
	}

	db.Find(&advanceRequests)

	for _, ar := range advanceRequests {
		summary.TotalAmount = summary.TotalAmount + ar.Amount
		switch ar.Status {
		case "pending":
			summary.Pending++
		case "approved":
			summary.Approved++
		case "rejected":
			summary.Rejected++
		case "completed":
			summary.Completed++
		}

		if filters["start_date"] != nil && filters["end_date"] != nil {
			date := ar.DateSubmitted.Format("2006-01-02")
			entry := summary.DailyTotal[date]
			switch ar.Status {
			case "approved", "completed":
				entry.Approved += ar.Amount
			case "pending":
				entry.Pending += ar.Amount
			case "rejected":
				entry.Rejected += ar.Amount
			}
			summary.DailyTotal[date] = entry
		}
	}
	summary.Total = len(advanceRequests)
	return summary, nil
}

// GetSelectableAdvanceRequests returns Approved ARs for the given user that have no
// expense_requests currently linking them with status in ('pending','approved').
func (r *AdvanceRequestsRepo) GetSelectableAdvanceRequests(userID uint) ([]models.AdvanceRequests, error) {
	var advanceRequests []models.AdvanceRequests
	err := r.db.
		Where("user_id = ? AND status = 'approved'", userID).
		Where("NOT EXISTS (SELECT 1 FROM expense_requests er WHERE er.advance_request_id = advance_requests.id AND er.status IN ('pending','approved'))").
		Preload("Projects").
		Preload("GLAccounts").
		Preload("PaymentMethods", func(db *gorm.DB) *gorm.DB { return db.Select("CODE, DESCRIPTION") }).
		Order("created_at DESC").
		Find(&advanceRequests).Error
	return advanceRequests, err
}

func (r *AdvanceRequestsRepo) CreateAdvanceRequest(advanceRequest *models.AdvanceRequests) error {
	tx := r.db.Begin()
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
			log.Printf("PANIC recovered in CreateAdvanceRequest: %v", rec)
		}
	}()

	var requestUser models.Users
	err := tx.Preload("Roles").Preload("Departments").Where("id = ?", advanceRequest.UserID).First(&requestUser).Error
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("User with ID %d not found", advanceRequest.UserID)
		}
		return fmt.Errorf("Failed to retrieve user: %w", err)
	}

	if requestUser.DepartmentID == nil {
		tx.Rollback()
		return fmt.Errorf("User (ID %d - %s) has no department assigned", requestUser.ID, requestUser.Name)
	}

	var userRoleName string
	if requestUser.Roles != nil {
		userRoleName = requestUser.Roles.Name
	} else {
		userRoleName = "Unknown Role"
	}

	approvalPolicy, err := r.findHighestAdvancePolicy(tx, advanceRequest, *requestUser.DepartmentID)
	if err != nil {
		tx.Rollback()
		return err
	}

	var approvalPoliciesUsers []models.ApprovalPoliciesUsers
	if err := tx.Preload("Approver").Where("approval_policy_id = ?", approvalPolicy.ID).Order("level ASC").Find(&approvalPoliciesUsers).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("Failed to retrieve approver users: %w", err)
	}

	if len(approvalPoliciesUsers) == 0 {
		tx.Rollback()
		return fmt.Errorf("No approver users found")
	}

	if err := tx.Create(advanceRequest).Error; err != nil {
		tx.Rollback()
		return err
	}

	type pendingNotification struct {
		userID  uint
		message string
		nType   string
	}
	var notifications []pendingNotification

	for i, approverPolicyUser := range approvalPoliciesUsers {
		advanceApproval := models.AdvanceApprovals{
			RequestID:  advanceRequest.ID,
			ApproverID: approverPolicyUser.UserID,
			Level:      approverPolicyUser.Level,
			Status:     "pending",
			IsFinal:    i == len(approvalPoliciesUsers)-1,
		}
		if err := tx.Create(&advanceApproval).Error; err != nil {
			tx.Rollback()
			return err
		}

		if approverPolicyUser.Level == advanceRequest.CurrentApproverLevel {
			message := fmt.Sprintf(
				"%s (%s) has created a new advance request (#%d) for your approval. Amount: $%.2f",
				requestUser.Name,
				userRoleName,
				advanceRequest.ID,
				advanceRequest.Amount,
			)
			notifications = append(notifications, pendingNotification{
				userID:  approverPolicyUser.UserID,
				message: message,
				nType:   "advance_new_request",
			})
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	go func() {
		for _, n := range notifications {
			notification := &models.Notification{
				UserID:    n.userID,
				ExpenseID: advanceRequest.ID,
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
					"advanceId": fmt.Sprintf("%d", advanceRequest.ID),
					"type":      n.nType,
				}
				r.notificationRepo.SendPushNotification(tokens, "New Advance Request", n.message, data)
			}

			utilities.SendWebSocketMessage(
				n.userID,
				utilities.WebSocketMessagePayload{
					ID:        notification.ID,
					Message:   n.message,
					Type:      n.nType,
					ExpenseID: advanceRequest.ID,
					IsRead:    false,
					CreatedAt: notification.CreatedAt.Format(time.RFC3339),
				},
			)
		}
	}()

	return nil
}

func (r *AdvanceRequestsRepo) findHighestAdvancePolicy(tx *gorm.DB, request *models.AdvanceRequests, departmentID uint) (*models.ApprovalPolicies, error) {
	var approvalPolicy models.ApprovalPolicies
	err := tx.Where(
		"policy_type = 'advance' AND (department_id = ? OR department_id IS NULL) AND project = ? AND ? BETWEEN min_amount AND max_amount AND (NOT EXISTS (SELECT 1 FROM approval_policy_gl_accounts WHERE approval_policy_id = approval_policies.id) OR EXISTS (SELECT 1 FROM approval_policy_gl_accounts WHERE approval_policy_id = approval_policies.id AND gl_account_dockey = CAST(? AS UNSIGNED)))",
		departmentID, request.Project, request.Amount, request.GLAccount,
	).Order("NOT EXISTS (SELECT 1 FROM approval_policy_gl_accounts WHERE approval_policy_id = approval_policies.id) ASC").First(&approvalPolicy).Error
	if err != nil {
		return nil, fmt.Errorf("No advance approval policy found")
	}
	return &approvalPolicy, nil
}

func (r *AdvanceRequestsRepo) UpdateAdvanceRequest(id uint, advanceRequest *models.AdvanceRequests) error {
	tx := r.db.Begin()

	var old models.AdvanceRequests
	if err := tx.First(&old, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	if old.Status != "pending" {
		tx.Rollback()
		return fmt.Errorf("Only pending advance requests can be edited")
	}

	old.Description = advanceRequest.Description
	old.PaymentMethod = advanceRequest.PaymentMethod
	old.GLAccount = advanceRequest.GLAccount

	if old.Attachment != nil {
		if !advanceRequest.KeepLegacyAttachment {
			oldFilePath := filepath.Join(r.uploadDir, *old.Attachment)
			if _, err := os.Stat(oldFilePath); err == nil {
				os.Remove(oldFilePath)
			}
			old.Attachment = nil
		}
	}

	var existingAttachments []models.AdvanceRequestAttachments
	if err := tx.Where("advance_request_id = ?", old.ID).Find(&existingAttachments).Error; err != nil {
		tx.Rollback()
		return err
	}

	keptIDsMap := make(map[uint]bool)
	for _, kid := range advanceRequest.KeptAttachmentIDs {
		keptIDsMap[kid] = true
	}

	for _, att := range existingAttachments {
		if !keptIDsMap[att.ID] {
			if err := tx.Delete(&att).Error; err != nil {
				tx.Rollback()
				return err
			}
			filePath := filepath.Join(r.uploadDir, att.FilePath)
			if _, err := os.Stat(filePath); err == nil {
				os.Remove(filePath)
			}
		}
	}

	if len(advanceRequest.Attachments) > 0 {
		for _, att := range advanceRequest.Attachments {
			att.AdvanceRequestID = old.ID
			if err := tx.Create(&att).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	if old.Project != advanceRequest.Project || old.Amount != advanceRequest.Amount {
		old.Project = advanceRequest.Project
		old.Amount = advanceRequest.Amount
		old.CurrentApproverLevel = 1

		if err := tx.Save(&old).Error; err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Where("request_id = ?", old.ID).Delete(&models.AdvanceApprovals{}).Error; err != nil {
			tx.Rollback()
			return err
		}

		var requestUser models.Users
		tx.Where("id = ?", advanceRequest.UserID).First(&requestUser)

		approvalPolicy, err := r.findHighestAdvancePolicy(tx, advanceRequest, *requestUser.DepartmentID)
		if err != nil {
			tx.Rollback()
			return err
		}

		var approvalPoliciesUsers []models.ApprovalPoliciesUsers
		tx.Preload("Approver").Where("approval_policy_id = ?", approvalPolicy.ID).Order("level ASC").Find(&approvalPoliciesUsers)

		if len(approvalPoliciesUsers) == 0 {
			tx.Rollback()
			return fmt.Errorf("No approver users found")
		}

		for i, approverPolicyUser := range approvalPoliciesUsers {
			advanceApproval := models.AdvanceApprovals{
				RequestID:  old.ID,
				ApproverID: approverPolicyUser.UserID,
				Level:      approverPolicyUser.Level,
				Status:     "pending",
				IsFinal:    i == len(approvalPoliciesUsers)-1,
			}
			if err := tx.Create(&advanceApproval).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	if err := tx.Save(&old).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *AdvanceRequestsRepo) DeleteAdvanceRequest(id uint) error {
	tx := r.db.Begin()

	var ar models.AdvanceRequests
	if err := tx.First(&ar, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	if ar.Status != "pending" {
		tx.Rollback()
		return fmt.Errorf("Only pending advance requests can be deleted")
	}

	var linkedCount int64
	if err := tx.Model(&models.ExpenseRequests{}).Where("advance_request_id = ?", id).Count(&linkedCount).Error; err != nil {
		tx.Rollback()
		return err
	}
	if linkedCount > 0 {
		tx.Rollback()
		return fmt.Errorf("Cannot delete advance request: it is referenced by one or more expense requests")
	}

	if err := tx.Where("request_id = ?", id).Delete(&models.AdvanceApprovals{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("advance_request_id = ?", id).Delete(&models.AdvanceRequestAttachments{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("id = ?", id).Delete(&models.AdvanceRequests{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
