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

func (c ChangeTopic) Execute(cm *ConnectionManager, server *Connection, channel *Channel, input string) {
	if server == nil {
		return
	}
	split := strings.SplitN(input, " ", 2)
	if len(split) != 2 {
		return
	}
	server.connection.Send("TOPIC", split[0], split[1])
}
