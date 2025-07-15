package irc

import (
	"fmt"
)

type Mode struct{}

func (c Mode) GetName() string {
	return "mode"
}

func (c Mode) GetHelp() string {
	return "Sets user or channel modes. Usage: /mode <target> [<mode> [parameters]]"
}

func (c Mode) Execute(_ *ServerManager, window *Window, input string) error {
	if window == nil {
		return ErrNoServer
	}

	if input == "" {
		return fmt.Errorf("mode command at least a target")
	}

	window.connection.SendRaw(fmt.Sprintf("MODE %s", input))
	return nil
}
