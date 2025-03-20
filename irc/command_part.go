package irc

import (
	"errors"
)

type Part struct{}

func (c Part) GetName() string {
	return "part"
}

func (c Part) GetHelp() string {
	return "Parts a channel"
}

func (c Part) Execute(cm *ConnectionManager, server *Connection, channel *Channel, input string) error {
	if server == nil || channel == nil {
		return errors.New("not on a server or channel")
	}
	server.RemoveChannel(channel.id)
	return nil
}
