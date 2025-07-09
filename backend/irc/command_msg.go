package irc

import (
	"fmt"
	"strings"
)

type Msg struct{}

func init() {
	RegisterCommand(&Msg{})
}

func (c Msg) GetName() string {
	return "msg"
}

func (c Msg) GetHelp() string {
	return "Sends a message to a channel or user"
}

func (c Msg) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c Msg) GetArgSpecs() []Argument {
	return []Argument{
		{
			Name:        "message",
			Type:        ArgTypeString,
			Required:    true,
			Description: "The message to send",
			Validator:   validateNonEmpty,
		},
	}
}

func (c Msg) GetFlagSpecs() []Flag {
	return []Flag{}
}

func (c Msg) GetAliases() []string {
	return []string{"m", "say"}
}

func (c Msg) GetContext() CommandContext {
	return ContextChannelOrQuery
}

func (c Msg) InjectDependencies(*CommandDependencies) {
	return
}

func (c Msg) Execute(_ *ServerManager, window *Window, input string) error {
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

	return window.connection.SendMessage(window.GetID(), strings.Join(args[0:], " "))
}
