package irc

import "time"

type Msg struct{}

func (c Msg) GetName() string {
	return "msg"
}

func (c Msg) GetHelp() string {
	return "Performs an action in a channel or private message"
}

func (c Msg) Execute(_ *ConnectionManager, window *Window, input string) error {
	if window == nil {
		return NoServerError
	}
	return window.connection.SendMessage(time.Now(), window.GetID(), input)
}
