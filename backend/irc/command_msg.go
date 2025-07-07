package irc

import (
	"fmt"
)

type Msg struct{}

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
			Name:        "target",
			Type:        ArgTypeChannelOrNick,
			Required:    false,
			Description: "Target channel or nickname (defaults to current window)",
		},
		{
			Name:        "message",
			Type:        ArgTypeRestOfInput,
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
	return ContextConnected
}

func (c Msg) Execute(_ *ServerManager, window *Window, input string) error {
	if window == nil {
		return NoServerError
	}

	parsed, err := Parse(c, input)
	if err != nil {
		return fmt.Errorf("argument parsing error: %w", err)
	}

	target, err := parsed.GetArgStringWithTargetFallback("target", window)
	if err != nil {
		return err
	}

	message, err := parsed.GetArgString("message")
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	return window.connection.SendMessage(target, message)
}
