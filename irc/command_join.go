package irc

import "errors"

type Join struct{}

func (c Join) GetName() string {
	return "join"
}

func (c Join) GetHelp() string {
	return "Joins a channel"
}

func (c Join) Execute(cm *ConnectionManager, server *Connection, channel *Channel, input string) error {
	if server == nil {
		return errors.New("not on a server")
	}
	return server.JoinChannel(input, "")
}
