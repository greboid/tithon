package irc

import (
	"errors"
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
			err := cm.commands[i].Execute(connections, server, channel, input)
			if err != nil {
				if channel != nil {
					channel.AddMessage(NewMessage("", err.Error(), Error))
				} else if server != nil {
					server.AddMessage(NewMessage("", err.Error(), Error))
				} else {
					slog.Error("Command error", "input", input, "error", err)
				}
			}
			return
		}
	}
	if channel != nil {
		channel.AddMessage(NewMessage("", "Unknown command: "+input, Error))
	} else if server != nil {
		server.AddMessage(NewMessage("", "Unknown command: "+input, Error))
	} else {
		slog.Error("Unknown command", "input", input)
	}
}
