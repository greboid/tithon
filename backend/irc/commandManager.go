package irc

import (
	"errors"
	"fmt"
	"github.com/enescakir/emoji"
	"github.com/greboid/tithon/config"
	"log/slog"
	"regexp"
	"slices"
	"strings"
)

var (
	NoServerError  = errors.New("no server specified")
	NoChannelError = errors.New("no channel specified")
)

type CommandContext int

const (
	ContextAny CommandContext = iota
	ContextConnected
	ContextServer
	ContextChannel
	ContextQuery
	ContextChannelOrQuery
)

func (c CommandContext) String() string {
	switch c {
	case ContextAny:
		return "any"
	case ContextConnected:
		return "connected"
	case ContextServer:
		return "server"
	case ContextChannel:
		return "channel"
	case ContextQuery:
		return "query"
	case ContextChannelOrQuery:
		return "channel or query"
	default:
		return "unknown"
	}
}

type Command interface {
	GetName() string
	GetHelp() string
	Execute(*ServerManager, *Window, string) error
	GetArgSpecs() []Argument
	GetFlagSpecs() []Flag
	GetUsage() string
	GetAliases() []string
	GetContext() CommandContext
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

func NewCommandManager(conf *config.Config, showSettings chan bool) *CommandManager {
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
		&Help{cm: cm},
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
		if first == cm.commands[i].GetName() || slices.Contains(cm.commands[i].GetAliases(), first) {
			if !cm.validateContext(cm.commands[i], window) {
				cm.showContextError(window, cm.commands[i])
				return
			}

			var args string
			if len(input) > len(first) {
				args = strings.TrimPrefix(input, first+" ")
				args = emoji.Parse(args)
			}

			err := cm.commands[i].Execute(connections, window, args)
			if err != nil {
				cm.showCommandError(window, cm.commands[i], err.Error())
			}
			return
		}
	}

	cm.showError(window, fmt.Sprintf("Command '%s' not found. Use /help to see all available commands.", input))
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

func (cm *CommandManager) validateContext(command Command, window *Window) bool {
	context := command.GetContext()

	switch context {
	case ContextAny:
		return true
	case ContextConnected:
		return window != nil
	case ContextServer:
		return window != nil && window.IsServer()
	case ContextChannel:
		return window != nil && window.IsChannel()
	case ContextQuery:
		return window != nil && window.IsQuery()
	case ContextChannelOrQuery:
		return window != nil && (window.IsChannel() || window.IsQuery())
	default:
		return false
	}
}

func (cm *CommandManager) showContextError(window *Window, command Command) {
	context := command.GetContext()
	var message string

	switch context {
	case ContextConnected:
		message = fmt.Sprintf("Command '/%s' requires a connection to a server", command.GetName())
	case ContextServer:
		message = fmt.Sprintf("Command '/%s' can only be used in a server window", command.GetName())
	case ContextChannel:
		message = fmt.Sprintf("Command '/%s' can only be used in a channel", command.GetName())
	case ContextQuery:
		message = fmt.Sprintf("Command '/%s' can only be used in a query (private message)", command.GetName())
	case ContextChannelOrQuery:
		message = fmt.Sprintf("Command '/%s' can only be used in a channel or query", command.GetName())
	default:
		message = fmt.Sprintf("Command '/%s' cannot be used in this context", command.GetName())
	}

	cm.showError(window, message)
}

func (cm *CommandManager) showError(window *Window, message string) {
	if window != nil {
		window.AddMessage(NewError(cm.conf.UISettings.TimestampFormat, false, message))
	} else {
		slog.Error("Command error", "message", message)
	}
}
