package irc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name       string
		timeFormat string
		me         bool
		nickname   string
		message    string
		tags       map[string]string
		highlights []string
		wantType   MessageType
		wantMsg    string
	}{
		{
			name:       "Basic message",
			timeFormat: "15:04:05",
			me:         false,
			nickname:   "testuser",
			message:    "Hello, world!",
			tags:       nil,
			highlights: nil,
			wantType:   Normal,
			wantMsg:    "Hello, world!",
		},
		{
			name:       "Action message",
			timeFormat: "15:04:05",
			me:         false,
			nickname:   "testuser",
			message:    "\001ACTION waves\001",
			tags:       nil,
			highlights: nil,
			wantType:   Action,
			wantMsg:    "waves",
		},
		{
			name:       "Message with highlight",
			timeFormat: "15:04:05",
			me:         false,
			nickname:   "testuser",
			message:    "Hello, targetuser!",
			tags:       nil,
			highlights: []string{"targetuser"},
			wantType:   Highlight,
			wantMsg:    "Hello, targetuser!",
		},
		{
			name:       "Action with highlight",
			timeFormat: "15:04:05",
			me:         false,
			nickname:   "testuser",
			message:    "\001ACTION waves at targetuser\001",
			tags:       nil,
			highlights: []string{"targetuser"},
			wantType:   HighlightAction,
			wantMsg:    "waves at targetuser",
		},
		{
			name:       "Message with tags",
			timeFormat: "15:04:05",
			me:         false,
			nickname:   "testuser",
			message:    "Hello, world!",
			tags:       map[string]string{"time": "2023-01-01T12:00:00.000Z"},
			highlights: nil,
			wantType:   Normal,
			wantMsg:    "Hello, world!",
		},
		{
			name:       "Message from me",
			timeFormat: "15:04:05",
			me:         true,
			nickname:   "testuser",
			message:    "Hello, world!",
			tags:       nil,
			highlights: nil,
			wantType:   Normal,
			wantMsg:    "Hello, world!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewMessage(tt.timeFormat, tt.me, tt.nickname, tt.message, tt.tags, tt.highlights...)

			assert.NotNil(t, msg, "NewMessage() should not return nil")
			assert.Equal(t, tt.wantType, msg.GetType(), "NewMessage() type mismatch")
			assert.Equal(t, tt.wantMsg, msg.GetMessage(), "NewMessage() message mismatch")
			assert.Equal(t, tt.nickname, msg.GetNickname(), "NewMessage() nickname mismatch")
			assert.Equal(t, tt.me, msg.isMe(), "NewMessage() me flag mismatch")
			assert.NotEmpty(t, msg.GetTimestamp(), "NewMessage() timestamp should not be empty")
		})
	}
}

func TestNewNotice(t *testing.T) {
	tests := []struct {
		name       string
		timeFormat string
		me         bool
		nickname   string
		message    string
		tags       map[string]string
		highlights []string
		wantType   MessageType
	}{
		{
			name:       "Basic notice",
			timeFormat: "15:04:05",
			me:         false,
			nickname:   "testuser",
			message:    "Notice message",
			tags:       nil,
			highlights: nil,
			wantType:   Notice,
		},
		{
			name:       "Notice with highlight",
			timeFormat: "15:04:05",
			me:         false,
			nickname:   "testuser",
			message:    "Notice to targetuser",
			tags:       nil,
			highlights: []string{"targetuser"},
			wantType:   HighlightNotice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewNotice(tt.timeFormat, tt.me, tt.nickname, tt.message, tt.tags, tt.highlights...)

			assert.NotNil(t, msg, "NewNotice() should not return nil")
			assert.Equal(t, tt.wantType, msg.GetType(), "NewNotice() type mismatch")
			assert.Equal(t, tt.message, msg.GetMessage(), "NewNotice() message mismatch")
			assert.Equal(t, tt.nickname, msg.GetNickname(), "NewNotice() nickname mismatch")
			assert.NotEmpty(t, msg.GetTimestamp(), "NewMessage() timestamp should not be empty")
		})
	}
}

