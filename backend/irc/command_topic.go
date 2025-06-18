package irc

import (
	"strings"
)

type ChangeTopic struct{}

func (c ChangeTopic) GetName() string {
	return "topic"
}

func (c ChangeTopic) GetHelp() string {
	return "Changes the topic"
}

func (c ChangeTopic) Execute(_ *ServerManager, window *Window, input string) error {
	if window == nil {
		return NoServerError
	}
	split := strings.SplitN(input, " ", 2)
	if len(split) != 2 {
		return NoChannelError
	}
	return window.connection.connection.Send("TOPIC", split[0], split[1])
}
