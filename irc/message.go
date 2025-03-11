package irc

import (
	"fmt"
	"github.com/ergochat/irc-go/ircfmt"
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
	output := html.EscapeString(message)
	output = m.parseIRCFormatting(output)
	regex := regexp.MustCompile(`(?P<url>https?://\S+)`)
	output = regex.ReplaceAllString(output, "<a target='_blank' href='${url}'>${url}</a>")
	return output
}

func (m *Message) parseIRCFormatting(message string) string {
	split := ircfmt.Split(message)
	var out strings.Builder
	for i := range split {
		var classes []string
		if split[i].ForegroundColor.IsSet {
			classes = append(classes, fmt.Sprintf("fg-%d", split[i].ForegroundColor.Value))
		}
		if split[i].BackgroundColor.IsSet {
			classes = append(classes, fmt.Sprintf("bg-%d", split[i].ForegroundColor.Value))
		}
		if split[i].Bold {
			classes = append(classes, "bold")
		}
		if split[i].Monospace {
			classes = append(classes, "monospace")
		}
		if split[i].Strikethrough {
			classes = append(classes, "strikethrough")
		}
		if split[i].Underline {
			classes = append(classes, "underline")
		}
		if split[i].Italic {
			classes = append(classes, "italic")
		}
		if split[i].ReverseColor {
			classes = append(classes, "reverseColour")
		}
		if len(classes) > 0 {
			out.WriteString(`<span class="`)
			out.WriteString(strings.Join(classes, " "))
			out.WriteString(`">`)
		}
		out.WriteString(split[i].Content)
		if len(classes) > 0 {
			out.WriteString("</span>")
		}
	}
	return out.String()
}
