package irc

import (
	"fmt"
)

type SendAction struct{}

func (c SendAction) GetName() string {
	return "me"
}

func (c SendAction) GetHelp() string {
	return "Performs an action in a channel or query"
}

func (c SendAction) Execute(_ *ServerManager, window *Window, input string) error {
	if window == nil {
		return ErrNoServer
	}
	input = fmt.Sprintf("\001ACTION %s\001", input)
	return window.connection.SendMessage(window.GetID(), input)
}
