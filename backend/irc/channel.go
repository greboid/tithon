package irc

import (
	uniqueid "github.com/albinj12/unique-id"
)

type Channel struct {
	*Window
	topic *Topic
	users []*User
}

func NewChannel(connection *Connection, name string) *Channel {
	s, _ := uniqueid.Generateid("a", 5, "c")
	channel := &Channel{
		Window: &Window{
			id:         s,
			name:       name,
			title:      "No topic Set",
			messages:   make([]*Message, 0),
			connection: connection,
			hasUsers:   true,
		},
		topic: NewTopic("No topic Set"),
		users: nil,
	}
	return channel
}

func (c *Channel) SetTopic(topic *Topic) {
	c.topic = topic
}

func (c *Channel) GetTopic() *Topic {
	if c.topic == nil {
		return NewTopic("")
	}
	return c.topic
}
