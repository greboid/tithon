package irc

type Join struct{}

func (c Join) GetName() string {
	return "join"
}

func (c Join) GetHelp() string {
	return "Joins a channel"
}

func (c Join) Execute(cm *ConnectionManager, server *Connection, channel *Channel, input string) {
	if server == nil {
		return
	}
	server.JoinChannel(input, "")
}
