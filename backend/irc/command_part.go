package irc

import (
	"fmt"
)

type Part struct{}

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

func (c Part) Execute(_ *ServerManager, window *Window, input string) error {
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
		// Send part message before leaving
		window.connection.SendMessage(window.GetID(), "PART "+window.GetID()+" :"+message)
	}
	
	window.connection.RemoveChannel(window.GetID())
	return nil
}
