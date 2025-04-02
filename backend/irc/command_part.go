package irc

type Part struct{}

func (c Part) GetName() string {
	return "part"
}

func (c Part) GetHelp() string {
	return "Parts a channel"
}

func (c Part) Execute(_ *ConnectionManager, window *Window, _ string) error {
	if window == nil {
		return NoServerError
	}
	window.connection.RemoveChannel(window.GetID())
	return nil
}
