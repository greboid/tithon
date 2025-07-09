package irc

import (
	"fmt"
	"strings"
)

type Quit struct{}

func init() {
	RegisterCommand(&Quit{})
}

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
			Type:        ArgTypeString,
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

func (c Quit) InjectDependencies(*CommandDependencies) {
	return
}

func (c Quit) Execute(cm *ServerManager, window *Window, input string) error {
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
	message := ""
	if len(args) >= 0 {
		message = strings.Join(args[0:], " ")
	}

	if message != "" {
		window.connection.SendMessage(window.GetID(), "QUIT :"+message)
	}

	cm.RemoveConnection(window.connection.GetID())
	return nil
}
