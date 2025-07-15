package irc

import (
	"errors"
	"strings"
)

type QueryCommand struct{}

func (c QueryCommand) GetName() string {
	return "query"
}

func (c QueryCommand) GetHelp() string {
	return "Opens a query window with the specified user"
}

func (c QueryCommand) Execute(_ *ServerManager, window *Window, input string) error {
	if window == nil {
		return ErrNoServer
	}

	parts := strings.SplitN(input, " ", 2)
	if len(parts) == 0 || parts[0] == "" {
		return errors.New("no user specified")
	}
	target := parts[0]

	_, err := window.connection.GetQueryByName(target)
	if err != nil {
		window.connection.AddQuery(target)
	}

	if len(parts) > 1 && parts[1] != "" {
		err = window.connection.SendQuery(target, parts[1])
		if err != nil {
			return err
		}
	}

	return nil
}
