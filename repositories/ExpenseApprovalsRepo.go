package repositories

import (
	"fmt"
	"log"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/utilities"
	"time"

	firebase "firebase.google.com/go/v4"

	"gorm.io/gorm"
)

type ExpenseApprovalsRepo struct {
	db               *gorm.DB
	notificationRepo *NotificationRepo
	deviceTokenRepo *DeviceTokenRepo
}

func NewExpenseApprovalsRepo(db *gorm.DB, firebaeApp *firebase.App) *ExpenseApprovalsRepo {
	return &ExpenseApprovalsRepo{
		db:               db,
		notificationRepo: NewNotificationRepo(db, firebaeApp),
		deviceTokenRepo: NewDeviceTokenRepo(db),
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
			tx.Rollback()
		}
	}()

	var expenseApprovalToUpdate models.ExpenseApprovals
	if err := tx.Preload("Users").Where("id = ?", id).First(&expenseApprovalToUpdate).Error; err != nil {
		tx.Rollback()
		return err
	}

	var expenseRequest models.ExpenseRequests
	tx.Preload("Approvals").Preload("User").Preload("GLAccounts").Where("id = ?", expenseApprovalToUpdate.RequestID).First(&expenseRequest)

	originalRequestCreatorID := expenseRequest.UserID
	originalRequestCreatorName := expenseRequest.User.Name
	expenseDescription := expenseRequest.Description

	var notificationMessage string
	var notificationType string
	var targetUserID uint

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

		notificationMessage = fmt.Sprintf(
			"Your expense request (#%d - '%s')has been REJECTED by %s (Level %d). Reason: %s",
			expenseRequest.ID,
			expenseDescription,
			expenseApprovalToUpdate.Users.Name,
			expenseApprovalToUpdate.Level,
			*expenseApproval.Comments,
		)
		notificationType = "rejected"
		targetUserID = originalRequestCreatorID

	} else if expenseApproval.Status == "approved" {
		expenseRequest.CurrentApproverLevel = expenseApprovalToUpdate.Level + 1

		var nextApproverApproval models.ExpenseApprovals
		err := tx.Where("request_id = ? AND level = ?", expenseRequest.ID, expenseRequest.CurrentApproverLevel).
			Preload("Users").
			First(&nextApproverApproval).Error

		if err == nil {
			notificationMessage = fmt.Sprintf(
				"You have a new expense request (#%d - '%s') from %s to approve. (Current Level: %d)",
				expenseRequest.ID,
				expenseDescription,
				originalRequestCreatorName,
				expenseRequest.CurrentApproverLevel,
			)
			notificationType = "pending_approval"
			targetUserID = nextApproverApproval.ApproverID
		} else if err == gorm.ErrRecordNotFound {
			expenseRequest.Status = "approved"

			notificationMessage = fmt.Sprintf(
				"Your expense request (#%d - '%s') has been fully APPROVED!",
				expenseRequest.ID, expenseDescription,
			)
			notificationType = "approved_final"
			targetUserID = originalRequestCreatorID
		} else {
			tx.Rollback()
			return fmt.Errorf("error finding next approver: %w", err)
		}
	} else {
		return tx.Commit().Error
	}

	if err := tx.Save(&expenseRequest).Error; err != nil {
		tx.Rollback()
		return err
	}

	if targetUserID != 0 && notificationMessage != "" && notificationType != "" {
		notification := &models.Notification{
			UserID:    targetUserID,
			ExpenseID: expenseRequest.ID,
			Message:   notificationMessage,
			Type:      notificationType,
			IsRead:    false,
		}
		tokens, err := r.deviceTokenRepo.GetTokensByUserID(targetUserID)
			if err != nil {
				log.Printf("Error fetching device tokens for user %d: %v", targetUserID, err)
			} else if len(tokens) > 0 {
				data := map[string]string{
					"expenseId": fmt.Sprintf("%d", expenseRequest.ID),
					"type":      notificationType,
				}
				r.notificationRepo.SendPushNotification(tokens, "New Expense Request", notificationMessage, data)
			}

		if err := r.notificationRepo.CreateNotification(notification); err != nil {
			fmt.Printf("Error saving notification to DB for user %d: %v\n", targetUserID, err)
		}

		go utilities.SendWebSocketMessage(
			targetUserID,
			utilities.WebSocketMessagePayload{
				ID:        notification.ID,
				Message:   notificationMessage,
				Type:      notificationType,
				ExpenseID: expenseRequest.ID,
				IsRead:    false,
				CreatedAt: notification.CreatedAt.Format(time.RFC3339),
			},
		)
	}

	return tx.Commit().Error
}