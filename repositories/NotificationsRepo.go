package repositories

import (
	"shwetaik-expense-management-api/models"

	"gorm.io/gorm"
)

type NotificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepo(db *gorm.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
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
