package irc

import (
	"fmt"
)

type Whois struct{}

func init() {
	RegisterCommand(&Whois{})
}

func (c Whois) GetName() string {
	return "whois"
}

func (c Whois) GetHelp() string {
	return "Looks up information about a user"
}

func (c Whois) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c Whois) GetArgSpecs() []Argument {
	return []Argument{
		{
			Name:        "nickname",
			Type:        ArgTypeNick,
			Required:    true,
			Description: "The nickname to look up",
			Validator:   validateNonEmpty,
		},
	}
}

func (c Whois) GetFlagSpecs() []Flag {
	return []Flag{}
}

func (c Whois) GetAliases() []string {
	return []string{"w", "who"}
}

func (c Whois) GetContext() CommandContext {
	return ContextConnected
}

func (c Whois) InjectDependencies(*CommandDependencies) {
	return
}

func (c Whois) Execute(_ *ServerManager, window *Window, input string) error {
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
		return fmt.Errorf("incorrect number of arguments: nickname")
	}

	window.connection.SendRaw(fmt.Sprintf("whois :%s", args[0]))
	return nil
}
