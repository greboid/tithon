package irc

import (
	"errors"
)

type Reconnect struct{}

func (c Reconnect) GetName() string {
	return "reconnect"
}

func (c Reconnect) GetHelp() string {
	return "Reconnects to the current server. Usage: /reconnect"
}

func (c Reconnect) Execute(_ *ServerManager, window *Window, _ string) error {
	if window == nil {
		return errors.New("no window specified")
	}
	connection := window.GetServer()
	if connection == nil {
		return errors.New("not connected to a server")
	}
	connection.Disconnect()
	connection.Connect()

	return nil
}
