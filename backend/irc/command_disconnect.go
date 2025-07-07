package irc

import (
	"errors"
	"fmt"
)

type Disconnect struct{}

func init() {
	RegisterCommand(&Disconnect{})
}

func (c Disconnect) GetName() string {
	return "disconnect"
}

func (c Disconnect) GetHelp() string {
	return "Disconnects from the current server"
}

func (c Disconnect) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c Disconnect) GetArgSpecs() []Argument {
	return []Argument{}
}

func (c Disconnect) GetFlagSpecs() []Flag {
	return []Flag{}
}

func (c Disconnect) GetAliases() []string {
	return []string{"dc"}
}

func (c Disconnect) GetContext() CommandContext {
	return ContextConnected
}

func (c Disconnect) InjectDependencies(*CommandDependencies) {
	return
}

func (c Disconnect) Execute(_ *ServerManager, window *Window, input string) error {
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
	return nil
}
