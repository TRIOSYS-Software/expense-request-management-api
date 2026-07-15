package repositories

import (
	"errors"
	"fmt"
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/utilities"
	"time"

	firebase "firebase.google.com/go/v4"

	"gorm.io/gorm"
)

type AdvanceApprovalsRepo struct {
	db               *gorm.DB
	notificationRepo *NotificationRepo
	deviceTokenRepo  *DeviceTokenRepo
}

func NewAdvanceApprovalsRepo(db *gorm.DB, firebaseApp *firebase.App) *AdvanceApprovalsRepo {
	return &AdvanceApprovalsRepo{
		db:               db,
		notificationRepo: NewNotificationRepo(db, firebaseApp),
		deviceTokenRepo:  NewDeviceTokenRepo(db),
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

func (r *AdvanceApprovalsRepo) UpdateAdvanceApproval(id uint, advanceApproval *models.AdvanceApprovals) error {
	tx := r.db.Begin()
	defer func() {
		if rec := recover(); rec != nil {
			_ = tx.Rollback()
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
		broadcastNotificationRemovals(advanceRequest.ID, removedPerUser)
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
	broadcastNotificationRemovals(advanceRequest.ID, removedPerUser)
	return nil
}

func (r *AdvanceApprovalsRepo) sendSingleNotification(
	tx *gorm.DB,
	userID uint,
	advanceID uint,
	message string,
	notificationType string,
) {
	notification := &models.Notification{
		UserID:    userID,
		ExpenseID: advanceID,
		Message:   message,
		Type:      notificationType,
		IsRead:    false,
	}

	_ = tx.Create(notification).Error

	tokens, err := r.deviceTokenRepo.GetTokensByUserID(userID)
	if err == nil && len(tokens) > 0 {
		data := map[string]string{
			"advanceId": fmt.Sprintf("%d", advanceID),
			"type":      notificationType,
		}
		go r.notificationRepo.SendPushNotification(tokens, "Advance Request", message, data)
	}

	go utilities.SendWebSocketMessage(
		userID,
		utilities.WebSocketMessagePayload{
			ID:        notification.ID,
			Message:   message,
			Type:      notificationType,
			ExpenseID: advanceID,
			IsRead:    false,
			CreatedAt: notification.CreatedAt.Format(time.RFC3339),
		},
	)
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
