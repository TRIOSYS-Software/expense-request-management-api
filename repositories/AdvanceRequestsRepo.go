package repositories

import (
	"fmt"
	"shwetaik-expense-management-api/dtos"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/notifications"
	"shwetaik-expense-management-api/storage"
	"time"

	firebase "firebase.google.com/go/v4"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AdvanceRequestsRepo struct {
	db               *gorm.DB
	notificationRepo *NotificationRepo
	deviceTokenRepo  *DeviceTokenRepo
	notifier         *notifier
	attachments      storage.AttachmentStore
}

func NewAdvanceRequestsRepo(db *gorm.DB, firebaseApp *firebase.App, attachments storage.AttachmentStore) *AdvanceRequestsRepo {
	notificationRepo := NewNotificationRepo(db, firebaseApp)
	deviceTokenRepo := NewDeviceTokenRepo(db)
	return &AdvanceRequestsRepo{
		db:               db,
		notificationRepo: notificationRepo,
		deviceTokenRepo:  deviceTokenRepo,
		notifier:         newNotifier(notificationRepo, deviceTokenRepo),
		attachments:      attachments,
	}
}

func applyAdvanceFilters(db *gorm.DB, filter *dtos.AdvanceRequestFilterDTO) *gorm.DB {
	if filter == nil {
		return db
	}
	return applyRequestListFilters(db, "advance_requests", &requestListFilter{
		StartDate: filter.StartDate,
		EndDate:   filter.EndDate,
		Search:    filter.Search,
		MinAmount: filter.MinAmount,
		MaxAmount: filter.MaxAmount,
	})
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

	_ = FillAdvanceBalances(r.db, advanceRequests)
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
		Preload("ExpenseRequests", func(db *gorm.DB) *gorm.DB {
			return db.Order("expense_requests.created_at DESC")
		}).
		First(&advanceRequest, id).Error
	if err != nil {
		return &advanceRequest, err
	}
	single := []models.AdvanceRequests{advanceRequest}
	if ferr := FillAdvanceBalances(r.db, single); ferr == nil {
		advanceRequest.RemainingBalance = single[0].RemainingBalance
		advanceRequest.SettledAmount = single[0].SettledAmount
	}
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

	_ = FillAdvanceBalances(r.db, advanceRequests)
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

	_ = FillAdvanceBalances(r.db, advanceRequests)
	return advanceRequests, total
}

func (r *AdvanceRequestsRepo) GetSelectableAdvanceRequests(userID uint) ([]models.AdvanceRequests, error) {
	var advanceRequests []models.AdvanceRequests
	err := r.db.
		Where("user_id = ? AND status = 'approved'", userID).
		Preload("Projects").
		Preload("GLAccounts").
		Preload("PaymentMethods", func(db *gorm.DB) *gorm.DB { return db.Select("CODE, DESCRIPTION") }).
		Order("created_at DESC").
		Find(&advanceRequests).Error
	if err != nil {
		return nil, err
	}

	// One grouped query for every balance, rather than one query per advance.
	if err := FillAdvanceBalances(r.db, advanceRequests); err != nil {
		return nil, err
	}

	selectable := make([]models.AdvanceRequests, 0, len(advanceRequests))
	for i := range advanceRequests {
		// FillAdvanceBalances already snaps sub-1-Kyat dust to 0, so only advances
		// with a real (>= 1 Kyat) remainder stay selectable.
		if advanceRequests[i].RemainingBalance >= settledThreshold {
			selectable = append(selectable, advanceRequests[i])
		}
	}
	return selectable, nil
}

// NotifyNewAdvanceRequest delivers the notifications a create produced. Call
// only after the transaction has committed.
func (r *AdvanceRequestsRepo) NotifyNewAdvanceRequest(requestID uint, pending []PendingNotification) {
	go func() {
		for _, n := range pending {
			r.notifier.dispatch(notifications.Notification{
				UserID:    n.UserID,
				RequestID: requestID,
				Message:   n.Message,
				Type:      n.Type,
				PushTitle: "New Advance Request",
				Kind:      notifications.Advance,
			})
		}
	}()
}

func (r *AdvanceRequestsRepo) findHighestAdvancePolicy(tx *gorm.DB, request *models.AdvanceRequests, departmentID uint) (*models.ApprovalPolicies, error) {
	policy, err := findMatchingPolicy(tx, "advance", departmentID, request.Project, request.Amount, request.GLAccount)
	if err != nil {
		return nil, fmt.Errorf("No advance approval policy found")
	}
	return policy, nil
}

// UpdateAdvanceRequestTx applies an update inside the caller's transaction.
func (r *AdvanceRequestsRepo) UpdateAdvanceRequestTx(id uint, advanceRequest *models.AdvanceRequests) error {
	tx := r.db

	var old models.AdvanceRequests
	if err := tx.First(&old, id).Error; err != nil {
		return err
	}

	if old.Status != "pending" {
		return fmt.Errorf("Only pending advance requests can be edited")
	}

	old.Description = advanceRequest.Description
	old.PaymentMethod = advanceRequest.PaymentMethod
	old.GLAccount = advanceRequest.GLAccount

	if old.Attachment != nil {
		if !advanceRequest.KeepLegacyAttachment {
			r.attachments.Remove(*old.Attachment)
			old.Attachment = nil
		}
	}

	var existingAttachments []models.AdvanceRequestAttachments
	if err := tx.Where("advance_request_id = ?", old.ID).Find(&existingAttachments).Error; err != nil {
		return err
	}

	keptIDsMap := make(map[uint]bool)
	for _, kid := range advanceRequest.KeptAttachmentIDs {
		keptIDsMap[kid] = true
	}

	for _, att := range existingAttachments {
		if !keptIDsMap[att.ID] {
			if err := tx.Unscoped().Delete(&att).Error; err != nil {
				return err
			}
			r.attachments.Remove(att.FilePath)
		}
	}

	if len(advanceRequest.Attachments) > 0 {
		for _, att := range advanceRequest.Attachments {
			att.AdvanceRequestID = old.ID
			if err := tx.Create(&att).Error; err != nil {
				return err
			}
		}
	}

	if old.Project != advanceRequest.Project || old.Amount != advanceRequest.Amount {
		old.Project = advanceRequest.Project
		old.Amount = advanceRequest.Amount
		old.CurrentApproverLevel = 1

		if err := tx.Save(&old).Error; err != nil {
			return err
		}

		if err := tx.Where("request_id = ?", old.ID).Delete(&models.AdvanceApprovals{}).Error; err != nil {
			return err
		}

		var requestUser models.Users
		tx.Where("id = ?", advanceRequest.UserID).First(&requestUser)

		approvalPolicy, err := r.findHighestAdvancePolicy(tx, advanceRequest, *requestUser.DepartmentID)
		if err != nil {
			return err
		}

		var approvalPoliciesUsers []models.ApprovalPoliciesUsers
		tx.Preload("Approver").Where("approval_policy_id = ?", approvalPolicy.ID).Order("level ASC").Find(&approvalPoliciesUsers)

		if len(approvalPoliciesUsers) == 0 {
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
				return err
			}
		}
	}

	if err := tx.Save(&old).Error; err != nil {
		return err
	}

	return nil
}

