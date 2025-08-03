package controllers

import (
	"net/http"
	"shwetaik-expense-management-api/services"
	"strconv"

	"github.com/labstack/echo/v4"
)

type NotificationController struct {
	NotificationService *services.NotificationsService
}

func NewNotificationController(notificationService *services.NotificationsService) *NotificationController {
	return &NotificationController{NotificationService: notificationService}
}

func (c *NotificationController) GetNotificationsByUserID(ctx echo.Context) error {
	userID := ctx.Param("userID")
	id, err := strconv.Atoi(userID)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, "Invalid user ID")
	}

	notifications, err := c.NotificationService.GetNotificationsByUserID(uint(id))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, notifications)
}

func (c *NotificationController) MarkNotificationAsRead(ctx echo.Context) error {
	notificationID := ctx.Param("id")
	id, err := strconv.Atoi(notificationID)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, "Invalid notification ID")
	}

	err = c.NotificationService.MarkNotificationAsRead(uint(id))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, "Notification marked as read")
}

func (c *NotificationController) MarkAllNotificationsAsRead(ctx echo.Context) error {
	userID := ctx.Param("userID")
	id, err := strconv.Atoi(userID)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, "Invalid user ID")
	}

	err = c.NotificationService.MarkAllNotificationsAsRead(uint(id))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, "All notifications marked as read")
}

func (c *NotificationController) DeleteNotification(ctx echo.Context) error {
	notificationID := ctx.Param("id")
	id, err := strconv.Atoi(notificationID)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, "Invalid notification ID")
	}

	err = c.NotificationService.DeleteNotification(uint(id))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, "Notification deleted successfully")
}

func (c *NotificationController) ClearAllNotifications(ctx echo.Context) error {
	userID := ctx.Param("userID")
	id, err := strconv.Atoi(userID)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, "Invalid user ID")
	}

	err = c.NotificationService.ClearAllNotifications(uint(id))
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err.Error())
	}
	return ctx.JSON(http.StatusOK, "All notifications cleared")
}
