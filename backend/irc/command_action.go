package irc

import (
	"fmt"
	"strings"
)

type SendAction struct{}

func init() {
	RegisterCommand(&SendAction{})
}

func (c SendAction) GetName() string {
	return "me"
}

func (c SendAction) GetHelp() string {
	return "Performs an action in a channel or to a user"
}

func (c SendAction) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c SendAction) GetArgSpecs() []Argument {
	return []Argument{}
}

func (c SendAction) GetFlagSpecs() []Flag {
	return []Flag{}
}

func (c SendAction) GetAliases() []string {
	return []string{"action"}
}

func (c SendAction) GetContext() CommandContext {
	return ContextChannelOrQuery
}

func (c SendAction) InjectDependencies(*CommandDependencies) {
	return
}

func (c SendAction) Execute(_ *ServerManager, window *Window, input string) error {
	if window == nil {
		return NoServerError
	}

	parsed, err := Parse(c, input)
	if err != nil {
		return fmt.Errorf("argument parsing error: %w", err)
	}

	args, err := parsed.GetArgs()
	if err != nil {
		return fmt.Errorf("failed to get arguments: %w", err)
	}
	if len(args) == 0 {
		return fmt.Errorf("incorrect number of arguments: message")
	}

	return window.connection.SendMessage(window.GetID(), fmt.Sprintf("\001ACTION %s\001", strings.Join(args[0:], " ")))
}
