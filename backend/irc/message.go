package irc

import (
	"bytes"
	"fmt"
	"github.com/ergochat/irc-go/ircfmt"
	"golang.org/x/net/html"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const v3TimestampFormat = "2006-01-02T15:04:05.000Z"

type MessageType int
type EventType int

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

const (
	EventJoin = iota
	EventPart
	EventQuit
	EventKick
	EventNick
	EventTopic
	EventMode
	EventConnecting
	EventDisconnected
	EventWhois
)

type Message struct {
	timestamp       time.Time
	nickname        string
	message         string
	messageType     MessageType
	highlights      []string
	me              bool
	timestampFormat string
	tags            map[string]string
	nowFunc         func() time.Time
}

func NewNotice(timeFormat string, me bool, nickname string, message string, tags map[string]string, highlights ...string) *Message {
	return newMessage(timeFormat, me, nickname, message, Notice, tags, highlights)
}

func NewEvent(eventType EventType, timeFormat string, me bool, message string) *Message {
	return newMessage(timeFormat, me, "", message, Event, nil, nil)
}

func NewError(timeFormat string, me bool, message string) *Message {
	return newMessage(timeFormat, me, "", message, Error, nil, nil)
}

func NewMessage(timeFormat string, me bool, nickname string, message string, tags map[string]string, highlights ...string) *Message {
	return newMessage(timeFormat, me, nickname, message, Normal, tags, highlights)
}

func newMessage(timeFormat string, me bool, nickname string, message string, messageType MessageType, tags map[string]string, highlights []string) *Message {
	if tags == nil {
		tags = make(map[string]string)
	}
	m := &Message{
		nickname:        nickname,
		message:         message,
		messageType:     messageType,
		highlights:      highlights,
		me:              me,
		timestampFormat: timeFormat,
		tags:            tags,
	}
	return m.parse()
}

func (m *Message) parseTime() {
	if m.nowFunc == nil {
		m.nowFunc = time.Now
	}
	if messageTime := m.tags["time"]; messageTime != "" {
		parsedTime, err := time.Parse(v3TimestampFormat, messageTime)
		if err != nil {
			m.timestamp = m.nowFunc().In(time.Local)
			return
		}
		m.timestamp = parsedTime.In(time.Local)
		return
	}
	m.timestamp = m.nowFunc().In(time.Local)
}

func (m *Message) parse() *Message {
	m.parseTime()
	m.parseAction()
	m.parseHighlight()
	m.parseFormatting()
	return m
}

func (m *Message) parseAction() {
	if strings.HasPrefix(m.message, "\001ACTION") && strings.HasSuffix(m.message, "\001") {
		m.message = strings.TrimPrefix(m.message, "\001ACTION")
		m.message = strings.TrimSuffix(m.message, "\001")
		m.message = strings.TrimSpace(m.message)
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

func (m *Message) GetDisplayMessage() string {
	if m.messageType == Action {
		return fmt.Sprintf(`<span class="%s">%s</span> %s`, m.GetNameColour(), m.nickname, m.message)
	}
	return m.message
}

func (m *Message) GetPlainDisplayMessage() string {
	// TODO: This is awful, I should store the message before I add HTML to it
	node, err := html.Parse(strings.NewReader(m.GetDisplayMessage()))
	if err != nil {
		slog.Error("Error parsing message", "message", m, "error", err)
		return m.GetDisplayMessage()
	}
	var stripper func(node *html.Node, buf *bytes.Buffer)
	stripper = func(node *html.Node, buf *bytes.Buffer) {
		if node.Type == html.TextNode {
			buf.WriteString(node.Data)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			stripper(child, buf)
		}
	}
	stripped := &bytes.Buffer{}
	if m.messageType == Action {
		stripped.WriteString("* ")
	}
	stripper(node, stripped)
	return stripped.String()
}

func (m *Message) GetNickname() string {
	return m.nickname
}

func (m *Message) GetDisplayNickname() string {
	if m.messageType == Action {
		return ""
	}
	return m.nickname
}

func (m *Message) GetNameColour() string {
	if m.me {
		return "mecolour"
	}
	count := 0
	for i := range m.nickname {
		count += int(m.nickname[i])
	}
	count = 1 + (count % 32)
	return "nickcolour" + strconv.Itoa(count)
}

func (m *Message) GetTimestamp() string {
	return m.timestamp.Format(m.timestampFormat)
}

func (m *Message) GetTags() map[string]string {
	return m.tags
}

func (m *Message) isMe() bool {
	return m.me
}

func (m *Message) isHighlight() bool {
	for i := range m.highlights {
		if strings.Contains(strings.ToLower(m.message), strings.ToLower(m.highlights[i])) {
			return true
		}
	}
	return false
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
			classes = append(classes, fmt.Sprintf("bg-%d", split[i].BackgroundColor.Value))
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
