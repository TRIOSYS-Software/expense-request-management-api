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

type AdvanceApprovalsRepo struct {
	db               *gorm.DB
	notificationRepo *NotificationRepo
	deviceTokenRepo  *DeviceTokenRepo
	notifier         *notifier
}

func NewAdvanceApprovalsRepo(db *gorm.DB, firebaseApp *firebase.App) *AdvanceApprovalsRepo {
	notificationRepo := NewNotificationRepo(db, firebaseApp)
	deviceTokenRepo := NewDeviceTokenRepo(db)
	return &AdvanceApprovalsRepo{
		db:               db,
		notificationRepo: notificationRepo,
		deviceTokenRepo:  deviceTokenRepo,
		notifier:         newNotifier(notificationRepo, deviceTokenRepo),
	}
}

func (r *AdvanceApprovalsRepo) GetAdvanceApprovals() []models.AdvanceApprovals {
	var advanceApprovals []models.AdvanceApprovals
	r.db.Preload("Users", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, email")
	}).Find(&advanceApprovals)
	return advanceApprovals
}

func (r *AdvanceApprovalsRepo) GetAdvanceApprovalsByApproverID(approverID uint) []models.AdvanceApprovals {
	var advanceApprovals []models.AdvanceApprovals
	r.db.Where("approver_id = ?", approverID).Preload("Users", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, name, email")
	}).Find(&advanceApprovals)
	return advanceApprovals
}

func (r *AdvanceApprovalsRepo) UpdateAdvanceApproval(id uint, advanceApproval *models.AdvanceApprovals) (err error) {
	tx := r.db.Begin()
	defer func() {
		if rec := recover(); rec != nil {
			_ = tx.Rollback()
			err = fmt.Errorf("internal error in UpdateAdvanceApproval: %v", rec)
			log.Printf("PANIC recovered in UpdateAdvanceApproval: %v", rec)
		}
	}()

	var toUpdate models.AdvanceApprovals
	if err := tx.Preload("Users").Where("id = ?", id).First(&toUpdate).Error; err != nil {
		tx.Rollback()
		return err
	}

	var advanceRequest models.AdvanceRequests
	if err := tx.Preload("Approvals").Preload("User").
		Where("id = ?", toUpdate.RequestID).First(&advanceRequest).Error; err != nil {
		tx.Rollback()
		return err
	}

	if advanceRequest.Status != "pending" {
		tx.Rollback()
		return errors.New("this advance request has already been finalized; please refresh")
	}
	if toUpdate.Status != "pending" {
		tx.Rollback()
		return errors.New("this approval has already been processed; please refresh")
	}
	if toUpdate.Level != advanceRequest.CurrentApproverLevel {
		tx.Rollback()
		return errors.New("this request has already advanced past your level; please refresh")
	}

	originalRequestCreatorID := advanceRequest.UserID
	originalRequestCreatorName := advanceRequest.User.Name
	description := advanceRequest.Description

	toUpdate.Status = advanceApproval.Status
	toUpdate.Comments = advanceApproval.Comments

	if advanceApproval.Status == "approved" || advanceApproval.Status == "rejected" {
		now := time.Now()
		toUpdate.ApprovalDate = &now
	}

	if err := tx.Save(&toUpdate).Error; err != nil {
		tx.Rollback()
		return err
	}

	removedPerUser, err := r.notificationRepo.DeleteActionableForRequest(
		tx,
		advanceRequest.ID,
		[]string{"advance_new_request", "advance_pending_approval"},
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	if advanceApproval.Status == "rejected" {
		advanceRequest.Status = "rejected"
		advanceRequest.CurrentApproverLevel = toUpdate.Level

		comment := ""
		if advanceApproval.Comments != nil {
			comment = *advanceApproval.Comments
		}
		msg := fmt.Sprintf(
			"Your advance request (#%d - '%s') has been REJECTED by %s (Level %d). Reason: %s",
			advanceRequest.ID, description, toUpdate.Users.Name, toUpdate.Level, comment,
		)
		r.sendSingleNotification(tx, originalRequestCreatorID, advanceRequest.ID, msg, "advance_rejected")

		if err := tx.Save(&advanceRequest).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit().Error; err != nil {
			return err
		}
		notifications.BroadcastRemovals(advanceRequest.ID, removedPerUser)
		return nil
	}

	if advanceApproval.Status == "approved" {
		advanceRequest.CurrentApproverLevel = toUpdate.Level + 1

		var nextLevelApprovals []models.AdvanceApprovals
		if err := tx.Where("request_id = ? AND level = ?", advanceRequest.ID, advanceRequest.CurrentApproverLevel).
			Preload("Users").Find(&nextLevelApprovals).Error; err != nil {
			tx.Rollback()
			return err
		}

		if len(nextLevelApprovals) > 0 {
			for _, approver := range nextLevelApprovals {
				msg := fmt.Sprintf(
					"You have a new advance request (#%d - '%s') from %s to approve. (Level %d)",
					advanceRequest.ID, description, originalRequestCreatorName, advanceRequest.CurrentApproverLevel,
				)
				r.sendSingleNotification(tx, approver.ApproverID, advanceRequest.ID, msg, "advance_pending_approval")
			}
		} else {
			advanceRequest.Status = "approved"
			msg := fmt.Sprintf(
				"Your advance request (#%d - '%s') has been fully APPROVED!",
				advanceRequest.ID, description,
			)
			r.sendSingleNotification(tx, originalRequestCreatorID, advanceRequest.ID, msg, "advance_approved_final")
		}
	}

	if err := tx.Save(&advanceRequest).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}
	notifications.BroadcastRemovals(advanceRequest.ID, removedPerUser)
	return nil
}

// sendSingleNotification persists a notification inside tx and fans it out.
func (r *AdvanceApprovalsRepo) sendSingleNotification(
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
		PushTitle: "Advance Request",
		Kind:      notifications.Advance,
	})
}

func (r *AdvanceApprovalsRepo) UpdateAdvanceApprovalComment(id uint, comments string) error {
	result := r.db.
		Model(&models.AdvanceApprovals{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"comments": comments})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no advance approval updated (invalid id?)")
	}
	return nil
}
