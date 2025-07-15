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

func (c Whois) Execute(_ *ServerManager, window *Window, input string) error {
	if window == nil {
		return ErrNoServer
	}
	window.connection.SendRaw(fmt.Sprintf("whois :%s", input))
	return nil
}
