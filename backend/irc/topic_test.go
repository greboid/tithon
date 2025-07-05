package irc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTopic(t *testing.T) {
	testTime := time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		topic    string
		setBy    string
		setTime  time.Time
		expected *Topic
	}{
		{
			name:    "valid topic with all fields",
			topic:   "Welcome to the channel!",
			setBy:   "alice",
			setTime: testTime,
			expected: &Topic{
				topic:   "Welcome to the channel!",
				setBy:   "alice",
				setTime: testTime,
			},
		},
		{
			name:    "empty topic",
			topic:   "",
			setBy:   "bob",
			setTime: testTime,
			expected: &Topic{
				topic:   "",
				setBy:   "bob",
				setTime: testTime,
			},
		},
		{
			name:    "empty setBy",
			topic:   "Test topic",
			setBy:   "",
			setTime: testTime,
			expected: &Topic{
				topic:   "Test topic",
				setBy:   "",
				setTime: testTime,
			},
		},
		{
			name:    "zero time",
			topic:   "Test topic",
			setBy:   "charlie",
			setTime: time.Time{},
			expected: &Topic{
				topic:   "Test topic",
				setBy:   "charlie",
				setTime: time.Time{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewTopic(tt.topic, tt.setBy, tt.setTime)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTopic_GetTopic(t *testing.T) {
	testTime := time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		topic    *Topic
		expected string
	}{
		{
			name: "topic with content",
			topic: &Topic{
				topic:   "Welcome to the channel!",
				setBy:   "alice",
				setTime: testTime,
			},
			expected: "Welcome to the channel!",
		},
		{
			name: "empty topic",
			topic: &Topic{
				topic:   "",
				setBy:   "bob",
				setTime: testTime,
			},
			expected: "No Topic set",
		},
		{
			name: "topic with only spaces",
			topic: &Topic{
				topic:   "   ",
				setBy:   "charlie",
				setTime: testTime,
			},
			expected: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.topic.GetTopic()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTopic_GetSetBy(t *testing.T) {
	testTime := time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		topic    *Topic
		expected string
	}{
		{
			name: "topic with setBy",
			topic: &Topic{
				topic:   "Test topic",
				setBy:   "alice",
				setTime: testTime,
			},
			expected: "alice",
		},
		{
			name: "topic with empty setBy",
			topic: &Topic{
				topic:   "Test topic",
				setBy:   "",
				setTime: testTime,
			},
			expected: "",
		},
		{
			name: "topic with nickname containing special chars",
			topic: &Topic{
				topic:   "Test topic",
				setBy:   "alice|away",
				setTime: testTime,
			},
			expected: "alice|away",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.topic.GetSetBy()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTopic_GetSetTime(t *testing.T) {
	testTime := time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC)
	zeroTime := time.Time{}

	tests := []struct {
		name     string
		topic    *Topic
		expected time.Time
	}{
		{
			name: "topic with valid time",
			topic: &Topic{
				topic:   "Test topic",
				setBy:   "alice",
				setTime: testTime,
			},
			expected: testTime,
		},
		{
			name: "topic with zero time",
			topic: &Topic{
				topic:   "Test topic",
				setBy:   "bob",
				setTime: zeroTime,
			},
			expected: zeroTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.topic.GetSetTime()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTopic_GetDisplayTopic(t *testing.T) {
	testTime := time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC)
	zeroTime := time.Time{}

	tests := []struct {
		name     string
		topic    *Topic
		expected string
	}{
		{
			name: "empty topic",
			topic: &Topic{
				topic:   "",
				setBy:   "alice",
				setTime: testTime,
			},
			expected: "No Topic set",
		},
		{
			name: "topic with no metadata",
			topic: &Topic{
				topic:   "Welcome to the channel!",
				setBy:   "",
				setTime: zeroTime,
			},
			expected: "Welcome to the channel!",
		},
		{
			name: "topic with setBy and setTime",
			topic: &Topic{
				topic:   "Welcome to the channel!",
				setBy:   "alice",
				setTime: testTime,
			},
			expected: "Welcome to the channel! (set by alice on 2023-12-25 10:30:00)",
		},
		{
			name: "topic with setBy only",
			topic: &Topic{
				topic:   "Welcome to the channel!",
				setBy:   "alice",
				setTime: zeroTime,
			},
			expected: "Welcome to the channel! (set by alice)",
		},
		{
			name: "topic with setTime only",
			topic: &Topic{
				topic:   "Welcome to the channel!",
				setBy:   "",
				setTime: testTime,
			},
			expected: "Welcome to the channel! (set on 2023-12-25 10:30:00)",
		},
		{
			name: "topic with special characters",
			topic: &Topic{
				topic:   "Welcome! #test & <script>alert('xss')</script>",
				setBy:   "alice|away",
				setTime: testTime,
			},
			expected: "Welcome! #test & <script>alert('xss')</script> (set by alice|away on 2023-12-25 10:30:00)",
		},
		{
			name: "topic with unicode characters",
			topic: &Topic{
				topic:   "🎉 Welcome to the channel! 🎉",
				setBy:   "alice",
				setTime: testTime,
			},
			expected: "🎉 Welcome to the channel! 🎉 (set by alice on 2023-12-25 10:30:00)",
		},
		{
			name: "very long topic",
			topic: &Topic{
				topic:   "This is a very long topic that might be used to test how the display function handles longer strings and whether it properly formats them with the metadata",
				setBy:   "alice",
				setTime: testTime,
			},
			expected: "This is a very long topic that might be used to test how the display function handles longer strings and whether it properly formats them with the metadata (set by alice on 2023-12-25 10:30:00)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.topic.GetDisplayTopic()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTopicTimeFormat(t *testing.T) {
	testTime := time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC)

	topic := &Topic{
		topic:   "Test topic",
		setBy:   "alice",
		setTime: testTime,
	}

	result := topic.GetDisplayTopic()
	expected := "Test topic (set by alice on 2023-12-25 10:30:45)"
	assert.Equal(t, expected, result)
}
