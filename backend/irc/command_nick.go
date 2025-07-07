package irc

import (
	"fmt"
)

type Nick struct{}

func (c Nick) GetName() string {
	return "nick"
}

func (c Nick) GetHelp() string {
	return "Changes your nickname"
}

func (c Nick) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c Nick) GetArgSpecs() []Argument {
	return []Argument{
		{
			Name:        "nickname",
			Type:        ArgTypeNick,
			Required:    true,
			Description: "The new nickname to use",
			Validator:   validateNonEmpty,
		},
	}
}

func (c Nick) GetFlagSpecs() []Flag {
	return []Flag{}
}

func (c Nick) GetAliases() []string {
	return []string{"n", "nickname"}
}

func (c Nick) GetContext() CommandContext {
	return ContextConnected
}

func (c Nick) Execute(_ *ServerManager, window *Window, input string) error {
	if window == nil {
		return NoServerError
	}
	
	parsed, err := Parse(c, input)
	if err != nil {
		return fmt.Errorf("argument parsing error: %w", err)
	}

	nickname, err := parsed.GetArgString("nickname")
	if err != nil {
		return fmt.Errorf("failed to get nickname: %w", err)
	}

	window.GetServer().SetNick(nickname)
	return nil
}
