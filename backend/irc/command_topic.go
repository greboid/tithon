package irc

import (
	"fmt"
)

type ChangeTopic struct{}

func init() {
	RegisterCommand(&ChangeTopic{})
}

func (c ChangeTopic) GetName() string {
	return "topic"
}

func (c ChangeTopic) GetHelp() string {
	return "Changes the topic"
}

func (c ChangeTopic) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c ChangeTopic) GetArgSpecs() []Argument {
	return []Argument{
		{
			Name:        "channel",
			Type:        ArgTypeChannel,
			Required:    false,
			Description: "The channel to set topic for (defaults to current channel)",
			Validator:   validateNonEmpty,
		},
		{
			Name:        "topic",
			Type:        ArgTypeRestOfInput,
			Required:    true,
			Description: "The new topic text",
			Validator:   validateNonEmpty,
		},
	}
}

func (c ChangeTopic) GetFlagSpecs() []Flag {
	return []Flag{}
}

func (c ChangeTopic) GetAliases() []string {
	return []string{"t"}
}

func (c ChangeTopic) GetContext() CommandContext {
	return ContextConnected
}

func (c ChangeTopic) InjectDependencies(*CommandDependencies) {
	return
}

func (c ChangeTopic) Execute(_ *ServerManager, window *Window, input string) error {
	if window == nil {
		return NoServerError
	}

	parsed, err := Parse(c, input)
	if err != nil {
		return fmt.Errorf("argument parsing error: %w", err)
	}

	channel, err := parsed.GetArgStringWithChannelFallback("channel", window)
	if err != nil {
		return err
	}

	topic, err := parsed.GetArgString("topic")
	if err != nil {
		return fmt.Errorf("failed to get topic: %w", err)
	}

	return window.GetServer().SendTopic(channel, topic)
}
