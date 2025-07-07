package irc

import (
	"fmt"
)

type SendNotice struct{}

func (c SendNotice) GetName() string {
	return "notice"
}

func (c SendNotice) GetHelp() string {
	return "Sends a notice"
}

func (c SendNotice) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c SendNotice) GetArgSpecs() []Argument {
	return []Argument{
		{
			Name:        "message",
			Type:        ArgTypeRestOfInput,
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
	return ContextChannelOrQuery
}

func (c SendNotice) Execute(_ *ServerManager, window *Window, input string) error {
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

	return window.connection.SendNotice(window.GetID(), message)
}
