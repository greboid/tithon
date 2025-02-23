package irc

import (
	"fmt"
	"log/slog"
	"strings"
	"time"
)

const v3TimestampFormat = "2006-01-02T15:04:05.000Z"

type Message struct {
	timestamp time.Time
	nickname  string
	message   string
	isAction  bool
}

func NewMessage(nickname string, message string) *Message {
	return NewMessageWithTime(time.Now().Format(v3TimestampFormat), nickname, message)
}

func NewMessageWithTime(messageTime string, nickname string, message string) *Message {
	parsedTime, err := time.Parse(v3TimestampFormat, messageTime)
	if err != nil {
		slog.Error("Error parsing time from server", "time", messageTime, "error", err)
		parsedTime = time.Now()
	}
	ircmsg := &Message{
		timestamp: parsedTime,
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
