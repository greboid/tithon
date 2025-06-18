package irc

type Msg struct{}

func (c Msg) GetName() string {
	return "msg"
}

func (c Msg) GetHelp() string {
	return "Sends a message to a channel"
}

func (c Msg) Execute(_ *ServerManager, window *Window, input string) error {
	if window == nil {
		return NoServerError
	}
	return window.connection.SendMessage(window.GetID(), input)
}
