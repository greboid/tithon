package irc

import (
	uniqueid "github.com/albinj12/unique-id"
)

type PrivateMessage struct {
	Window
}

func NewPrivateMessage(connection *Connection, name string) *PrivateMessage {
	s, _ := uniqueid.Generateid("a", 5, "p")
	return &PrivateMessage{
		Window: Window{
			id:         s,
			name:       name,
			messages:   make([]*Message, 0),
			connection: connection,
		},
	}
}
