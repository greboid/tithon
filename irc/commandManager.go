package irc

import (
	"fmt"
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
	}}
}

func (cm *CommandManager) Execute(connections *ConnectionManager, server *Connection, channel *Channel, input string) {
	if !strings.HasPrefix(input, "/") {
		input = "/msg " + input
	}
	for i := range cm.commands {
		prefix := fmt.Sprintf("/%s ", cm.commands[i].GetName())
		if strings.HasPrefix(input, prefix) {
			input = strings.TrimPrefix(input, prefix)
			cm.commands[i].Execute(connections, server, channel, input)
			return
		}
	}
	slog.Error("Unable to find command", "input", input)
}
