package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestHandleRPLTopic(t *testing.T) {
	tests := []struct {
		name              string
		message           ircmsg.Message
		channels          []*Channel
		serverName        string
		expectedTopic     string
		expectUpdate      bool
		expectChannelFind bool
	}{
		{
			name: "Set topic for existing channel",
			message: ircmsg.Message{
				Command: "332",
				Params:  []string{"testnick", "#general", "Welcome to the channel!"},
			},
			channels: []*Channel{
				NewChannel(nil, "#general"),
				NewChannel(nil, "#other"),
			},
			serverName:        "irc.example.com",
			expectedTopic:     "Welcome to the channel!",
			expectUpdate:      true,
			expectChannelFind: true,
		},
		{
			name: "Set topic with multiple words",
			message: ircmsg.Message{
				Command: "332",
				Params:  []string{"user", "#test", "This", "is", "a", "multi-word", "topic"},
			},
			channels: []*Channel{
				NewChannel(nil, "#test"),
			},
			serverName:        "irc.test.net",
			expectedTopic:     "This is a multi-word topic",
			expectUpdate:      true,
			expectChannelFind: true,
		},
		{
			name: "Set empty topic",
			message: ircmsg.Message{
				Command: "332",
				Params:  []string{"nick", "#empty"},
			},
			channels: []*Channel{
				NewChannel(nil, "#empty"),
			},
			serverName:        "irc.server.org",
			expectedTopic:     "No Topic set",
			expectUpdate:      true,
			expectChannelFind: true,
		},
		{
			name: "Topic for non-existent channel",
			message: ircmsg.Message{
				Command: "332",
				Params:  []string{"nick", "#nonexistent", "Some topic"},
			},
			channels: []*Channel{
				NewChannel(nil, "#other"),
			},
			serverName:        "irc.example.net",
			expectedTopic:     "",
			expectUpdate:      true,
			expectChannelFind: false,
		},
		{
			name: "Replace existing topic",
			message: ircmsg.Message{
				Command: "332",
				Params:  []string{"nick", "#replace", "New topic"},
			},
			channels: func() []*Channel {
				ch := NewChannel(nil, "#replace")
				ch.SetTopic(NewTopic("Old topic", "olduser", time.Now()))
				return []*Channel{ch}
			}(),
			serverName:        "irc.network.com",
			expectedTopic:     "New topic",
			expectUpdate:      true,
			expectChannelFind: true,
		},
		{
			name: "Topic with special characters",
			message: ircmsg.Message{
				Command: "332",
				Params:  []string{"nick", "#special", "Topic with éspecial çharacters & symbols!"},
			},
			channels: []*Channel{
				NewChannel(nil, "#special"),
			},
			serverName:        "irc.unicode.net",
			expectedTopic:     "Topic with éspecial çharacters & symbols!",
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

			getChannels := func() []*Channel {
				return tt.channels
			}

			handler := HandleRPLTopic(setPendingUpdate, getServerName, getChannels)
			handler(tt.message)

			assert.Equal(t, tt.expectUpdate, pendingUpdateCalled, "setPendingUpdate should be called")

			// Find the channel that should have been updated
			for _, channel := range tt.channels {
				if channel.GetName() == tt.message.Params[1] {
					channelFound = true
					if tt.expectChannelFind {
						topic := channel.GetTopic()
						assert.NotNil(t, topic, "Channel should have a topic set")
						assert.Equal(t, tt.expectedTopic, topic.GetTopic(), "Topic text should match")
						assert.Equal(t, "", topic.GetSetBy(), "SetBy should be empty for RPL_TOPIC")
						assert.True(t, topic.GetSetTime().IsZero(), "SetTime should be zero for RPL_TOPIC")
						assert.Equal(t, tt.expectedTopic, channel.GetTitle(), "Channel title should be updated")
					}
					break
				}
			}

			assert.Equal(t, tt.expectChannelFind, channelFound, "Channel existence should match expectation")
		})
	}
}