func TestNewEvent(t *testing.T) {
	tests := []struct {
		name       string
		timeFormat string
		me         bool
		message    string
		wantType   MessageType
	}{
		{
			name:       "Basic event",
			timeFormat: "15:04:05",
			me:         false,
			message:    "User joined the channel",
			wantType:   Event,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewEvent(EventJoin, tt.timeFormat, tt.me, tt.message)

			assert.NotNil(t, msg, "NewEvent() should not return nil")
			assert.Equal(t, tt.wantType, msg.GetType(), "NewEvent() type mismatch")
			assert.Equal(t, tt.message, msg.GetMessage(), "NewEvent() message mismatch")
			assert.Empty(t, msg.GetNickname(), "NewEvent() nickname should be empty")
			assert.NotEmpty(t, msg.GetTimestamp(), "NewMessage() timestamp should not be empty")
		})
	}
}

func TestNewError(t *testing.T) {
	tests := []struct {
		name       string
		timeFormat string
		me         bool
		message    string
		wantType   MessageType
	}{
		{
			name:       "Basic error",
			timeFormat: "15:04:05",
			me:         false,
			message:    "Server error",
			wantType:   Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewError(tt.timeFormat, tt.me, tt.message)

			assert.NotNil(t, msg, "NewError() should not return nil")
			assert.Equal(t, tt.wantType, msg.GetType(), "NewError() type mismatch")
			assert.Equal(t, tt.message, msg.GetMessage(), "NewError() message mismatch")
			assert.Empty(t, msg.GetNickname(), "NewError() nickname should be empty")
			assert.NotEmpty(t, msg.GetTimestamp(), "NewMessage() timestamp should not be empty")
		})
	}
}

func TestMessage_parseTime(t *testing.T) {
	setTime := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	nowTime := time.Date(2023, 5, 4, 3, 2, 1, 0, time.UTC)
	nowFunc := func() time.Time {
		return nowTime
	}
	tests := []struct {
		name     string
		tags     map[string]string
		nowFunc  func() time.Time
		wantTime time.Time
	}{
		{
			name:     "With time tag",
			tags:     map[string]string{"time": setTime.Format(v3TimestampFormat)},
			nowFunc:  nowFunc,
			wantTime: setTime.In(time.Local),
		},
		{
			name:     "Without time tag",
			tags:     map[string]string{},
			nowFunc:  nowFunc,
			wantTime: nowTime.In(time.Local),
		},
		{
			name:     "With invalid time tag",
			tags:     map[string]string{"time": "invalid-time"},
			nowFunc:  nowFunc,
			wantTime: nowTime.In(time.Local),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				tags:    tt.tags,
				nowFunc: tt.nowFunc,
			}
			m.parseTime()
			assert.Equal(t, tt.wantTime, m.timestamp, "parseTime() timestamp mismatch")
		})
	}
}

func TestMessage_parseAction(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		messageType MessageType
		wantType    MessageType
		wantMessage string
	}{
		{
			name:        "Action message",
			message:     "\001ACTION waves\001",
			messageType: Normal,
			wantType:    Action,
			wantMessage: "waves",
		},
		{
			name:        "Non-action message",
			message:     "Hello, world!",
			messageType: Normal,
			wantType:    Normal,
			wantMessage: "Hello, world!",
		},
		{
			name:        "Incomplete action message",
			message:     "\001ACTION waves",
			messageType: Normal,
			wantType:    Normal,
			wantMessage: "\001ACTION waves",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				message:     tt.message,
				messageType: tt.messageType,
			}
			m.parseAction()
			assert.Equal(t, tt.wantType, m.messageType, "parseAction() type mismatch")
			assert.Equal(t, tt.wantMessage, m.message, "parseAction() message mismatch")
		})
	}
}

