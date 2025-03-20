package irc

import "errors"

type Msg struct{}

func (c Msg) GetName() string {
	return "msg"
}

func (c Msg) GetHelp() string {
	return "Performs an action in a channel or private message"
}

func (c Msg) Execute(_ *ConnectionManager, server *Connection, channel *Channel, input string) error {
	if server == nil || channel == nil {
		return errors.New("not on a server")
	}
	server.SendMessage(channel.GetID(), input)
	return nil
}
