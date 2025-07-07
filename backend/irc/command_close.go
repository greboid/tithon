package irc

import (
	"errors"
	"fmt"
)

type CloseCommand struct{}

func (c CloseCommand) GetName() string {
	return "close"
}

func (c CloseCommand) GetHelp() string {
	return "Closes a window, parting a channel and quitting a server"
}

func (c CloseCommand) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c CloseCommand) GetArgSpecs() []Argument {
	return []Argument{}
}

func (c CloseCommand) GetFlagSpecs() []Flag {
	return []Flag{}
}

func (c CloseCommand) GetAliases() []string {
	return []string{}
}

func (c CloseCommand) GetContext() CommandContext {
	return ContextConnected
}

func (c CloseCommand) Execute(connections *ServerManager, window *Window, input string) error {
	if window == nil {
		return NoServerError
	}

	_, err := Parse(c, input)
	if err != nil {
		return fmt.Errorf("argument parsing error: %w", err)
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