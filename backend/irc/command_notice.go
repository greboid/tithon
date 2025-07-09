package irc

import (
	"fmt"
	"strings"
)

type SendNotice struct{}

func init() {
	RegisterCommand(&SendNotice{})
}

func (c SendNotice) GetName() string {
	return "notice"
}

func (c SendNotice) GetHelp() string {
	return "Sends a notice to a channel or user"
}

func (c SendNotice) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c SendNotice) GetArgSpecs() []Argument {
	return []Argument{
		{
			Name:        "message",
			Type:        ArgTypeString,
			Required:    true,
			Description: "The notice message to send",
			Validator:   validateNonEmpty,
		},
	}
}

func (c SendNotice) GetFlagSpecs() []Flag {
	return []Flag{}
}

func (c SendNotice) GetAliases() []string {
	return []string{}
}

func (c SendNotice) GetContext() CommandContext {
	return ContextConnected
}

func (c SendNotice) InjectDependencies(*CommandDependencies) {
	return
}

func (c SendNotice) Execute(_ *ServerManager, window *Window, input string) error {
	if window == nil {
		return NoServerError
	}

	parsed, err := Parse(c, input)
	if err != nil {
		return fmt.Errorf("argument parsing error: %w", err)
	}

	args, err := parsed.GetArgs()
	if err != nil {
		return fmt.Errorf("failed to get arguments: %w", err)
	}
	if len(args) == 0 {
		return fmt.Errorf("incorrect number of arguments: message")
	}

	return window.connection.SendNotice(window.GetID(), strings.Join(args[0:], " "))
}
