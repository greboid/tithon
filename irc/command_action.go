package irc

import (
	"errors"
	"fmt"
)

type SendAction struct{}

func (c SendAction) GetName() string {
	return "me"
}

func (c SendAction) GetHelp() string {
	return "Performs an action in a channel or private message"
}

func (c SendAction) Execute(_ *ConnectionManager, server *Connection, channel *Channel, input string) error {
	if server == nil || channel == nil {
		return errors.New("not on a server or channel")
	}
	input = fmt.Sprintf("\001ACTION %s\001", input)
	return server.SendMessage(channel.GetID(), input)
}
