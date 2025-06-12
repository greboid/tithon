package irc

type Settings struct {
	showSettings chan bool
}

func (c Settings) GetName() string {
	return "settings"
}

func (c Settings) GetHelp() string {
	return "Shows the settings dialog"
}

func (c Settings) Execute(_ *ConnectionManager, _ *Window, _ string) error {
	c.showSettings <- true
	return nil
}
