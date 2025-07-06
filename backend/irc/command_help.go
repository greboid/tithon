package irc

import (
	"fmt"
	"sort"
	"strings"
)

type Help struct {
	cm *CommandManager
}

func (c Help) GetName() string {
	return "help"
}

func (c Help) GetHelp() string {
	return "Shows help for all commands or a specific command"
}

func (c Help) Execute(_ *ServerManager, window *Window, input string) error {
	if window == nil {
		return NoServerError
	}

	input = strings.TrimSpace(input)
	timestampFormat := c.cm.conf.UISettings.TimestampFormat

	if input != "" {
		for _, cmd := range c.cm.commands {
			if cmd.GetName() == input {
				helpText := fmt.Sprintf("/%s - %s", cmd.GetName(), cmd.GetHelp())
				window.AddMessage(NewEvent(EventHelp, timestampFormat, false, helpText))
				return nil
			}
		}

		errorText := fmt.Sprintf("Command '%s' not found. Use /help to see all available commands.", input)
		window.AddMessage(NewEvent(EventHelp, timestampFormat, false, errorText))
		return nil
	}

	var commands []string
	for _, cmd := range c.cm.commands {
		commands = append(commands, fmt.Sprintf("/%s - %s", cmd.GetName(), cmd.GetHelp()))
	}

	sort.Strings(commands)

	window.AddMessage(NewEvent(EventHelp, timestampFormat, false, "Available commands:"))

	for _, cmd := range commands {
		window.AddMessage(NewEvent(EventHelp, timestampFormat, false, cmd))
	}

	return nil
}
