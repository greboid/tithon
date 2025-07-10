package irc

import (
	"errors"
)

type Disconnect struct{}

func (c Disconnect) GetName() string {
	return "disconnect"
}

func (c Disconnect) GetHelp() string {
	return "Disconnects from the current server. Usage: /disconnect"
}

func (c Disconnect) Execute(_ *ServerManager, window *Window, _ string) error {
	if window == nil {
		return errors.New("no window specified")
	}

	connection := window.GetServer()
	if connection == nil {
		return errors.New("not connected to a server")
	}

	connection.Disconnect()
	return nil
}
