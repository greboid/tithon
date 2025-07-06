package irc

import (
	"fmt"
)

type Join struct{}

func (c Join) GetName() string {
	return "join"
}

func (c Join) GetHelp() string {
	return "Joins a channel"
}

func (c Join) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c Join) GetArgSpecs() []Argument {
	return []Argument{
		{
			Name:        "channel",
			Type:        ArgTypeChannel,
			Required:    true,
			Description: "The channel to join (e.g., #general)",
			Validator:   validateNonEmpty,
		},
	}
}

func (c Join) GetFlagSpecs() []Flag {
	return []Flag{
		{
			Name:        "key",
			Short:       "k",
			Type:        ArgTypeString,
			Required:    false,
			Default:     "",
			Description: "Channel key/password if required",
		},
	}
}

func (c Join) Execute(_ *ServerManager, window *Window, input string) error {
	if window == nil {
		return NoServerError
	}
	
	parsed, err := Parse(c, input)
	if err != nil {
		return fmt.Errorf("argument parsing error: %w", err)
	}

	channel, err := parsed.GetArgString("channel")
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	key, err := parsed.GetFlagString("key")
	if err != nil {
		return fmt.Errorf("failed to get key: %w", err)
	}

	return window.connection.JoinChannel(channel, key)
}
