package irc

import (
	"fmt"
	"github.com/greboid/tithon/config"
	"log/slog"
	"regexp"
	"sort"
)

type Notification struct {
	Title string
	Text  string
	Sound bool
	Popup bool
}

type Trigger struct {
	Network *regexp.Regexp
	Source  *regexp.Regexp
	Nick    *regexp.Regexp
	Message *regexp.Regexp
	Sound   bool
	Popup   bool
}

type NotificationManager struct {
	notifications        []Trigger
	pendingNotifications chan Notification
}

func NewNotificationManager(pendingNotifications chan Notification, triggers []config.NotificationTrigger) *NotificationManager {
	nm := &NotificationManager{
		pendingNotifications: pendingNotifications,
	}
	triggers = nm.sortTriggers(triggers)

	for i := range triggers {
		trigger := Trigger{
			Sound: triggers[i].Sound,
			Popup: triggers[i].Popup,
		}

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

func (cm *NotificationManager) sortTriggers(triggers []config.NotificationTrigger) []config.NotificationTrigger {
	sort.SliceStable(triggers, func(i, j int) bool {
		lenI := cm.getTriggerSpecificity(triggers[i])
		lenJ := cm.getTriggerSpecificity(triggers[j])
		if lenI != lenJ {
			return lenI > lenJ
		}
		if triggers[i].Sound != triggers[j].Sound {
			return triggers[i].Sound
		}
		return triggers[i].Popup
	})
	return triggers
}

func (cm *NotificationManager) getTriggerSpecificity(trigger config.NotificationTrigger) int {
	length := 0
	if trigger.Network != "" && trigger.Network != ".*" {
		length += len(trigger.Network)
	}
	if trigger.Source != "" && trigger.Source != ".*" {
		length += len(trigger.Source)
	}
	if trigger.Nick != "" && trigger.Nick != ".*" {
		length += len(trigger.Nick)
	}
	if trigger.Message != "" && trigger.Message != ".*" {
		length += len(trigger.Message)
	}
	return length
}

func (cm *NotificationManager) CheckAndNotify(network, source, nick, message string) bool {
	for i := range cm.notifications {
		if cm.notifications[i].Network.MatchString(network) &&
			cm.notifications[i].Source.MatchString(source) &&
			cm.notifications[i].Nick.MatchString(nick) &&
			cm.notifications[i].Message.MatchString(message) {
			cm.pendingNotifications <- Notification{
				Title: fmt.Sprintf("%s (%s)", nick, source),
				Text:  fmt.Sprintf("%s", message),
				Sound: cm.notifications[i].Sound,
				Popup: cm.notifications[i].Popup,
			}
		}
	}
	return false
}
