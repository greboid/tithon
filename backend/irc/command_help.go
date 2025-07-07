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

func (c Help) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c Help) GetArgSpecs() []Argument {
	return []Argument{
		{
			Name:        "command",
			Type:        ArgTypeString,
			Required:    false,
			Default:     "",
			Description: "Optional command name to get specific help for",
		},
	}
}

func (c Help) GetFlagSpecs() []Flag {
	return []Flag{}
}

func (c Help) GetAliases() []string {
	return []string{"h", "?"}
}

func (c Help) GetContext() CommandContext {
	return ContextAny
}

func (c Help) Execute(_ *ServerManager, window *Window, input string) error {
	if window == nil {
		return NoServerError
	}

	parsed, err := Parse(c, input)
	if err != nil {
		return fmt.Errorf("argument parsing error: %w", err)
	}

	command, err := parsed.GetArgString("command")
	if err != nil {
		return fmt.Errorf("failed to get command: %w", err)
	}

	command = strings.TrimSpace(command)
	timestampFormat := c.cm.conf.UISettings.TimestampFormat

	if command != "" {
		for _, cmd := range c.cm.commands {
			if cmd.GetName() == command {
				aliases := cmd.GetAliases()
				helpText := fmt.Sprintf("/%s", cmd.GetName())
				if len(aliases) > 0 {
					helpText += fmt.Sprintf(" (aliases: %s)", strings.Join(aliases, ", "))
				}
				helpText += "\n"
				helpText += cmd.GetUsage()
				window.AddMessage(NewEvent(EventHelp, timestampFormat, false, helpText))
				return nil
			}
		}

		errorText := fmt.Sprintf("Command '%s' not found. Use /help to see all available commands.", command)
		window.AddMessage(NewEvent(EventHelp, timestampFormat, false, errorText))
		return nil
	}

	var commands []string
	for _, cmd := range c.cm.commands {
		aliases := cmd.GetAliases()
		var cmdLine string
		if len(aliases) > 0 {
			cmdLine = fmt.Sprintf("/%s (%s) - %s", cmd.GetName(), strings.Join(aliases, ", "), cmd.GetHelp())
		} else {
			cmdLine = fmt.Sprintf("/%s - %s", cmd.GetName(), cmd.GetHelp())
		}
		commands = append(commands, cmdLine)
	}

	sort.Strings(commands)

	window.AddMessage(NewEvent(EventHelp, timestampFormat, false, "Available commands:"))

	for _, cmd := range commands {
		window.AddMessage(NewEvent(EventHelp, timestampFormat, false, cmd))
	}

	return nil
}
