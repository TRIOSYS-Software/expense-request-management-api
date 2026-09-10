package repositories

import (
	"errors"
	"fmt"
	"log"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/notifications"
	"time"

	firebase "firebase.google.com/go/v4"

	"gorm.io/gorm"
)

type ExpenseApprovalsRepo struct {
	db               *gorm.DB
	notificationRepo *NotificationRepo
	deviceTokenRepo  *DeviceTokenRepo
	notifier         *notifier
}

func NewExpenseApprovalsRepo(db *gorm.DB, firebaeApp *firebase.App) *ExpenseApprovalsRepo {
	notificationRepo := NewNotificationRepo(db, firebaeApp)
	deviceTokenRepo := NewDeviceTokenRepo(db)
	return &ExpenseApprovalsRepo{
		db:               db,
		notificationRepo: notificationRepo,
		deviceTokenRepo:  deviceTokenRepo,
		notifier:         newNotifier(notificationRepo, deviceTokenRepo),
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

func (r *ExpenseApprovalsRepo) UpdateExpenseApproval(id uint, expenseApproval *models.ExpenseApprovals) (err error) {
	tx := r.db.Begin()
	defer func() {
		if rec := recover(); rec != nil {
			_ = tx.Rollback()
			err = fmt.Errorf("internal error in UpdateExpenseApproval: %v", rec)
			log.Printf("PANIC recovered in UpdateExpenseApproval: %v", rec)
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

	if expenseRequest.Status != "pending" {
		tx.Rollback()
		return errors.New("this expense request has already been finalized; please refresh")
	}
	if expenseApprovalToUpdate.Status != "pending" {
		tx.Rollback()
		return errors.New("this approval has already been processed; please refresh")
	}
	if expenseApprovalToUpdate.Level != expenseRequest.CurrentApproverLevel {
		tx.Rollback()
		return errors.New("this request has already advanced past your level; please refresh")
	}

	originalRequestCreatorID := expenseRequest.UserID
	originalRequestCreatorName := expenseRequest.User.Name
	expenseDescription := expenseRequest.Description

	expenseApprovalToUpdate.Status = expenseApproval.Status
	expenseApprovalToUpdate.Comments = expenseApproval.Comments

	if expenseApproval.Status == "approved" || expenseApproval.Status == "rejected" {
		now := time.Now()
		expenseApprovalToUpdate.ApprovalDate = &now
	}

	if err := tx.Save(&expenseApprovalToUpdate).Error; err != nil {
		tx.Rollback()
		return err
	}

	removedPerUser, err := r.notificationRepo.DeleteActionableForRequest(
		tx,
		expenseRequest.ID,
		[]string{"new_request", "pending_approval"},
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	if expenseApproval.Status == "rejected" {
		expenseRequest.Status = "rejected"
		expenseRequest.CurrentApproverLevel = expenseApprovalToUpdate.Level

		// Comments is nullable: a rejection may legitimately arrive without one.
		// Mirrors the guard AdvanceApprovalsRepo already applies.
		comment := ""
		if expenseApproval.Comments != nil {
			comment = *expenseApproval.Comments
		}

		msg := fmt.Sprintf(
			"Your expense request (#%d - '%s') has been REJECTED by %s (Level %d). Reason: %s",
			expenseRequest.ID,
			expenseDescription,
			expenseApprovalToUpdate.Users.Name,
			expenseApprovalToUpdate.Level,
			comment,
		)

		r.sendSingleNotification(tx, originalRequestCreatorID, expenseRequest.ID, msg, "rejected")

		if err := tx.Save(&expenseRequest).Error; err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit().Error; err != nil {
			return err
		}
		notifications.BroadcastRemovals(expenseRequest.ID, removedPerUser)
		return nil
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

	if expenseRequest.Status == "approved" && expenseRequest.AdvanceRequestID != nil {
		var advance models.AdvanceRequests
		if err := tx.First(&advance, *expenseRequest.AdvanceRequestID).Error; err == nil {
			if advance.Status == "approved" {
				settled, serr := AdvanceFullySettled(tx, &advance)
				if serr != nil {
					tx.Rollback()
					return serr
				}
				if settled {
					advance.Status = "completed"
					if err := tx.Save(&advance).Error; err != nil {
						tx.Rollback()
						return err
					}
					advanceMsg := fmt.Sprintf(
						"Your advance request (#%d) has been fully settled and COMPLETED via expense request #%d.",
						advance.ID, expenseRequest.ID,
					)
					r.sendSingleNotification(tx, advance.UserID, advance.ID, advanceMsg, "advance_completed")
				}
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	notifications.BroadcastRemovals(expenseRequest.ID, removedPerUser)
	return nil
}

// sendSingleNotification persists a notification inside tx and fans it out.
func (r *ExpenseApprovalsRepo) sendSingleNotification(
	tx *gorm.DB,
	userID uint,
	requestID uint,
	message string,
	notificationType string,
) {
	r.notifier.dispatchTx(tx, notifications.Notification{
		UserID:    userID,
		RequestID: requestID,
		Message:   message,
		Type:      notificationType,
		PushTitle: "Expense Request",
		Kind:      notifications.Expense,
	})
}

func (r *ExpenseApprovalsRepo) UpdateExpenseApprovalComment(
	id uint,
	comments string,
) error {
	result := r.db.
		Model(&models.ExpenseApprovals{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"comments": comments,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("no expense approval updated (invalid id?)")
	}

	return nil
}