func TestMessage_parseHighlight(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		messageType MessageType
		highlights  []string
		wantType    MessageType
	}{
		{
			name:        "Normal message with highlight",
			message:     "Hello, targetuser!",
			messageType: Normal,
			highlights:  []string{"targetuser"},
			wantType:    Highlight,
		},
		{
			name:        "Action message with highlight",
			message:     "waves at targetuser",
			messageType: Action,
			highlights:  []string{"targetuser"},
			wantType:    HighlightAction,
		},
		{
			name:        "Notice message with highlight",
			message:     "Notice to targetuser",
			messageType: Notice,
			highlights:  []string{"targetuser"},
			wantType:    HighlightNotice,
		},
		{
			name:        "Message without highlight",
			message:     "Hello, world!",
			messageType: Normal,
			highlights:  []string{"targetuser"},
			wantType:    Normal,
		},
		{
			name:        "Case insensitive highlight",
			message:     "Hello, TARGETUSER!",
			messageType: Normal,
			highlights:  []string{"targetuser"},
			wantType:    Highlight,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				message:     tt.message,
				messageType: tt.messageType,
				highlights:  tt.highlights,
			}
			m.parseHighlight()
			assert.Equal(t, tt.wantType, m.messageType, "parseHighlight() type mismatch")
		})
	}
}

func TestMessage_GetTypeDisplay(t *testing.T) {
	tests := []struct {
		name        string
		messageType MessageType
		want        string
	}{
		{
			name:        "Normal message",
			messageType: Normal,
			want:        "normal",
		},
		{
			name:        "Action message",
			messageType: Action,
			want:        "action",
		},
		{
			name:        "Notice message",
			messageType: Notice,
			want:        "notice",
		},
		{
			name:        "Event message",
			messageType: Event,
			want:        "event",
		},
		{
			name:        "Error message",
			messageType: Error,
			want:        "error",
		},
		{
			name:        "Highlight message",
			messageType: Highlight,
			want:        "highlight",
		},
		{
			name:        "Highlight action message",
			messageType: HighlightAction,
			want:        "highlight action",
		},
		{
			name:        "Unknown message",
			messageType: 999,
			want:        "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				messageType: tt.messageType,
			}
			assert.Equal(t, tt.want, m.GetTypeDisplay(), "GetTypeDisplay() returned unexpected result")
		})
	}
}

func TestMessage_GetDisplayMessage(t *testing.T) {
	tests := []struct {
		name        string
		messageType MessageType
		nickname    string
		message     string
		want        string
	}{
		{
			name:        "Normal message",
			messageType: Normal,
			nickname:    "testuser",
			message:     "Hello, world!",
			want:        "Hello, world!",
		},
		{
			name:        "Action message",
			messageType: Action,
			nickname:    "testuser",
			message:     "waves",
			want:        `<span class="nickcolour32">testuser</span> waves`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				messageType: tt.messageType,
				nickname:    tt.nickname,
				message:     tt.message,
			}
			assert.Equal(t, tt.want, m.GetDisplayMessage(), "GetDisplayMessage() returned unexpected result")
		})
	}
}

func TestMessage_GetDisplayNickname(t *testing.T) {
	tests := []struct {
		name        string
		messageType MessageType
		nickname    string
		want        string
	}{
		{
			name:        "Normal message",
			messageType: Normal,
			nickname:    "testuser",
			want:        "testuser",
		},
		{
			name:        "Action message",
			messageType: Action,
			nickname:    "testuser",
			want:        "",
		},
		{
			name:        "Notice message",
			messageType: Notice,
			nickname:    "testuser",
			want:        "testuser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				messageType: tt.messageType,
				nickname:    tt.nickname,
			}
			assert.Equal(t, tt.want, m.GetDisplayNickname(), "GetDisplayNickname() returned unexpected result")
		})
	}
}

func TestMessage_GetNameColour(t *testing.T) {
	tests := []struct {
		name     string
		nickname string
		me       bool
		want     string
	}{
		{
			name:     "Message from me",
			nickname: "testuser",
			me:       true,
			want:     "mecolour",
		},
		{
			name:     "Message from other user",
			nickname: "testuser",
			me:       false,
			want:     "nickcolour32",
		},
		{
			name:     "Different nickname",
			nickname: "otheruser",
			me:       false,
			want:     "nickcolour2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				nickname: tt.nickname,
				me:       tt.me,
			}
			assert.Equal(t, tt.want, m.GetNameColour(), "GetNameColour() returned unexpected result")
		})
	}
}

