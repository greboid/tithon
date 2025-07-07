package irc

import (
	"fmt"
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
			Type:        ArgTypeRestOfInput,
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

	nickname, err := parsed.GetArgString("nickname")
	if err != nil {
		return fmt.Errorf("failed to get nickname: %w", err)
	}

	message, err := parsed.GetArgString("message")
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	_, err = window.connection.GetQueryByName(nickname)
	if err != nil {
		window.connection.AddQuery(nickname)
	}

	if message != "" {
		err = window.connection.SendQuery(nickname, message)
		if err != nil {
			return err
		}
	}

	return nil
}
