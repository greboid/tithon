package irc

import (
	"errors"
	"fmt"
)

type Reconnect struct{}

func (c Reconnect) GetName() string {
	return "reconnect"
}

func (c Reconnect) GetHelp() string {
	return "Reconnects to the current server"
}

func (c Reconnect) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c Reconnect) GetArgSpecs() []Argument {
	return []Argument{}
}

func (c Reconnect) GetFlagSpecs() []Flag {
	return []Flag{}
}

func (c Reconnect) GetAliases() []string {
	return []string{"rc"}
}

func (c Reconnect) GetContext() CommandContext {
	return ContextConnected
}

func (c Reconnect) Execute(_ *ServerManager, window *Window, input string) error {
	if window == nil {
		return errors.New("no window specified")
	}
	
	_, err := Parse(c, input)
	if err != nil {
		return fmt.Errorf("argument parsing error: %w", err)
	}
	
	connection := window.GetServer()
	if connection == nil {
		return errors.New("not connected to a server")
	}
	connection.Disconnect()
	connection.Connect()

	return nil
}
