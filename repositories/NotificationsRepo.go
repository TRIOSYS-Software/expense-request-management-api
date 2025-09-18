package repositories

import (
	"context"
	"log"
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
)

type NotificationRepo struct {
	db *gorm.DB
	firebaseApp     *firebase.App     
}

func (r *NotificationRepo) SendPushNotification(tokens []string, title string, body string, data map[string]string) {
    ctx := context.Background()
    client, err := r.firebaseApp.Messaging(ctx)
    if err != nil {
        log.Printf("Error: Unable to get Firebase Messaging client: %v", err)
        return
    }

    message := &messaging.MulticastMessage{
        Notification: &messaging.Notification{
            Title: title,
            Body:  body,
        },
        Data:   data, 
        Tokens: tokens,
    }

    br, err := client.SendEachForMulticast(ctx, message)
    if err != nil {
        log.Printf("Error: Failed to send FCM message: %v", err)
        return
    }
    log.Printf("FCM push notification sent: %d successful, %d failed.", br.SuccessCount, br.FailureCount)

    if br.FailureCount > 0 {
        for idx, resp := range br.Responses {
            if !resp.Success {
                log.Printf(" -> Failed token: %s (Error: %v)", tokens[idx], resp.Error)
            }
        }
    }
}

func NewNotificationRepo(db *gorm.DB,
	app *firebase.App,) *NotificationRepo {
	return &NotificationRepo{
		db: db,
		firebaseApp:     app,}
}

func (r *NotificationRepo) CreateNotification(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

func (r *NotificationRepo) GetNotificationsByUserID(userID uint) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&notifications).Error
	return notifications, err
}

func (r *NotificationRepo) MarkNotificationAsRead(id uint) error {
	return r.db.Model(&models.Notification{}).Where("id = ?", id).Update("is_read", true).Error
}

func (r *NotificationRepo) MarkAllNotificationsAsRead(userID uint) error {
	return r.db.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Update("is_read", true).Error
}

func (r *NotificationRepo) DeleteNotification(id uint) error {
	return r.db.Delete(&models.Notification{}, id).Error
}

func (r *NotificationRepo) ClearAllNotifications(userID uint) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.Notification{}).Error
}
