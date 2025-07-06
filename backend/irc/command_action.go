package irc

import (
	"fmt"
)

type SendAction struct{}

func (c SendAction) GetName() string {
	return "me"
}

func (c SendAction) GetHelp() string {
	return "Performs an action in a channel or query"
}

func (c SendAction) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c SendAction) GetArgSpecs() []Argument {
	return []Argument{
		{
			Name:        "message",
			Type:        ArgTypeString,
			Required:    true,
			Description: "The action message to send",
			Validator:   validateNonEmpty,
		},
	}
}

func (c SendAction) GetFlagSpecs() []Flag {
	return []Flag{}
}

func (c SendAction) Execute(_ *ServerManager, window *Window, input string) error {
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

	actionMessage := fmt.Sprintf("\001ACTION %s\001", message)
	return window.connection.SendMessage(window.GetID(), actionMessage)
}
