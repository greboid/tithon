package irc

import (
	"github.com/greboid/tithon/config"
	"log/slog"
	"regexp"
)

type Notification struct {
	Text string
}

type Trigger struct {
	Network *regexp.Regexp
	Source  *regexp.Regexp
	Nick    *regexp.Regexp
	Message *regexp.Regexp
}

type NotificationManager struct {
	notifications        []Trigger
	pendingNotifications chan Notification
}

func NewNotificationManager(pendingNotifications chan Notification, triggers []config.NotificationTrigger) *NotificationManager {
	nm := &NotificationManager{
		pendingNotifications: pendingNotifications,
	}
	for i := range triggers {
		trigger := Trigger{}

		if triggers[i].Network == "" {
			triggers[i].Network = ".*"
		}
		reg, err := regexp.Compile(triggers[i].Network)
		if err != nil {
			slog.Error("Invalid network regexp", "error", err)
			continue
		}
		trigger.Network = reg

		if triggers[i].Source == "" {
			triggers[i].Source = ".*"
		}
		reg, err = regexp.Compile(triggers[i].Source)
		if err != nil {
			slog.Error("Invalid source regexp", "error", err)
			continue
		}
		trigger.Source = reg

		if triggers[i].Nick == "" {
			triggers[i].Nick = ".*"
		}
		reg, err = regexp.Compile(triggers[i].Nick)
		if err != nil {
			slog.Error("Invalid nick regexp", "error", err)
			continue
		}
		trigger.Nick = reg

		if triggers[i].Message == "" {
			triggers[i].Message = ".*"
		}
		reg, err = regexp.Compile(triggers[i].Message)
		if err != nil {
			slog.Error("Invalid message regexp", "error", err)
			continue
		}
		trigger.Message = reg

		nm.notifications = append(nm.notifications, trigger)
	}
	return nm
}

func (cm *NotificationManager) SendNotification(text string) {
	cm.pendingNotifications <- Notification{Text: text}
}

func (cm *NotificationManager) IsNotification(network, source, nick, message string) bool {
	for i := range cm.notifications {
		if cm.notifications[i].Network.MatchString(network) &&
			cm.notifications[i].Source.MatchString(source) &&
			cm.notifications[i].Nick.MatchString(nick) &&
			cm.notifications[i].Message.MatchString(message) {
			slog.Debug("Notification matched")
			return true
		}
	}
	slog.Debug("Notification miss")
	return false
}
