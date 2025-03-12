package irc

type Part struct{}

func (c Part) GetName() string {
	return "part"
}

func (c Part) GetHelp() string {
	return "Parts a channel"
}

func (c Part) Execute(cm *ConnectionManager, server *Connection, channel *Channel, input string) {
	if server == nil || channel == nil {
		return
	}
	server.RemoveChannel(channel.id)
}