func TestMessage_isHighlight(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		highlights []string
		want       bool
	}{
		{
			name:       "Message with highlight",
			message:    "Hello, targetuser!",
			highlights: []string{"targetuser"},
			want:       true,
		},
		{
			name:       "Message without highlight",
			message:    "Hello, world!",
			highlights: []string{"targetuser"},
			want:       false,
		},
		{
			name:       "Case insensitive highlight",
			message:    "Hello, TARGETUSER!",
			highlights: []string{"targetuser"},
			want:       true,
		},
		{
			name:       "Multiple highlights - first match",
			message:    "Hello, targetuser!",
			highlights: []string{"targetuser", "otheruser"},
			want:       true,
		},
		{
			name:       "Multiple highlights - second match",
			message:    "Hello, otheruser!",
			highlights: []string{"targetuser", "otheruser"},
			want:       true,
		},
		{
			name:       "Multiple highlights - no match",
			message:    "Hello, world!",
			highlights: []string{"targetuser", "otheruser"},
			want:       false,
		},
		{
			name:       "Empty highlights",
			message:    "Hello, targetuser!",
			highlights: []string{},
			want:       false,
		},
		{
			name:       "Nil highlights",
			message:    "Hello, targetuser!",
			highlights: nil,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				message:    tt.message,
				highlights: tt.highlights,
			}
			assert.Equal(t, tt.want, m.isHighlight(), "isHighlight() returned unexpected result")
		})
	}
}

func TestMessage_GetPlainDisplayMessage(t *testing.T) {
	tests := []struct {
		name        string
		messageType MessageType
		nickname    string
		message     string
		want        string
	}{
		{
			name:        "Normal message",
			messageType: Normal,
			nickname:    "testuser",
			message:     "Hello, world!",
			want:        "Hello, world!",
		},
		{
			name:        "Action message",
			messageType: Action,
			nickname:    "testuser",
			message:     "waves",
			want:        "* testuser waves",
		},
		{
			name:        "Message with HTML",
			messageType: Normal,
			nickname:    "testuser",
			message:     "<a href='https://example.com'>Link</a>",
			want:        "Link",
		},
		{
			name:        "Message with HTML, full link",
			messageType: Normal,
			nickname:    "testuser",
			message:     "<a href='https://example.com'>https://example.com</a>",
			want:        "https://example.com",
		},
		{
			name:        "Action message with HTML",
			messageType: Action,
			nickname:    "testuser",
			message:     "<b>waves</b>",
			want:        "* testuser waves",
		},
		{
			name:        "Action message with a nick colour",
			messageType: Action,
			nickname:    "<span class=\"nickcolour12\">testuser</span>",
			message:     "<b>waves</b>",
			want:        "* testuser waves",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				messageType: tt.messageType,
				nickname:    tt.nickname,
				message:     tt.message,
			}
			assert.Equal(t, tt.want, m.GetPlainDisplayMessage(), "GetPlainDisplayMessage() returned unexpected result")
		})
	}
}

