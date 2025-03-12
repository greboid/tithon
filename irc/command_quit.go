package irc

type Quit struct{}

func (c Quit) GetName() string {
	return "quit"
}

func (c Quit) GetHelp() string {
	return "Quits a server, removing it from the list"
}

func (c Quit) Execute(cm *ConnectionManager, server *Connection, channel *Channel, input string) {
	if server == nil {
		return
	}
	cm.RemoveConnection(server.GetID())
}