// CloseAdvanceRequestTx closes an advance inside the caller's transaction and
// returns the requester notification to send once it commits.
func (r *AdvanceRequestsRepo) CloseAdvanceRequestTx(id uint, actorUserID uint, comment *string) (*PendingNotification, error) {
	var ar models.AdvanceRequests
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("User").First(&ar, id).Error; err != nil {
		return nil, err
	}

	var actor models.Users
	if err := r.db.Preload("Roles").First(&actor, actorUserID).Error; err != nil {
		return nil, err
	}
	isAdmin := actor.Roles != nil && actor.Roles.IsAdmin
	if !isAdmin && ar.UserID != actorUserID {
		return nil, fmt.Errorf("Only the requester or an admin can close this advance request")
	}

	if ar.Status != "approved" {
		return nil, fmt.Errorf("Only approved advance requests can be closed")
	}

	var pendingCount int64
	if err := r.db.Model(&models.ExpenseRequests{}).
		Where("advance_request_id = ? AND status = ?", id, "pending").
		Count(&pendingCount).Error; err != nil {
		return nil, err
	}
	if pendingCount > 0 {
		return nil, fmt.Errorf("Cannot close advance request: linked expense requests are still pending")
	}

	now := time.Now()
	ar.Status = "closed"
	ar.ClosureComment = comment
	ar.ClosedByUserID = &actorUserID
	ar.ClosedAt = &now

	if err := r.db.Save(&ar).Error; err != nil {
		return nil, err
	}

	return &PendingNotification{
		UserID:  ar.UserID,
		Message: fmt.Sprintf("Your advance request (#%d - '%s') has been manually closed.", ar.ID, ar.Description),
		Type:    "advance_closed",
	}, nil
}

// NotifyAdvanceClosed delivers the close notification. Call only after commit.
func (r *AdvanceRequestsRepo) NotifyAdvanceClosed(requestID uint, n *PendingNotification) {
	if n == nil {
		return
	}
	go r.notifier.dispatch(notifications.Notification{
		UserID:    n.UserID,
		RequestID: requestID,
		Message:   n.Message,
		Type:      n.Type,
		PushTitle: "Advance Request Closed",
		Kind:      notifications.Advance,
	})
}

func (r *AdvanceRequestsRepo) UpdateSendToSQLACCStatus(id uint, status bool) error {
	return r.db.Model(&models.AdvanceRequests{}).Where("id = ?", id).Update("is_send_to_sqlacc", status).Error
}

// SoftDeleteAdvanceRequest marks the AR (and its attachments) as deleted and cascade-soft-deletes
// every linked expense request (and their attachments). Records remain in the DB for audit but are
// filtered out of default listings by GORM's soft-delete behavior.
func (r *AdvanceRequestsRepo) SoftDeleteAdvanceRequest(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var ar models.AdvanceRequests
		if err := tx.First(&ar, id).Error; err != nil {
			return err
		}

		var linkedERs []models.ExpenseRequests
		if err := tx.Where("advance_request_id = ?", id).Find(&linkedERs).Error; err != nil {
			return err
		}
		for _, er := range linkedERs {
			if err := tx.Where("expense_request_id = ?", er.ID).Delete(&models.ExpenseRequestAttachments{}).Error; err != nil {
				return err
			}
			if err := tx.Delete(&models.ExpenseRequests{}, er.ID).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("advance_request_id = ?", id).Delete(&models.AdvanceRequestAttachments{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.AdvanceRequests{}, id).Error
	})
}

func (r *AdvanceRequestsRepo) CountLinkedExpenseRequests(id uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.ExpenseRequests{}).Where("advance_request_id = ?", id).Count(&count).Error
	return count, err
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
	if err := tx.Unscoped().Where("advance_request_id = ?", id).Delete(&models.AdvanceRequestAttachments{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Unscoped().Where("id = ?", id).Delete(&models.AdvanceRequests{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
