package repositories

import (
	"fmt"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/utilities"
	"time"

	firebase "firebase.google.com/go/v4"

	"gorm.io/gorm"
)

type ExpenseApprovalsRepo struct {
	db               *gorm.DB
	notificationRepo *NotificationRepo
	deviceTokenRepo  *DeviceTokenRepo
}

func NewExpenseApprovalsRepo(db *gorm.DB, firebaeApp *firebase.App) *ExpenseApprovalsRepo {
	return &ExpenseApprovalsRepo{
		db:               db,
		notificationRepo: NewNotificationRepo(db, firebaeApp),
		deviceTokenRepo:  NewDeviceTokenRepo(db),
	}
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
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
		}
	}()

	var expenseApprovalToUpdate models.ExpenseApprovals
	if err := tx.Preload("Users").Where("id = ?", id).
		First(&expenseApprovalToUpdate).Error; err != nil {
		tx.Rollback()
		return err
	}

	var expenseRequest models.ExpenseRequests
	if err := tx.Preload("Approvals").Preload("User").Preload("GLAccounts").
		Where("id = ?", expenseApprovalToUpdate.RequestID).First(&expenseRequest).Error; err != nil {
		tx.Rollback()
		return err
	}

	originalRequestCreatorID := expenseRequest.UserID
	originalRequestCreatorName := expenseRequest.User.Name
	expenseDescription := expenseRequest.Description

	expenseApprovalToUpdate.Status = expenseApproval.Status
	expenseApprovalToUpdate.Comments = expenseApproval.Comments
	expenseApprovalToUpdate.ApprovalDate = expenseApproval.ApprovalDate

	if err := tx.Save(&expenseApprovalToUpdate).Error; err != nil {
		tx.Rollback()
		return err
	}

	if expenseApproval.Status == "rejected" {
		expenseRequest.Status = "rejected"
		expenseRequest.CurrentApproverLevel = expenseApprovalToUpdate.Level

		msg := fmt.Sprintf(
			"Your expense request (#%d - '%s') has been REJECTED by %s (Level %d). Reason: %s",
			expenseRequest.ID,
			expenseDescription,
			expenseApprovalToUpdate.Users.Name,
			expenseApprovalToUpdate.Level,
			*expenseApproval.Comments,
		)

		r.sendSingleNotification(tx, originalRequestCreatorID, expenseRequest.ID, msg, "rejected")

		if err := tx.Save(&expenseRequest).Error; err != nil {
			tx.Rollback()
			return err
		}

		return tx.Commit().Error
	}

	if expenseApproval.Status == "approved" {

		expenseRequest.CurrentApproverLevel = expenseApprovalToUpdate.Level + 1

		var nextLevelApprovals []models.ExpenseApprovals
		if err := tx.Where("request_id = ? AND level = ?",
			expenseRequest.ID,
			expenseRequest.CurrentApproverLevel,
		).
			Preload("Users").
			Find(&nextLevelApprovals).Error; err != nil {

			tx.Rollback()
			return err
		}

		if len(nextLevelApprovals) > 0 {

			for _, approver := range nextLevelApprovals {

				msg := fmt.Sprintf(
					"You have a new expense request (#%d - '%s') from %s to approve. (Level %d)",
					expenseRequest.ID,
					expenseDescription,
					originalRequestCreatorName,
					expenseRequest.CurrentApproverLevel,
				)

				r.sendSingleNotification(tx, approver.ApproverID, expenseRequest.ID, msg, "pending_approval")
			}

		} else {
			expenseRequest.Status = "approved"

			msg := fmt.Sprintf(
				"Your expense request (#%d - '%s') has been fully APPROVED!",
				expenseRequest.ID, expenseDescription,
			)

			r.sendSingleNotification(tx, originalRequestCreatorID, expenseRequest.ID, msg, "approved_final")
		}
	}

	if err := tx.Save(&expenseRequest).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *ExpenseApprovalsRepo) sendSingleNotification(
	tx *gorm.DB,
	userID uint,
	expenseID uint,
	message string,
	notificationType string,
) {

	notification := &models.Notification{
		UserID:    userID,
		ExpenseID: expenseID,
		Message:   message,
		Type:      notificationType,
		IsRead:    false,
	}

	_ = r.notificationRepo.CreateNotification(notification)

	tokens, err := r.deviceTokenRepo.GetTokensByUserID(userID)
	if err == nil && len(tokens) > 0 {
		data := map[string]string{
			"expenseId": fmt.Sprintf("%d", expenseID),
			"type":      notificationType,
		}
		r.notificationRepo.SendPushNotification(tokens, "Expense Request", message, data)
	}

	go utilities.SendWebSocketMessage(
		userID,
		utilities.WebSocketMessagePayload{
			ID:        notification.ID,
			Message:   message,
			Type:      notificationType,
			ExpenseID: expenseID,
			IsRead:    false,
			CreatedAt: notification.CreatedAt.Format(time.RFC3339),
		},
	)
}
