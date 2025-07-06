package irc

import (
	"fmt"
)

type Notify struct {
	nm Notifier
}

func (c Notify) GetName() string {
	return "notify"
}

func (c Notify) GetHelp() string {
	return "Shows a notification"
}

func (c Notify) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c Notify) GetArgSpecs() []Argument {
	return []Argument{
		{
			Name:        "message",
			Type:        ArgTypeString,
			Required:    true,
			Description: "The notification message to display",
			Validator:   validateNonEmpty,
		},
	}
}

func (c Notify) GetFlagSpecs() []Flag {
	return []Flag{
		{
			Name:        "sound",
			Type:        ArgTypeBool,
			Required:    false,
			Default:     false,
			Description: "Play notification sound",
		},
		{
			Name:        "popup",
			Type:        ArgTypeBool,
			Required:    false,
			Default:     false,
			Description: "Show popup notification",
		},
	}
}

func (c Notify) Execute(_ *ServerManager, window *Window, input string) error {
	if window == nil {
		return NoServerError
	}
	
	parsed, err := Parse(c, input)
	if err != nil {
		return fmt.Errorf("argument parsing error: %w", err)
	}

	message, err := parsed.GetArgString("message")
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	sound, err := parsed.GetFlagBool("sound")
	if err != nil {
		return fmt.Errorf("failed to get sound flag: %w", err)
	}

	popup, err := parsed.GetFlagBool("popup")
	if err != nil {
		return fmt.Errorf("failed to get popup flag: %w", err)
	}

	c.nm.showNotification(Notification{
		Text:  message,
		Sound: sound,
		Popup: popup,
	})
	return nil
}
