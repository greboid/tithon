package irc

type Notify struct {
	nm Notifier
}

func (c Notify) GetName() string {
	return "notify"
}

func (c Notify) GetHelp() string {
	return "Shows a notification"
}

func (c Notify) Execute(_ *ServerManager, window *Window, input string) error {
	if window == nil {
		return ErrNoServer
	}
	c.nm.showNotification(Notification{
		Text:  input,
		Sound: false,
		Popup: false,
	})
	return nil
}
