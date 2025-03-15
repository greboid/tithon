package irc

type Nick struct{}

func (c Nick) GetName() string {
	return "nick"
}

func (c Nick) GetHelp() string {
	return "Changes your nickname"
}

func (c Nick) Execute(cm *ConnectionManager, server *Connection, channel *Channel, input string) {
	if server == nil {
		return
	}
	server.connection.SetNick(input)
}
