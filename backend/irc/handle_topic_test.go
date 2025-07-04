package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestHandleTopic(t *testing.T) {
	type args struct {
		linkRegex        *regexp.Regexp
		timestampFormat  string
		setPendingUpdate setPendingUpdate
		getChannelByName getChannelByName
		getServerName    getServerName
		currentNick      currentNick
	}
	tests := []struct {
		name             string
		args             args
		message          ircmsg.Message
		wantChannelName  string
		wantTopic        string
		wantSetBy        string
		wantMessage      string
		wantChannelError bool
	}{
		{
			name: "Set topic successfully",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name: "#test",
							},
						}, nil
					}
					return nil, assert.AnError
				},
				getServerName: func() string { return "irc.example.com" },
				currentNick:   func() string { return "testnick" },
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "TOPIC",
				Params:  []string{"#test", "Welcome to the test channel"},
			},
			wantChannelName:  "#test",
			wantTopic:        "Welcome to the test channel",
			wantSetBy:        "user1",
			wantMessage:      "user1 changed the topic: Welcome to the test channel",
			wantChannelError: false,
		},
		{
			name: "Unset topic successfully",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name: "#test",
							},
						}, nil
					}
					return nil, assert.AnError
				},
				getServerName: func() string { return "irc.example.com" },
				currentNick:   func() string { return "testnick" },
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "TOPIC",
				Params:  []string{"#test", ""},
			},
			wantChannelName:  "#test",
			wantTopic:        "No Topic set",
			wantSetBy:        "user1",
			wantMessage:      "user1 unset the topic",
			wantChannelError: false,
		},
		{
			name: "Topic change by current user",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name: "#test",
							},
						}, nil
					}
					return nil, assert.AnError
				},
				getServerName: func() string { return "irc.example.com" },
				currentNick:   func() string { return "testnick" },
			},
			message: ircmsg.Message{
				Source:  "testnick!nick@example.com",
				Command: "TOPIC",
				Params:  []string{"#test", "I changed the topic"},
			},
			wantChannelName:  "#test",
			wantTopic:        "I changed the topic",
			wantSetBy:        "testnick",
			wantMessage:      "testnick changed the topic: I changed the topic",
			wantChannelError: false,
		},
		{
			name: "Topic with multiple words",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name: "#test",
							},
						}, nil
					}
					return nil, assert.AnError
				},
				getServerName: func() string { return "irc.example.com" },
				currentNick:   func() string { return "testnick" },
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "TOPIC",
				Params:  []string{"#test", "Welcome", "to", "the", "test", "channel"},
			},
			wantChannelName:  "#test",
			wantTopic:        "Welcome to the test channel",
			wantSetBy:        "user1",
			wantMessage:      "user1 changed the topic: Welcome to the test channel",
			wantChannelError: false,
		},
		{
			name: "Channel not found",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
				},
				getServerName: func() string { return "irc.example.com" },
				currentNick:   func() string { return "testnick" },
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "TOPIC",
				Params:  []string{"#nonexistent", "Some topic"},
			},
			wantChannelName:  "#nonexistent",
			wantChannelError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pendingUpdateCalled bool
			var channel *Channel

			setPendingUpdate := func() {
				pendingUpdateCalled = true
			}

			getChannelByName := func(name string) (*Channel, error) {
				if tt.wantChannelError {
					return nil, assert.AnError
				}
				window := &Window{
					name: name,
				}
				channel = &Channel{
					Window: window,
				}
				return channel, nil
			}

			handler := HandleTopic(tt.args.linkRegex, tt.args.timestampFormat, setPendingUpdate, getChannelByName, tt.args.getServerName, tt.args.currentNick)

			handler(tt.message)

			assert.True(t, pendingUpdateCalled, "setPendingUpdate should have been called")

			if tt.wantChannelError {
				return
			}

			assert.NotNil(t, channel, "Channel should have been retrieved")
			assert.Equal(t, tt.wantChannelName, channel.GetName(), "Channel name should match")

			assert.NotNil(t, channel.GetTopic(), "Topic should have been set")
			assert.Equal(t, tt.wantTopic, channel.GetTopic().GetTopic(), "Topic text should match")
			assert.Equal(t, tt.wantSetBy, channel.GetTopic().GetSetBy(), "Topic setBy should match")

			assert.Equal(t, channel.GetTopic().GetDisplayTopic(), channel.GetTitle(), "Channel title should match topic display")

			messages := channel.GetMessages()
			assert.NotEmpty(t, messages, "At least one message should have been added")
			lastMessage := messages[len(messages)-1]
			assert.Equal(t, tt.wantMessage, lastMessage.GetMessage(), "Last message should match expected")
			assert.Equal(t, MessageType(Event), lastMessage.GetType(), "Message type should be Event")
		})
	}
}
