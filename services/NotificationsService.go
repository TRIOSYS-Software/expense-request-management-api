package services

import (
	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
)

type NotificationsService struct {
	repo *repositories.NotificationRepo
}

func NewNotificationService(repo *repositories.NotificationRepo) *NotificationsService {
	return &NotificationsService{repo: repo}
}

func (s *NotificationsService) GetNotificationsByUserID(userID uint) ([]models.Notification, error) {
	return s.repo.GetNotificationsByUserID(userID)
}

func (s *NotificationsService) MarkNotificationAsRead(id uint) error {
	return s.repo.MarkNotificationAsRead(id)
}

func (s *NotificationsService) MarkAllNotificationsAsRead(userID uint) error {
	return s.repo.MarkAllNotificationsAsRead(userID)
}

func (s *NotificationsService) DeleteNotification(id uint) error {
	return s.repo.DeleteNotification(id)
}

func (s *NotificationsService) ClearAllNotifications(userID uint) error {
	return s.repo.ClearAllNotifications(userID)
}
