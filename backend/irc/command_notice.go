package irc

type SendNotice struct{}

func (c SendNotice) GetName() string {
	return "notice"
}

func (c SendNotice) GetHelp() string {
	return "Sends a notice"
}

func (c SendNotice) Execute(_ *ServerManager, window *Window, input string) error {
	if window == nil {
		return ErrNoServer
	}
	return window.connection.SendNotice(window.GetID(), input)
}
