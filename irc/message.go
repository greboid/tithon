package irc

import (
	"html"
	"log/slog"
	"regexp"
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
	ircmsg.message = ircmsg.parseFormatting(message)
	return ircmsg
}

func (m *Message) IsAction() bool {
	return m.isAction
}

func (m *Message) GetMessage() string {
	return m.message
}

func (m *Message) GetNickname() string {
	return m.nickname
}

func (m *Message) GetTimestamp() string {
	return m.timestamp.Format(time.TimeOnly)
}

func (m *Message) parseFormatting(message string) string {
	regex := regexp.MustCompile(`(?P<url>https?://\S+|www\.\S+)`)
	message = html.EscapeString(message)
	output := regex.ReplaceAllString(message, `<a target="_blank" href="$url">$url<a>`)
	return output
}
