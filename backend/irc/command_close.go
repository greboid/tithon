package irc

import (
	"errors"
)

type CloseCommand struct{}

func (c CloseCommand) GetName() string {
	return "close"
}

func (c CloseCommand) GetHelp() string {
	return "Closes a window, parting a channel and quitting a server"
}

func (c CloseCommand) Execute(connections *ServerManager, window *Window, _ string) error {
	if window == nil {
		return NoServerError
	}

	if window.IsServer() {
		connections.RemoveConnection(window.GetID())
		return nil
	} else if window.IsChannel() {
		window.connection.RemoveChannel(window.GetID())
		return nil
	} else if window.IsQuery() {
		window.connection.RemoveQuery(window.GetID())
		return nil
	}
	return errors.New("window not found")
}