func TestMessage_parseFormatting(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    "Plain text without formatting",
			message: "Hello, world!",
			want:    "Hello, world!",
		},
		{
			name:    "Text with HTML special characters",
			message: "Hello <world> & \"friends\"",
			want:    "Hello &lt;world&gt; &amp; &#34;friends&#34;",
		},
		{
			name:    "Text with URL",
			message: "Check out https://example.com",
			want:    "Check out <a target='_blank' href='https://example.com'>https://example.com</a>",
		},
		{
			name:    "Text with URL with subdomain",
			message: "Check out https://test.example.com",
			want:    "Check out <a target='_blank' href='https://test.example.com'>https://test.example.com</a>",
		},
		{
			name:    "Text with URL in a tag",
			message: "Check out <a href=\"https://example.com\">https://example.com</a>",
			want:    "Check out &lt;a href=&#34;<a target='_blank' href='https://example.com'>https://example.com</a>&#34;&gt;<a target='_blank' href='https://example.com'>https://example.com</a>&lt;/a&gt;",
		},
		{
			name:    "Text with multiple URLs",
			message: "Visit https://example.com and http://test.org",
			want:    "Visit <a target='_blank' href='https://example.com'>https://example.com</a> and <a target='_blank' href='http://test.org'>http://test.org</a>",
		},
		{
			name:    "URL with path and query parameters",
			message: "https://example.com/path?param=value&other=123",
			want:    "<a target='_blank' href='https://example.com/path?param=value&amp;other=123'>https://example.com/path?param=value&amp;other=123</a>",
		},
		{
			name:    "URL with special characters in path",
			message: "https://example.com/path-with-[brackets]",
			want:    "<a target='_blank' href='https://example.com/path-with-[brackets]'>https://example.com/path-with-[brackets]</a>",
		},
		{
			name:    "Text with IRC formatting and URL",
			message: "\x02Bold\x02 text with https://example.com link",
			want:    "<span class=\"bold\">Bold</span> text with <a target='_blank' href='https://example.com'>https://example.com</a> link",
		},
		{
			name:    "No protocol in link",
			message: "text with example.com link",
			want:    "text with <a target='_blank' href='example.com'>example.com</a> link",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				message: tt.message,
			}
			m.parseFormatting()
			assert.Equal(t, tt.want, m.message, "parseFormatting() message mismatch")
		})
	}
}

func TestMessage_parseIRCFormatting(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    "Plain text without formatting",
			message: "Hello, world!",
			want:    "Hello, world!",
		},
		{
			name:    "Text with foreground color",
			message: "\x0304Red text\x03",
			want:    "<span class=\"fg-4\">Red text</span>",
		},
		{
			name:    "Text with background color",
			message: "\x0304,02Red text on blue background\x03",
			want:    "<span class=\"fg-4 bg-2\">Red text on blue background</span>",
		},
		{
			name:    "Text with bold formatting",
			message: "\x02Bold text\x02",
			want:    "<span class=\"bold\">Bold text</span>",
		},
		{
			name:    "Text with monospace formatting",
			message: "\x11Monospace text\x11",
			want:    "<span class=\"monospace\">Monospace text</span>",
		},
		{
			name:    "Text with strikethrough formatting",
			message: "\x1eStrikethrough text\x1e",
			want:    "<span class=\"strikethrough\">Strikethrough text</span>",
		},
		{
			name:    "Text with underline formatting",
			message: "\x1fUnderlined text\x1f",
			want:    "<span class=\"underline\">Underlined text</span>",
		},
		{
			name:    "Text with italic formatting",
			message: "\x1dItalic text\x1d",
			want:    "<span class=\"italic\">Italic text</span>",
		},
		{
			name:    "Text with reverse color formatting",
			message: "\x16Reverse color text\x16",
			want:    "<span class=\"reverseColour\">Reverse color text</span>",
		},
		{
			name:    "Text with multiple formatting attributes",
			message: "\x02\x034\x1fBold, red, underlined text\x1f\x03\x02",
			want:    "<span class=\"fg-4 bold underline\">Bold, red, underlined text</span>",
		},
		{
			name:    "Multiple segments with different formatting",
			message: "Normal \x02bold\x02 normal",
			want:    "Normal <span class=\"bold\">bold</span> normal",
		},
		{
			name:    "Nested formatting",
			message: "\x02Bold \x034and red\x03 just bold\x02",
			want:    "<span class=\"bold\">Bold </span><span class=\"fg-4 bold\">and red</span><span class=\"bold\"> just bold</span>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				message: tt.message,
			}
			m.parseIRCFormatting()
			assert.Equal(t, tt.want, m.message, "parseIRCFormatting() message mismatch")
		})
	}
}
