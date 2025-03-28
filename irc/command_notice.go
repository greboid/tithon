package irc

import "time"

type SendNotice struct{}

func (c SendNotice) GetName() string {
	return "notice"
}

func (c SendNotice) GetHelp() string {
	return "Sends a notice"
}

func (c SendNotice) Execute(_ *ConnectionManager, window *Window, input string) error {
	if window == nil {
		return NoServerError
	}
	return window.connection.SendNotice(time.Now(), window.GetID(), input)
}
