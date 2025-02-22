package irc

import (
	"fmt"
	"time"
)

type Message struct {
	timestamp time.Time
	nickname  string
	message   string
}

func NewMessage(nickname string, message string) *Message {
	return &Message{
		timestamp: time.Now(),
		nickname:  nickname,
		message:   message,
	}
}

func (m *Message) GetMessage() string {
	return fmt.Sprintf("[%s] <%s> %s", m.timestamp.Format(time.TimeOnly), m.nickname, m.message)
}
