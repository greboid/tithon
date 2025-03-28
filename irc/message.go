package irc

import (
	"fmt"
	"github.com/ergochat/irc-go/ircfmt"
	"html"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const v3TimestampFormat = "2006-01-02T15:04:05.000Z"

type MessageType int

const (
	Normal = iota
	Action
	Notice
	Event
	Error

	Highlight
	HighlightAction
	HighlightNotice
)

type Message struct {
	timestampString string
	timestamp       time.Time
	nickname        string
	message         string
	messageType     MessageType
}

func NewNotice(nickname string, message string) *Message {
	return newMessage(time.Now().Format(v3TimestampFormat), nickname, message, Notice)
}
func NewNoticeWithTime(messageTime string, nickname string, message string) *Message {
	return newMessage(messageTime, nickname, message, Notice)
}
func NewEvent(message string) *Message {
	return newMessage(time.Now().Format(v3TimestampFormat), "", message, Event)
}
func NewEventWithTime(messageTime string, message string) *Message {
	return newMessage(messageTime, "", message, Event)
}
func NewError(message string) *Message {
	return newMessage(time.Now().Format(v3TimestampFormat), "", message, Error)
}
func NewErrorWithTime(messageTime string, message string) *Message {
	return newMessage(messageTime, "", message, Error)
}

func NewMessage(nickname string, message string) *Message {
	return newMessage(time.Now().Format(v3TimestampFormat), nickname, message, Normal)
}

func NewMessageWithTime(messageTime string, nickname string, message string) *Message {
	return newMessage(messageTime, nickname, message, Normal)
}

func newMessage(timestamp string, nickname string, message string, messageType MessageType) *Message {
	parsedTime, err := time.Parse(v3TimestampFormat, timestamp)
	if err != nil {
		slog.Error("Error parsing time from server", "time", timestamp, "error", err)
		parsedTime = time.Now()
	}
	m := &Message{
		timestamp:   parsedTime,
		nickname:    nickname,
		message:     message,
		messageType: messageType,
	}
	return m.parse()
}

func (m *Message) parse() *Message {

	m.parseAction()
	m.parseHighlight()
	m.parseFormatting()
	return m
}

func (m *Message) parseAction() {
	if strings.HasPrefix(m.message, "\001ACTION") && strings.HasSuffix(m.message, "\001") {
		m.message = strings.TrimPrefix(m.message, "\001ACTION")
		m.message = strings.TrimSuffix(m.message, "\001")
		m.messageType = Action
	}
}

func (m *Message) parseHighlight() {
	if !m.isHighlight() {
		return
	}
	switch m.messageType {
	case Action:
		m.messageType = HighlightAction
	case Notice:
		m.messageType = HighlightNotice
	case Normal:
		m.messageType = Highlight
	default:
		slog.Debug("No highlight type defined", "message", m)
	}
}

func (m *Message) GetType() MessageType {
	return m.messageType
}

func (m *Message) GetTypeDisplay() string {
	switch m.messageType {
	case Normal:
		return "normal"
	case Notice:
		return "notice"
	case Action:
		return "action"
	case Event:
		return "event"
	case Error:
		return "error"
	case Highlight:
		return "highlight"
	case HighlightAction:
		return "highlight action"
	default:
		return "unknown"
	}
}

func (m *Message) GetMessage() string {
	return m.message
}

func (m *Message) GetNickname() string {
	return m.nickname
}

func (m *Message) GetNameColour() string {
	count := 0
	for i := range m.nickname {
		count += int(m.nickname[i])
	}
	count = 1 + (count % 32)
	return "nickcolour" + strconv.Itoa(count)
}

func (m *Message) GetTimestamp() string {
	return m.timestamp.Format(time.TimeOnly)
}

func (m *Message) isHighlight() bool {
	return strings.Contains(strings.ToLower(m.message), "greboid")
}

func (m *Message) parseFormatting() {
	output := html.EscapeString(m.message)
	urlRegex := regexp.MustCompile(`(?P<url>https?://[A-Za-z0-9-._~:/?#\[\]@!$&'()*+,;%=]+)`)
	output = urlRegex.ReplaceAllString(output, "<a target='_blank' href='${url}'>${url}</a>")
	m.message = output
	m.parseIRCFormatting()
}

func (m *Message) parseIRCFormatting() {
	split := ircfmt.Split(m.message)
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
	m.message = out.String()
}
