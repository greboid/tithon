package irc

import (
	"fmt"
	"strings"
)

type QueryCommand struct{}

func init() {
	RegisterCommand(&QueryCommand{})
}

func (c QueryCommand) GetName() string {
	return "query"
}

func (c QueryCommand) GetHelp() string {
	return "Opens a query window with the specified user"
}

func (c QueryCommand) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c QueryCommand) GetArgSpecs() []Argument {
	return []Argument{
		{
			Name:        "nickname",
			Type:        ArgTypeNick,
			Required:    true,
			Description: "The nickname to open a query with",
			Validator:   validateNonEmpty,
		},
		{
			Name:        "message",
			Type:        ArgTypeString,
			Required:    false,
			Default:     "",
			Description: "Optional initial message to send",
		},
	}
}

func (c QueryCommand) GetFlagSpecs() []Flag {
	return []Flag{}
}

func (c QueryCommand) GetAliases() []string {
	return []string{"q"}
}

func (c QueryCommand) GetContext() CommandContext {
	return ContextConnected
}

func (c QueryCommand) InjectDependencies(*CommandDependencies) {
	return
}

func (c QueryCommand) Execute(_ *ServerManager, window *Window, input string) error {
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
		return fmt.Errorf("incorrect number of arguments: nickname message")
	}
	message := ""
	if len(args) >= 2 {
		message = strings.Join(args[1:], " ")
	}

	_, err = window.connection.GetQueryByName(args[0])
	if err != nil {
		window.connection.AddQuery(args[0])
	}

	if message != "" {
		err = window.connection.SendQuery(args[0], message)
		if err != nil {
			return err
		}
	}

	return nil
}
