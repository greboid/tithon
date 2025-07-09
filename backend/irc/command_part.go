package irc

import (
	"fmt"
	"strings"
)

type Part struct{}

func init() {
	RegisterCommand(&Part{})
}

func (c Part) GetName() string {
	return "part"
}

func (c Part) GetHelp() string {
	return "Parts a channel"
}

func (c Part) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c Part) GetArgSpecs() []Argument {
	return []Argument{
		{
			Name:        "message",
			Type:        ArgTypeString,
			Required:    false,
			Default:     "",
			Description: "Optional part message",
		},
	}
}

func (c Part) GetFlagSpecs() []Flag {
	return []Flag{}
}

func (c Part) GetAliases() []string {
	return []string{"p", "leave"}
}

func (c Part) GetContext() CommandContext {
	return ContextChannel
}

func (c Part) InjectDependencies(*CommandDependencies) {
	return
}

func (c Part) Execute(_ *ServerManager, window *Window, input string) error {
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
	message := ""
	if len(args) >= 2 {
		message = strings.Join(args[1:], " ")
	}

	if args[0] != "" {
		window.connection.SendMessage(window.GetID(), "PART "+window.GetID()+" :"+message)
	}

	window.connection.RemoveChannel(window.GetID())
	return nil
}
