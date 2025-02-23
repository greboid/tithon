package irc

import (
	"fmt"
	"strings"
	"time"
)

type Message struct {
	timestamp time.Time
	nickname  string
	message   string
	isAction  bool
}

func NewMessage(nickname string, message string) *Message {
	ircmsg := &Message{
		timestamp: time.Now(),
		nickname:  nickname,
	}
	if strings.HasPrefix(message, "\001ACTION") && strings.HasSuffix(message, "\001") {
		message = strings.TrimPrefix(message, "\001ACTION")
		message = strings.TrimSuffix(message, "\001")
		ircmsg.isAction = true
	}
	ircmsg.message = message
	return ircmsg
}

func (m *Message) GetMessage() string {
	if m.isAction {
		return fmt.Sprintf("[%s] * %s %s", m.timestamp.Format(time.TimeOnly), m.nickname, m.message)
	}
	return fmt.Sprintf("[%s] <%s> %s", m.timestamp.Format(time.TimeOnly), m.nickname, m.message)
}
