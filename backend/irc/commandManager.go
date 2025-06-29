package irc

import (
	"errors"
	"github.com/enescakir/emoji"
	"github.com/greboid/tithon/config"
	"log/slog"
	"regexp"
	"strings"
)

var (
	NoServerError  = errors.New("no server specified")
	NoChannelError = errors.New("no channel specified")
)

type Command interface {
	GetName() string
	GetHelp() string
	Execute(*ServerManager, *Window, string) error
}

type Notifier interface {
	showNotification(notification Notification)
}

type CommandManager struct {
	commands  []Command
	conf      *config.Config
	nm        NotificationManager
	LinkRegex *regexp.Regexp
}

func NewCommandManager(linkRegex *regexp.Regexp, conf *config.Config, showSettings chan bool) *CommandManager {
	cm := &CommandManager{}
	cm.commands = []Command{
		&SendAction{},
		&Msg{},
		&Quit{},
		&Join{},
		&Part{},
		&Nick{},
		&ChangeTopic{},
		&SendNotice{},
		&Whois{},
		&Notify{nm: cm},
		&QueryCommand{},
		&AddServer{},
		&Disconnect{},
		&Reconnect{},
		&CloseCommand{},
		&Settings{
			showSettings: showSettings,
		},
	}
	cm.LinkRegex = linkRegex
	cm.conf = conf
	return cm
}

func (cm *CommandManager) Execute(connections *ServerManager, window *Window, input string) {
	if !strings.HasPrefix(input, "/") {
		input = "/msg " + input
	}
	input = strings.TrimPrefix(input, "/")
	first := strings.Split(input, " ")[0]
	for i := range cm.commands {
		if first == cm.commands[i].GetName() {
			if len(input) == len(first) {
				err := cm.commands[i].Execute(connections, window, "")
				if err != nil {
					cm.showCommandError(window, cm.commands[i], err.Error())
				}
				return
			}
			input = strings.TrimPrefix(input, first+" ")
			input = emoji.Parse(input)
			err := cm.commands[i].Execute(connections, window, input)
			if err != nil {
				cm.showCommandError(window, cm.commands[i], err.Error())
			}
			return
		}
	}
	cm.showError(window, "Unknown command: "+input)
}

func (cm *CommandManager) SetNotificationManager(nm NotificationManager) {
	cm.nm = nm
}

func (cm *CommandManager) showNotification(notification Notification) {
	cm.nm.SendNotification(notification)
}

func (cm *CommandManager) showCommandError(window *Window, command Command, message string) {
	cm.showError(window, "Command Error: "+command.GetName()+": "+message)
}

func (cm *CommandManager) showError(window *Window, message string) {
	if window != nil {
		window.AddMessage(NewError(cm.LinkRegex, cm.conf.UISettings.TimestampFormat, false, message))
	} else {
		slog.Error("Command error", "message", message)
	}
}
