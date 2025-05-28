package irc

import (
	"github.com/greboid/tithon/config"
	"log/slog"
	"regexp"
)

type Notification struct {
	Text string
}

type NotificationManager struct {
	notifications        []config.NotificationTrigger
	pendingNotifications chan Notification
}

func NewNotificationManager(pendingNotifications chan Notification, triggers []config.NotificationTrigger) *NotificationManager {
	return &NotificationManager{
		pendingNotifications: pendingNotifications,
		notifications:        triggers,
	}
}

func (cm *NotificationManager) SendNotification(text string) {
	cm.pendingNotifications <- Notification{Text: text}
}

func (cm *NotificationManager) IsNotification(network, source, nick, message string) bool {
	for i := range cm.notifications {
		if regexp.MustCompile(cm.notifications[i].Network).MatchString(network) &&
			regexp.MustCompile(cm.notifications[i].Source).MatchString(source) &&
			regexp.MustCompile(cm.notifications[i].Nick).MatchString(nick) &&
			regexp.MustCompile(cm.notifications[i].Message).MatchString(message) {
			slog.Debug("Notification matched")
			return true
		}
	}
	slog.Debug("Notification miss")
	return false
}
