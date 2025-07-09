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
	"sync"
)

var (
	NoServerError   = errors.New("no server specified")
	NoChannelError  = errors.New("no channel specified")
	commandRegistry = make([]Command, 0)
	registryMutex   sync.RWMutex
)

type CommandContext int

const (
	ContextAny CommandContext = iota
	ContextConnected
	ContextServer
	ContextChannel
	ContextQuery
	ContextChannelOrQuery
	ContextAnyWithTarget
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

type CommandDependencies struct {
	CommandManager *CommandManager
	ShowSettings   chan bool
	Notifier       Notifier
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
	InjectDependencies(*CommandDependencies)
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
	cm.LinkRegex = linkRegex
	cm.conf = conf

	registeredCommands := GetRegisteredCommands()

	deps := &CommandDependencies{
		CommandManager: cm,
		ShowSettings:   showSettings,
		Notifier:       cm,
	}

	cm.commands = make([]Command, 0)
	for _, cmd := range registeredCommands {
		cmd.InjectDependencies(deps)
		cm.commands = append(cm.commands, cmd)
	}

	return cm
}

func (cm *CommandManager) Execute(connections *ServerManager, window *Window, input string) {
	if !strings.HasPrefix(input, "/") {
		input = "/msg " + input
	}
	input = strings.TrimPrefix(input, "/")
	first := strings.Split(input, " ")[0]
	cmdIndex := slices.IndexFunc(cm.commands, func(command Command) bool {
		return command.GetName() == first || slices.Contains(command.GetAliases(), first)
	})
	if cmdIndex == -1 {
		cm.showError(window, fmt.Sprintf("Command '%s' not found. Use /help to see all available commands.", first))
		return
	}
	cmd := cm.commands[cmdIndex]

	if !cm.validateContext(cmd, window) {
		cm.showContextError(window, cmd)
		return
	}

	var args string
	if len(input) > len(first) {
		args = strings.TrimPrefix(input, first+" ")
		args = emoji.Parse(args)
	}

	err := cmd.Execute(connections, window, args)
	if err != nil {
		cm.showCommandError(window, cmd, err.Error())
	}
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

func RegisterCommand(cmd Command) {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	commandRegistry = append(commandRegistry, cmd)
}

func GetRegisteredCommands() []Command {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	commands := make([]Command, 0)
	for _, cmd := range commandRegistry {
		commands = append(commands, cmd)
	}
	return commands
}
