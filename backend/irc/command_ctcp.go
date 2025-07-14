package irc

import (
	"fmt"
	"strings"
)

type CTCPCommand struct{}

func (c *CTCPCommand) GetName() string {
	return "ctcp"
}

func (c *CTCPCommand) GetHelp() string {
	return "Send a CTCP query to a user. Usage: /ctcp <nick> <command> [parameters]"
}

func (c *CTCPCommand) Execute(sm *ServerManager, window *Window, input string) error {
	if window.connection == nil {
		return NoServerError
	}

	parts := strings.SplitN(input, " ", 3)
	if len(parts) < 2 {
		return fmt.Errorf("usage: /ctcp <nick> <command> [parameters]")
	}

	target := parts[0]
	command := strings.ToUpper(parts[1])
	parameters := ""
	if len(parts) > 2 {
		parameters = parts[2]
	}

	if command != "VERSION" && command != "version" {
		return fmt.Errorf("unknown CTCP message type")
	}
	window.connection.SendRaw(fmt.Sprintf("PRIVMSG %s :%s", target, FormatCTCPReply(command, parameters)))

	displayMessage := fmt.Sprintf("CTCP %s query sent to %s", command, target)
	if parameters != "" {
		displayMessage += fmt.Sprintf(" with parameters: %s", parameters)
	}

	message := NewEvent(EventHelp, sm.timestampFormat, false, displayMessage)
	window.AddMessage(message)

	return nil
}
