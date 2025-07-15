package irc

type Quit struct{}

func (c Quit) GetName() string {
	return "quit"
}

func (c Quit) GetHelp() string {
	return "Quits a server, removing it from the list"
}

func (c Quit) Execute(cm *ServerManager, window *Window, _ string) error {
	if window == nil {
		return ErrNoServer
	}
	cm.RemoveConnection(window.connection.GetID())
	return nil
}
