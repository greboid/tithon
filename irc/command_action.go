package irc

import (
	"fmt"
)

type Action struct{}

func (c Action) GetName() string {
	return "me"
}

func (c Action) GetHelp() string {
	return "Performs an action in a channel or private message"
}

func (c Action) Execute(_ *ConnectionManager, server *Connection, channel *Channel, input string) {
	if server == nil || channel == nil {
		return
	}
	input = fmt.Sprintf("\001ACTION %s\001", input)
	server.SendMessage(channel.GetID(), input)
}
