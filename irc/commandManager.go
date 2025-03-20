package irc

import (
	"errors"
	"github.com/enescakir/emoji"
	"log/slog"
	"strings"
)

var (
	NoServerError  = errors.New("no server specified")
	NoChannelError = errors.New("no channel specified")
)

type Command interface {
	GetName() string
	GetHelp() string
	Execute(*ConnectionManager, *Connection, *Channel, string) error
}

type CommandManager struct {
	commands []Command
}

func NewCommandManager() *CommandManager {
	return &CommandManager{[]Command{
		&SendAction{},
		&Msg{},
		&Quit{},
		&Join{},
		&Part{},
		&Nick{},
		&ChangeTopic{},
	}}
}

func (cm *CommandManager) Execute(connections *ConnectionManager, server *Connection, channel *Channel, input string) {
	if !strings.HasPrefix(input, "/") {
		input = "/msg " + input
	}
	input = strings.TrimPrefix(input, "/")
	first := strings.Split(input, " ")[0]
	for i := range cm.commands {
		if first == cm.commands[i].GetName() {
			input = strings.TrimPrefix(input, first+" ")
			input = emoji.Parse(input)
			err := cm.commands[i].Execute(connections, server, channel, input)
			if err != nil {
				cm.showCommandError(server, channel, cm.commands[i], err.Error())
			}
			return
		}
	}
	cm.showError(server, channel, "Unknown command: "+input)
}

func (cm *CommandManager) showCommandError(server *Connection, channel *Channel, command Command, message string) {
	cm.showError(server, channel, "Command Error: "+command.GetName()+": "+message)
}

func (cm *CommandManager) showError(server *Connection, channel *Channel, message string) {
	if channel != nil {
		channel.AddMessage(NewMessage("", message, Error))
	} else if server != nil {
		server.AddMessage(NewMessage("", message, Error))
	} else {
		slog.Error("Command error", "message", message)
	}
}
