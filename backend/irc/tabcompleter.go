package irc

import "fmt"

type TabCompleter interface {
	Complete(input string, position int) (string, int)
}

type NoopTabCompleter struct{}

type PMTabCompleter struct{}

type ChannelTabCompleter struct {
	channel      *Channel
	lastInput    string
	lastPosition int
}

func NewPrivateMessageTabCompleter(pm *PrivateMessage) TabCompleter {
	return &NoopTabCompleter{}
}

func NewConnectionTabCompleter(connection *Connection) TabCompleter {
	return &NoopTabCompleter{}
}

func NewChannelTabCompleter(channel *Channel) TabCompleter {
	return &ChannelTabCompleter{}
}

func (t *NoopTabCompleter) Complete(input string, position int) (string, int) {
	return input, position
}

func (t *ChannelTabCompleter) Complete(input string, position int) (string, int) {
	fmt.Println("Tab completing")
	return input, position
}
