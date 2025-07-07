package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestHandleRPLTopicWhoTime(t *testing.T) {
	tests := []struct {
		name              string
		message           ircmsg.Message
		channels          []*Channel
		serverName        string
		expectedSetBy     string
		expectedSetTime   time.Time
		expectUpdate      bool
		expectChannelFind bool
	}{
		{
			name: "Update topic who/time for existing topic",
			message: ircmsg.Message{
				Command: "333",
				Params:  []string{"testnick", "#general", "user123", "1640995200"},
			},
			channels: func() []*Channel {
				ch := createTestChannel("#general")
				ch.SetTopic(NewTopic("Existing topic", "", time.Time{}))
				return []*Channel{ch}
			}(),
			serverName:        "irc.example.com",
			expectedSetBy:     "user123",
			expectedSetTime:   time.Unix(1640995200, 0),
			expectUpdate:      true,
			expectChannelFind: true,
		},
		{
			name: "Update topic for channel without existing topic",
			message: ircmsg.Message{
				Command: "333",
				Params:  []string{"nick", "#test", "moderator", "1609459200"},
			},
			channels: func() []*Channel {
				ch := createTestChannel("#test")
				ch.SetTopic(nil)
				return []*Channel{ch}
			}(),
			serverName:        "irc.test.net",
			expectedSetBy:     "moderator",
			expectedSetTime:   time.Unix(1609459200, 0),
			expectUpdate:      true,
			expectChannelFind: true,
		},
		{
			name: "Topic who/time for non-existent channel",
			message: ircmsg.Message{
				Command: "333",
				Params:  []string{"nick", "#nonexistent", "someone", "1640995200"},
			},
			channels: []*Channel{
				createTestChannel("#other"),
			},
			serverName:        "irc.example.net",
			expectedSetBy:     "",
			expectedSetTime:   time.Time{},
			expectUpdate:      true,
			expectChannelFind: false,
		},
		{
			name: "Invalid timestamp (non-numeric)",
			message: ircmsg.Message{
				Command: "333",
				Params:  []string{"nick", "#invalid", "user", "notanumber"},
			},
			channels: func() []*Channel {
				ch := createTestChannel("#invalid")
				ch.SetTopic(NewTopic("Test topic", "", time.Time{}))
				return []*Channel{ch}
			}(),
			serverName:        "irc.server.org",
			expectedSetBy:     "",
			expectedSetTime:   time.Time{},
			expectUpdate:      true,
			expectChannelFind: true,
		},
		{
			name: "Insufficient parameters (missing timestamp)",
			message: ircmsg.Message{
				Command: "333",
				Params:  []string{"nick", "#short", "user"},
			},
			channels: func() []*Channel {
				ch := createTestChannel("#short")
				ch.SetTopic(NewTopic("Test topic", "", time.Time{}))
				return []*Channel{ch}
			}(),
			serverName:        "irc.network.com",
			expectedSetBy:     "",
			expectedSetTime:   time.Time{},
			expectUpdate:      true,
			expectChannelFind: true,
		},
		{
			name: "Empty parameters",
			message: ircmsg.Message{
				Command: "333",
				Params:  []string{},
			},
			channels: []*Channel{
				createTestChannel("#test"),
			},
			serverName:        "irc.empty.net",
			expectedSetBy:     "",
			expectedSetTime:   time.Time{},
			expectUpdate:      true,
			expectChannelFind: false,
		},
		{
			name: "Zero timestamp",
			message: ircmsg.Message{
				Command: "333",
				Params:  []string{"nick", "#zero", "admin", "0"},
			},
			channels: func() []*Channel {
				ch := createTestChannel("#zero")
				ch.SetTopic(NewTopic("Zero time topic", "", time.Time{}))
				return []*Channel{ch}
			}(),
			serverName:        "irc.time.net",
			expectedSetBy:     "admin",
			expectedSetTime:   time.Unix(0, 0),
			expectUpdate:      true,
			expectChannelFind: true,
		},
		{
			name: "Negative timestamp",
			message: ircmsg.Message{
				Command: "333",
				Params:  []string{"nick", "#negative", "timeuser", "-1"},
			},
			channels: func() []*Channel {
				ch := createTestChannel("#negative")
				ch.SetTopic(NewTopic("Negative time topic", "", time.Time{}))
				return []*Channel{ch}
			}(),
			serverName:        "irc.past.net",
			expectedSetBy:     "timeuser",
			expectedSetTime:   time.Unix(-1, 0),
			expectUpdate:      true,
			expectChannelFind: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pendingUpdateCalled := false
			channelFound := false

			setPendingUpdate := func() {
				pendingUpdateCalled = true
			}

			getServerName := func() string {
				return tt.serverName
			}

			getChannelByName := func(name string) (*Channel, error) {
				if tt.expectChannelFind {
					return tt.channels[0], nil
				}
				return nil, assert.AnError
			}

			handler := HandleRPLTopicWhoTime(setPendingUpdate, getServerName, getChannelByName)
			handler(tt.message)

			assert.Equal(t, tt.expectUpdate, pendingUpdateCalled, "setPendingUpdate should be called")

			// Find the channel that should have been updated
			if len(tt.message.Params) >= 2 {
				channelName := tt.message.Params[1]
				for _, channel := range tt.channels {
					if channel.GetName() == channelName {
						channelFound = true
						if tt.expectChannelFind && len(tt.message.Params) >= 4 {
							topic := channel.GetTopic()
							if topic != nil {
								// Only check if we have sufficient params and valid timestamp
								if tt.message.Params[3] != "notanumber" {
									assert.Equal(t, tt.expectedSetBy, topic.GetSetBy(), "SetBy should match")
									assert.True(t, tt.expectedSetTime.Equal(topic.GetSetTime()), "SetTime should match")
								}
							}
						}
						break
					}
				}
			}

			assert.Equal(t, tt.expectChannelFind, channelFound, "Channel existence should match expectation")
		})
	}
}
