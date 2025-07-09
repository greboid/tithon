package irc

import (
	"fmt"
	"strings"
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
			Name:        "topic",
			Type:        ArgTypeString,
			Required:    false,
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

	channel, err := parsed.GetFlagString("channel")
	if err != nil {
		channel = window.GetName()
	}

	args, err := parsed.GetArgs()
	if err != nil {
		return fmt.Errorf("failed to get arguments: %w", err)
	}
	if len(args) == 0 {
		return fmt.Errorf("incorrect number of arguments: topic")
	}
	return window.GetServer().SendTopic(channel, strings.Join(args[0:], " "))
}
