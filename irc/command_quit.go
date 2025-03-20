package irc

import "errors"

type Quit struct{}

func (c Quit) GetName() string {
	return "quit"
}

func (c Quit) GetHelp() string {
	return "Quits a server, removing it from the list"
}

func (c Quit) Execute(cm *ConnectionManager, server *Connection, channel *Channel, input string) error {
	if server == nil {
		return errors.New("not on a server")
	}
	cm.RemoveConnection(server.GetID())
	return nil
}
