package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestHandleKick(t *testing.T) {
	type args struct {
		linkRegex        *regexp.Regexp
		timestampFormat  string
		setPendingUpdate setPendingUpdate
		currentNick      currentNick
		getChannelByName getChannelByName
		removeChannel    removeChannel
		addMessage       addMessage
	}
	tests := []struct {
		name               string
		args               args
		message            ircmsg.Message
		wantChannelName    string
		wantChannelError   bool
		wantSelfKicked     bool
		wantOtherKicked    string
		wantKickMessage    string
		wantServerMessage  string
		wantChannelRemoved string
		wantKickReason     string
	}{
		{
			name: "Current user kicked from channel",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								id:       "channel-id-123",
								name:     "#test",
								users:    []*User{NewUser("testnick", ""), NewUser("user1", "")},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						}, nil
					}
					return nil, assert.AnError
				},
				removeChannel: func(id string) {},
				addMessage:    func(msg *Message) {},
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "KICK",
				Params:  []string{"#test", "testnick", "Goodbye!"},
			},
			wantChannelName:    "#test",
			wantSelfKicked:     true,
			wantChannelRemoved: "channel-id-123",
			wantServerMessage:  "user1!user@example.com has kicked you from #test (Goodbye!)",
			wantKickReason:     "Goodbye!",
		},
		{
			name: "Other user kicked from channel",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								id:       "channel-id-123",
								name:     "#test",
								users:    []*User{NewUser("testnick", ""), NewUser("user1", ""), NewUser("user2", "")},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						}, nil
					}
					return nil, assert.AnError
				},
				removeChannel: func(id string) {},
				addMessage:    func(msg *Message) {},
			},
			message: ircmsg.Message{
				Source:  "testnick!nick@example.com",
				Command: "KICK",
				Params:  []string{"#test", "user1", "No", "spamming", "allowed"},
			},
			wantChannelName: "#test",
			wantOtherKicked: "user1",
			wantKickMessage: "testnick!nick@example.com has kicked user1 from #test (No spamming allowed)",
			wantKickReason:  "No spamming allowed",
		},
		{
			//TODO Should not show brackets with no message
			name: "Kick without reason",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								id:       "channel-id-123",
								name:     "#test",
								users:    []*User{NewUser("testnick", ""), NewUser("user1", "")},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						}, nil
					}
					return nil, assert.AnError
				},
				removeChannel: func(id string) {},
				addMessage:    func(msg *Message) {},
			},
			message: ircmsg.Message{
				Source:  "testnick!nick@example.com",
				Command: "KICK",
				Params:  []string{"#test", "user1"},
			},
			wantChannelName: "#test",
			wantOtherKicked: "user1",
			wantKickMessage: "testnick!nick@example.com has kicked user1 from #test",
			wantKickReason:  "",
		},
		{
			name: "Kick from unknown channel",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
				},
				removeChannel: func(id string) {},
				addMessage:    func(msg *Message) {},
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "KICK",
				Params:  []string{"#nonexistent", "testnick", "Bye"},
			},
			wantChannelName:  "#nonexistent",
			wantChannelError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pendingUpdateCalled bool
			var channel *Channel
			var serverMessage *Message
			var channelRemoved string

			setPendingUpdate := func() {
				pendingUpdateCalled = true
			}
			getChannelByName := func(name string) (*Channel, error) {
				if tt.wantChannelError {
					return nil, assert.AnError
				}
				channel = &Channel{
					Window: &Window{
						id:       "channel-id-123",
						name:     name,
						users:    []*User{NewUser("testnick", ""), NewUser("user1", ""), NewUser("user2", "")},
						messages: make([]*Message, 0),
						hasUsers: true,
					},
				}
				return channel, nil
			}
			removeChannel := func(id string) {
				channelRemoved = id
			}
			addMessage := func(msg *Message) {
				serverMessage = msg
			}

			handler := HandleKick(tt.args.linkRegex, tt.args.timestampFormat, setPendingUpdate, tt.args.currentNick, getChannelByName, removeChannel, addMessage)
			handler(tt.message)

			assert.True(t, pendingUpdateCalled, "setPendingUpdate should have been called")

			if tt.wantChannelError {
				// For channel error cases, we can't test much more since the function returns early
				return
			}

			assert.NotNil(t, channel, "Channel should have been retrieved")
			assert.Equal(t, tt.wantChannelName, channel.GetName(), "Channel name should match")

			if tt.wantSelfKicked {
				assert.Equal(t, tt.wantChannelRemoved, channelRemoved, "Channel should have been removed")
				assert.NotNil(t, serverMessage, "Server message should have been added")
				assert.Equal(t, tt.wantServerMessage, serverMessage.GetMessage(), "Server message should match expected")
				assert.Equal(t, MessageType(Event), serverMessage.GetType(), "Message type should be Event")
				return
			}

			if tt.wantOtherKicked != "" {
				users := channel.GetUsers()
				userFound := false
				for _, user := range users {
					if user.GetNickListDisplay() == tt.wantOtherKicked {
						userFound = true
						break
					}
				}
				assert.False(t, userFound, "Kicked user should be removed from channel")

				messages := channel.GetMessages()
				assert.NotEmpty(t, messages, "At least one message should have been added to the channel")
				lastMessage := messages[len(messages)-1]
				assert.Equal(t, tt.wantKickMessage, lastMessage.GetMessage(), "Kick message should match expected")
				assert.Equal(t, MessageType(Event), lastMessage.GetType(), "Message type should be Event")
			}

			if !tt.wantSelfKicked {
				assert.Empty(t, channelRemoved, "No channel should have been removed for other user kicks")
			}
		})
	}
}
