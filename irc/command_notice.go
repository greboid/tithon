package irc

type SendNotice struct{}

func (c SendNotice) GetName() string {
	return "notice"
}

func (c SendNotice) GetHelp() string {
	return "Sends a notice"
}

func (c SendNotice) Execute(cm *ConnectionManager, server *Connection, channel *Channel, input string) error {
	if server == nil {
		return NoServerError
	}
	if channel == nil {
		return NoChannelError
	}
	return server.SendNotice(channel.GetID(), input)
}
