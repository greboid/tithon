package irc

import (
	"time"
	uniqueid "github.com/albinj12/unique-id"
)

type Channel struct {
	*Window
	topic        *Topic
	channelModes []*ChannelMode // Store channel modes
}

func NewChannel(connection ServerInterface, name string) *Channel {
	s, _ := uniqueid.Generateid("a", 5, "c")
	channel := &Channel{
		Window: &Window{
			id:         s,
			name:       name,
			title:      "No topic Set",
			messages:   make([]*Message, 0),
			connection: connection,
			hasUsers:   true,
			isChannel:  true,
		},
		topic:        NewTopic("No topic Set", "", time.Time{}),
		channelModes: make([]*ChannelMode, 0),
	}
	channel.Window.tabCompleter = NewChannelTabCompleter(channel)
	return channel
}

func (c *Channel) SetTopic(topic *Topic) {
	c.topic = topic
}

func (c *Channel) GetTopic() *Topic {
	if c.topic == nil {
		return NewTopic("", "", time.Time{})
	}
	return c.topic
}

func (c *Channel) GetChannelModes() []*ChannelMode {
	return c.channelModes
}

func (c *Channel) GetChannelMode(mode string) *ChannelMode {
	for _, m := range c.channelModes {
		if m.Mode == mode {
			return m
		}
	}
	return nil
}

func (c *Channel) SetChannelMode(modeType rune, mode string, parameter string, set bool) {
	for i, m := range c.channelModes {
		if m.Mode == mode {
			c.channelModes[i].Parameter = parameter
			c.channelModes[i].Set = set
			if !set && modeType == 'A' {
				c.channelModes = append(c.channelModes[:i], c.channelModes[i+1:]...)
			}
			return
		}
	}

	if set || modeType == 'A' {
		c.channelModes = append(c.channelModes, NewChannelMode(modeType, mode, parameter, set))
	}
}

func (c *Channel) RemoveChannelMode(mode string) {
	for i, m := range c.channelModes {
		if m.Mode == mode {
			c.channelModes = append(c.channelModes[:i], c.channelModes[i+1:]...)
			return
		}
	}
}
