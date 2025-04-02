package irc

type Join struct{}

func (c Join) GetName() string {
	return "join"
}

func (c Join) GetHelp() string {
	return "Joins a channel"
}

func (c Join) Execute(_ *ConnectionManager, window *Window, input string) error {
	if window == nil {
		return NoServerError
	}
	return window.connection.JoinChannel(input, "")
}
