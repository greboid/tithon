package irc

import (
	"fmt"
)

type Settings struct {
	showSettings chan bool
}

func (c Settings) GetName() string {
	return "settings"
}

func (c Settings) GetHelp() string {
	return "Shows the settings dialog"
}

func (c Settings) GetUsage() string {
	return GenerateDetailedHelp(c)
}

func (c Settings) GetArgSpecs() []Argument {
	return []Argument{}
}

func (c Settings) GetFlagSpecs() []Flag {
	return []Flag{}
}

func (c Settings) Execute(_ *ServerManager, _ *Window, input string) error {
	_, err := Parse(c, input)
	if err != nil {
		return fmt.Errorf("argument parsing error: %w", err)
	}
	
	c.showSettings <- true
	return nil
}
