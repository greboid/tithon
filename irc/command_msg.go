package irc

type Msg struct{}

func (c Msg) GetName() string {
	return "msg"
}

func (c Msg) GetHelp() string {
	return "Performs an action in a channel or private message"
}

func (c Msg) Execute(_ *ConnectionManager, server *Connection, channel *Channel, input string) error {
	if server == nil {
		return NoServerError
	}
	if channel == nil {
		return NoChannelError
	}
	server.SendMessage(channel.GetID(), input)
	return nil
}
