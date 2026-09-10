package repositories

import (
	"log"

	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/notifications"

	"gorm.io/gorm"
)

type notifier struct {
	notificationRepo *NotificationRepo
	dispatcher       *notifications.Dispatcher
}

func newNotifier(notificationRepo *NotificationRepo, deviceTokenRepo *DeviceTokenRepo) *notifier {
	return &notifier{
		notificationRepo: notificationRepo,
		dispatcher:       notifications.NewDispatcher(deviceTokenRepo, notificationRepo),
	}
}

func record(n notifications.Notification) *models.Notification {
	return &models.Notification{
		UserID:    n.UserID,
		ExpenseID: n.RequestID,
		Message:   n.Message,
		Type:      n.Type,
		IsRead:    false,
	}
}

func (n *notifier) dispatchTx(tx *gorm.DB, out notifications.Notification) {
	row := record(out)
	_ = tx.Create(row).Error
	n.dispatcher.FanOutAsync(out, row.ID, row.CreatedAt)
}

func (n *notifier) dispatch(out notifications.Notification) {
	row := record(out)
	if err := n.notificationRepo.CreateNotification(row); err != nil {
		log.Printf("Error saving notification to DB for user %d: %v", out.UserID, err)
	}
	n.dispatcher.FanOut(out, row.ID, row.CreatedAt)
}
