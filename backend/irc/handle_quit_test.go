package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestHandleQuit(t *testing.T) {
	type args struct {
		linkRegex        *regexp.Regexp
		timestampFormat  string
		setPendingUpdate setPendingUpdate
		getChannels      getChannels
	}
	tests := []struct {
		name                    string
		args                    args
		message                 ircmsg.Message
		initialChannels         []*Channel
		wantQuitUser            string
		wantQuitMessage         string
		wantChannelsWithMessage []string
		wantQuitReason          string
	}{
		{
			name: "User quits from multiple channels",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				getChannels: func() []*Channel {
					return []*Channel{
						{
							Window: &Window{
								name:     "#test1",
								users:    []*User{NewUser("user1", ""), NewUser("user2", "")},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						},
						{
							Window: &Window{
								name:     "#test2",
								users:    []*User{NewUser("user1", ""), NewUser("user3", "")},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						},
						{
							Window: &Window{
								name:     "#test3",
								users:    []*User{NewUser("user2", ""), NewUser("user3", "")},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						},
					}
				},
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "QUIT",
				Params:  []string{"", "Leaving", "the", "network"},
			},
			wantQuitUser:            "user1",
			wantQuitMessage:         "user1!user@example.com has quit Leaving the network",
			wantChannelsWithMessage: []string{"#test1", "#test2"},
			wantQuitReason:          "Leaving the network",
		},
		{
			name: "User quits from single channel",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				getChannels: func() []*Channel {
					return []*Channel{
						{
							Window: &Window{
								name:     "#test1",
								users:    []*User{NewUser("user1", ""), NewUser("user2", "")},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						},
						{
							Window: &Window{
								name:     "#test2",
								users:    []*User{NewUser("user2", ""), NewUser("user3", "")},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						},
					}
				},
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "QUIT",
				Params:  []string{"", "Connection", "timeout"},
			},
			wantQuitUser:            "user1",
			wantQuitMessage:         "user1!user@example.com has quit Connection timeout",
			wantChannelsWithMessage: []string{"#test1"},
			wantQuitReason:          "Connection timeout",
		},
		{
			name: "User quits without reason",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				getChannels: func() []*Channel {
					return []*Channel{
						{
							Window: &Window{
								name:     "#test1",
								users:    []*User{NewUser("user1", ""), NewUser("user2", "")},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						},
					}
				},
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "QUIT",
				Params:  []string{""},
			},
			wantQuitUser:            "user1",
			wantQuitMessage:         "user1!user@example.com has quit ",
			wantChannelsWithMessage: []string{"#test1"},
			wantQuitReason:          "",
		},
		{
			name: "User quits from no channels",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				getChannels: func() []*Channel {
					return []*Channel{
						{
							Window: &Window{
								name:     "#test1",
								users:    []*User{NewUser("user2", ""), NewUser("user3", "")},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						},
					}
				},
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "QUIT",
				Params:  []string{"", "Bye"},
			},
			wantQuitUser:            "user1",
			wantChannelsWithMessage: []string{},
		},
		{
			name: "Empty channels list",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				getChannels: func() []*Channel {
					return []*Channel{}
				},
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "QUIT",
				Params:  []string{"", "Goodbye"},
			},
			wantQuitUser:            "user1",
			wantChannelsWithMessage: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pendingUpdateCalled bool
			var channels []*Channel

			setPendingUpdate := func() {
				pendingUpdateCalled = true
			}
			getChannels := func() []*Channel {
				channels = tt.args.getChannels()
				return channels
			}

			handler := HandleQuit(tt.args.linkRegex, tt.args.timestampFormat, setPendingUpdate, getChannels)
			handler(tt.message)

			assert.True(t, pendingUpdateCalled, "setPendingUpdate should have been called")

			channelsWithUser := 0
			channelsWithMessage := 0
			for _, channel := range channels {
				users := channel.GetUsers()
				userFound := false
				for _, user := range users {
					if user.GetNickListDisplay() == tt.wantQuitUser {
						userFound = true
						break
					}
				}
				if userFound {
					channelsWithUser++
				}

				messages := channel.GetMessages()
				if len(messages) > 0 {
					lastMessage := messages[len(messages)-1]
					if lastMessage.GetMessage() == tt.wantQuitMessage {
						channelsWithMessage++
					}
				}
			}

			assert.Equal(t, 0, channelsWithUser, "User should be removed from all channels")

			assert.Equal(t, len(tt.wantChannelsWithMessage), channelsWithMessage, "Quit message should be added to expected channels")

			if len(tt.wantChannelsWithMessage) > 0 {
				for _, expectedChannelName := range tt.wantChannelsWithMessage {
					var foundChannel *Channel
					for _, channel := range channels {
						if channel.GetName() == expectedChannelName {
							foundChannel = channel
							break
						}
					}
					assert.NotNil(t, foundChannel, "Expected channel should exist: "+expectedChannelName)

					messages := foundChannel.GetMessages()
					assert.NotEmpty(t, messages, "Channel should have messages: "+expectedChannelName)

					lastMessage := messages[len(messages)-1]
					assert.Equal(t, tt.wantQuitMessage, lastMessage.GetMessage(), "Quit message should match expected in channel: "+expectedChannelName)
					assert.Equal(t, MessageType(Event), lastMessage.GetType(), "Message type should be Event")
				}
			}
		})
	}
}
