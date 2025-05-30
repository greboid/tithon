package irc

import (
	"errors"
	"github.com/enescakir/emoji"
	"github.com/greboid/tithon/config"
	"log/slog"
	"strings"
	"time"
)

var (
	NoServerError  = errors.New("no server specified")
	NoChannelError = errors.New("no channel specified")
)

type Command interface {
	GetName() string
	GetHelp() string
	Execute(*ConnectionManager, *Window, string) error
}

type CommandManager struct {
	commands []Command
	conf     *config.Config
}

func NewCommandManager(conf *config.Config) *CommandManager {
	return &CommandManager{
		commands: []Command{
			&SendAction{},
			&Msg{},
			&Quit{},
			&Join{},
			&Part{},
			&Nick{},
			&ChangeTopic{},
			&SendNotice{},
			&Whois{},
		},
		conf: conf,
	}
}

func (cm *CommandManager) Execute(connections *ConnectionManager, window *Window, input string) {
	if !strings.HasPrefix(input, "/") {
		input = "/msg " + input
	}
	input = strings.TrimPrefix(input, "/")
	first := strings.Split(input, " ")[0]
	for i := range cm.commands {
		if first == cm.commands[i].GetName() {
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

func (cm *CommandManager) showCommandError(window *Window, command Command, message string) {
	cm.showError(window, "Command Error: "+command.GetName()+": "+message)
}

func (cm *CommandManager) showError(window *Window, message string) {
	if window != nil {
		window.AddMessage(NewError(time.Now(), cm.conf.UISettings.TimestampFormat, false, message))
	} else {
		slog.Error("Command error", "message", message)
	}
}
