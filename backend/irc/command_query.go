package irc

import (
	"errors"
	"strings"
)

type Query struct{}

func (c Query) GetName() string {
	return "query"
}

func (c Query) GetHelp() string {
	return "Opens a private message window with the specified user"
}

func (c Query) Execute(_ *ConnectionManager, window *Window, input string) error {
	if window == nil {
		return NoServerError
	}

	parts := strings.SplitN(input, " ", 2)
	if len(parts) == 0 || parts[0] == "" {
		return errors.New("no user specified")
	}
	target := parts[0]

	_, err := window.connection.GetPrivateMessageByName(target)
	if err != nil {
		window.connection.AddPrivateMessage(target)
	}

	if len(parts) > 1 && parts[1] != "" {
		err = window.connection.SendPrivateMessage(target, parts[1])
		if err != nil {
			return err
		}
	}

	return nil
}
