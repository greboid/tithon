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

	nm.notifications = nm.convertFromConfig(triggers)
	return nm
}

func (cm *NotificationManager) convertFromConfig(triggers []config.NotificationTrigger) []Trigger {
	triggers = cm.sortTriggers(triggers)
	var result []Trigger
	for i := range triggers {
		trigger, err := cm.AddNotification(triggers[i].Network, triggers[i].Source, triggers[i].Nick, triggers[i].Message, triggers[i].Sound, triggers[i].Popup)
		if err != nil {
			slog.Error("Invalid notification", "error", err)
			continue
		}
		result = append(result, *trigger)
	}
	return result
}

func (cm *NotificationManager) AddNotification(network, source, nick, message string, sound bool, popup bool) (*Trigger, error) {
	trigger := &Trigger{
		Sound: sound,
		Popup: popup,
	}

	reg, err := cm.compileRegex(network)
	if err != nil {
		return nil, fmt.Errorf("invalid network regex: %w", err)
	}
	trigger.Network = reg

	reg, err = cm.compileRegex(source)
	if err != nil {
		return nil, fmt.Errorf("invalid source regex: %w", err)
	}
	trigger.Source = reg

	reg, err = cm.compileRegex(nick)
	if err != nil {
		return nil, fmt.Errorf("invalid nick regex: %w", err)
	}
	trigger.Nick = reg

	reg, err = cm.compileRegex(message)
	if err != nil {
		return nil, fmt.Errorf("invalid message regex: %w", err)
	}
	trigger.Message = reg
	return trigger, nil
}

func (cm *NotificationManager) compileRegex(regex string) (*regexp.Regexp, error) {
	if regex == "" {
		return regexp.Compile(".*")
	}
	return regexp.Compile(regex)
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
			break
		}
	}
	return false
}
