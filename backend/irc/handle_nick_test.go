package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestHandleNick(t *testing.T) {
	type args struct {
		linkRegex        *regexp.Regexp
		timestampFormat  string
		setPendingUpdate func()
		currentNick      func() string
		addMessage       func(*Message)
		getChannels      func() []*Channel
	}
	tests := []struct {
		name                    string
		args                    args
		message                 ircmsg.Message
		initialChannels         []*Channel
		wantOldNick             string
		wantNewNick             string
		wantSelfNickChange      bool
		wantServerMessage       string
		wantChannelMessage      string
		wantChannelsWithMessage []string
		wantChannelsWithoutUser []string
	}{
		{
			name: "Current user changes nick",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				addMessage:       func(msg *Message) {},
				getChannels: func() []*Channel {
					return []*Channel{
						{
							Window: &Window{
								name:     "#test1",
								users:    []*User{NewUser("testnick", ""), NewUser("user1", "")},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						},
						{
							Window: &Window{
								name:     "#test2",
								users:    []*User{NewUser("testnick", ""), NewUser("user2", "")},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						},
					}
				},
			},
			message: ircmsg.Message{
				Source:  "testnick!nick@example.com",
				Command: "NICK",
				Params:  []string{"newnick"},
			},
			wantOldNick:             "testnick",
			wantNewNick:             "newnick",
			wantSelfNickChange:      true,
			wantServerMessage:       "Your nickname changed to newnick",
			wantChannelMessage:      "testnick is now known as newnick",
			wantChannelsWithMessage: []string{"#test1", "#test2"},
		},
		{
			name: "Other user changes nick in multiple channels",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				addMessage:       func(msg *Message) {},
				getChannels: func() []*Channel {
					return []*Channel{
						{
							Window: &Window{
								name:     "#test1",
								users:    []*User{NewUser("testnick", ""), NewUser("user1", "")},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						},
						{
							Window: &Window{
								name:     "#test2",
								users:    []*User{NewUser("user1", ""), NewUser("user2", "")},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						},
						{
							Window: &Window{
								name:     "#test3",
								users:    []*User{NewUser("testnick", ""), NewUser("user2", "")},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						},
					}
				},
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "NICK",
				Params:  []string{"newuser1"},
			},
			wantOldNick:             "user1",
			wantNewNick:             "newuser1",
			wantSelfNickChange:      false,
			wantChannelMessage:      "user1 is now known as newuser1",
			wantChannelsWithMessage: []string{"#test1", "#test2"},
		},
		{
			name: "User changes nick in single channel",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				addMessage:       func(msg *Message) {},
				getChannels: func() []*Channel {
					return []*Channel{
						{
							Window: &Window{
								name:     "#test1",
								users:    []*User{NewUser("testnick", ""), NewUser("user1", "")},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						},
						{
							Window: &Window{
								name:     "#test2",
								users:    []*User{NewUser("testnick", ""), NewUser("user2", "")},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						},
					}
				},
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "NICK",
				Params:  []string{"newuser1"},
			},
			wantOldNick:             "user1",
			wantNewNick:             "newuser1",
			wantSelfNickChange:      false,
			wantChannelMessage:      "user1 is now known as newuser1",
			wantChannelsWithMessage: []string{"#test1"},
		},
		{
			name: "User changes nick but not in any channels",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				addMessage:       func(msg *Message) {},
				getChannels: func() []*Channel {
					return []*Channel{
						{
							Window: &Window{
								name:     "#test1",
								users:    []*User{NewUser("testnick", ""), NewUser("user2", "")},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						},
					}
				},
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "NICK",
				Params:  []string{"newuser1"},
			},
			wantOldNick:             "user1",
			wantNewNick:             "newuser1",
			wantSelfNickChange:      false,
			wantChannelsWithMessage: []string{},
		},
		{
			name: "Empty channels list",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				addMessage:       func(msg *Message) {},
				getChannels: func() []*Channel {
					return []*Channel{}
				},
			},
			message: ircmsg.Message{
				Source:  "testnick!nick@example.com",
				Command: "NICK",
				Params:  []string{"newnick"},
			},
			wantOldNick:             "testnick",
			wantNewNick:             "newnick",
			wantSelfNickChange:      true,
			wantServerMessage:       "Your nickname changed to newnick",
			wantChannelsWithMessage: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pendingUpdateCalled bool
			var channels []*Channel
			var serverMessage *Message

			setPendingUpdate := func() {
				pendingUpdateCalled = true
			}
			addMessage := func(msg *Message) {
				serverMessage = msg
			}
			getChannels := func() []*Channel {
				channels = tt.args.getChannels()
				return channels
			}

			handler := HandleNick(tt.args.timestampFormat, setPendingUpdate, tt.args.currentNick, addMessage, getChannels)
			handler(tt.message)

			assert.True(t, pendingUpdateCalled, "setPendingUpdate should have been called")

			if tt.wantSelfNickChange {
				assert.NotNil(t, serverMessage, "Server message should have been added for self nick change")
				assert.Equal(t, tt.wantServerMessage, serverMessage.GetMessage(), "Server message should match expected")
				assert.Equal(t, MessageType(Event), serverMessage.GetType(), "Message type should be Event")
			} else {
				assert.Nil(t, serverMessage, "No server message should be added for other user nick changes")
			}

			channelsWithMessage := 0
			for _, channel := range channels {
				users := channel.GetUsers()
				oldNickFound := false
				newNickFound := false
				for _, user := range users {
					if user.GetNickListDisplay() == tt.wantOldNick {
						oldNickFound = true
					}
					if user.GetNickListDisplay() == tt.wantNewNick {
						newNickFound = true
					}
				}

				userWasInChannel := false
				for _, expectedChannel := range tt.wantChannelsWithMessage {
					if channel.GetName() == expectedChannel {
						userWasInChannel = true
						break
					}
				}

				if userWasInChannel {
					assert.False(t, oldNickFound, "Old nick should not be found in channel: "+channel.GetName())
					assert.True(t, newNickFound, "New nick should be found in channel: "+channel.GetName())
				}

				messages := channel.GetMessages()
				if len(messages) > 0 {
					lastMessage := messages[len(messages)-1]
					if lastMessage.GetMessage() == tt.wantChannelMessage {
						channelsWithMessage++
					}
				}
			}

			assert.Equal(t, len(tt.wantChannelsWithMessage), channelsWithMessage, "Nick change message should be added to expected channels")

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
					assert.Equal(t, tt.wantChannelMessage, lastMessage.GetMessage(), "Nick change message should match expected in channel: "+expectedChannelName)
					assert.Equal(t, MessageType(Event), lastMessage.GetType(), "Message type should be Event")
				}
			}
		})
	}
}
