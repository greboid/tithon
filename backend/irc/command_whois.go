package irc

import (
	"fmt"
)

type Whois struct{}

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

func (c Whois) Execute(_ *ServerManager, window *Window, input string) error {
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

	window.connection.SendRaw(fmt.Sprintf("whois :%s", nickname))
	return nil
}
