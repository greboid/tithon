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

// ChannelMode represents a channel mode
// IRC has four types of channel modes:
// - Type A: Modes that add or remove an address to/from a list (e.g., ban lists)
// - Type B: Modes that change a setting with a parameter (e.g., channel key)
// - Type C: Modes that change a setting only when set (e.g., channel limit)
// - Type D: Modes that change a simple boolean setting (e.g., invite-only)
type ChannelMode struct {
	Type      rune   // 'A', 'B', 'C', or 'D'
	Mode      string // The mode character
	Parameter string // The parameter for the mode (if applicable)
	Set       bool   // Whether the mode is set or not
}

func NewChannelMode(modeType rune, mode string, parameter string, set bool) *ChannelMode {
	return &ChannelMode{
		Type:      modeType,
		Mode:      mode,
		Parameter: parameter,
		Set:       set,
	}
}
