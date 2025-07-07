package irc

import (
	"fmt"
)

type Quit struct{}

func (c Quit) GetName() string {
	return "quit"
}

func (c Quit) GetHelp() string {
	return "Quits a server, removing it from the list"
}

func (c Quit) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c Quit) GetArgSpecs() []Argument {
	return []Argument{
		{
			Name:        "message",
			Type:        ArgTypeRestOfInput,
			Required:    false,
			Default:     "",
			Description: "Optional quit message",
		},
	}
}

func (c Quit) GetFlagSpecs() []Flag {
	return []Flag{}
}

func (c Quit) GetAliases() []string {
	return []string{"q"}
}

func (c Quit) GetContext() CommandContext {
	return ContextConnected
}

func (c Quit) Execute(cm *ServerManager, window *Window, input string) error {
	if window == nil {
		return NoServerError
	}

	parsed, err := Parse(c, input)
	if err != nil {
		return fmt.Errorf("argument parsing error: %w", err)
	}

	message, err := parsed.GetArgString("message")
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	if message != "" {
		window.connection.SendMessage(window.GetID(), "QUIT :"+message)
	}

	cm.RemoveConnection(window.connection.GetID())
	return nil
}
