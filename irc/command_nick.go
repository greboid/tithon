package irc

import "errors"

type Nick struct{}

func (c Nick) GetName() string {
	return "nick"
}

func (c Nick) GetHelp() string {
	return "Changes your nickname"
}

func (c Nick) Execute(cm *ConnectionManager, server *Connection, channel *Channel, input string) error {
	if server == nil {
		return errors.New("not on a server")
	}
	server.connection.SetNick(input)
	return nil
}
