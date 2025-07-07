package irc

import (
	"fmt"
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
			Name:        "target",
			Type:        ArgTypeChannelOrNick,
			Required:    false,
			Description: "Target channel or nickname (defaults to current window)",
		},
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

	target, err := parsed.GetArgStringWithTargetFallback("target", window)
	if err != nil {
		return err
	}

	message, err := parsed.GetArgString("message")
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	return window.connection.SendNotice(target, message)
}
