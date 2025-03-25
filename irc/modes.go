package irc

type UserMode struct {
	display string
	mode    string
}

func NewUserMode(display string, mode string) *UserMode {
	return &UserMode{
		display: display,
		mode:    mode,
	}
}
