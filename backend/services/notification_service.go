package services

import (
	"github.com/greboid/tithon/irc"
)

type NotificationService struct {
	pendingNotifications chan irc.Notification
}

func NewNotificationService(pendingNotifications chan irc.Notification) *NotificationService {
	return &NotificationService{
		pendingNotifications: pendingNotifications,
	}
}

func (ns *NotificationService) HasPendingNotifications() bool {
	return len(ns.pendingNotifications) > 0
}

func (ns *NotificationService) GetNextNotification() *irc.Notification {
	select {
	case notification := <-ns.pendingNotifications:
		return &notification
	default:
		return nil
	}
}

func (ns *NotificationService) GetNotificationChannel() <-chan irc.Notification {
	return ns.pendingNotifications
}
