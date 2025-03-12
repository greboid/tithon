package irc

import (
	"log/slog"
	"strings"
)

type Command interface {
	GetName() string
	GetHelp() string
	Execute(*ConnectionManager, *Connection, *Channel, string)
}

type CommandManager struct {
	commands []Command
}

func NewCommandManager() *CommandManager {
	return &CommandManager{[]Command{
		&Action{},
		&Msg{},
		&Quit{},
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
			input = strings.TrimPrefix(input, first)
			cm.commands[i].Execute(connections, server, channel, input)
			return
		}
	}
	slog.Error("Unable to find command", "input", input)
}
