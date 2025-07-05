package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestHandleOtherJoin(t *testing.T) {
	type args struct {
		linkRegex        *regexp.Regexp
		timestampFormat  string
		setPendingUpdate func()
		currentNick      func() string
		getChannelByName func(string) (*Channel, error)
	}
	tests := []struct {
		name             string
		args             args
		message          ircmsg.Message
		wantChannelName  string
		wantChannelError bool
		wantJoinMessage  string
		wantUserAdded    string
		wantSelfJoin     bool
		wantNoParams     bool
	}{
		{
			name: "Other user joins existing channel",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
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
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "JOIN",
				Params:  []string{"#test"},
			},
			wantChannelName:  "#test",
			wantChannelError: false,
			wantJoinMessage:  "user1!user@example.com has joined #test",
			wantUserAdded:    "user1",
		},
		{
			name: "Current user joins (should be ignored)",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
				},
			},
			message: ircmsg.Message{
				Source:  "testnick!nick@example.com",
				Command: "JOIN",
				Params:  []string{"#test"},
			},
			wantSelfJoin: true,
		},
		{
			name: "Join to unknown channel (error case)",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
				},
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "JOIN",
				Params:  []string{"#nonexistent"},
			},
			wantChannelName:  "#nonexistent",
			wantChannelError: true,
		},
		{
			name: "Invalid join message (no params)",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
				},
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "JOIN",
				Params:  []string{},
			},
			wantNoParams: true,
		},
		{
			name: "User with different hostmask joins",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
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
			},
			message: ircmsg.Message{
				Source:  "testuser!~test@hostname.example.com",
				Command: "JOIN",
				Params:  []string{"#test"},
			},
			wantChannelName:  "#test",
			wantChannelError: false,
			wantJoinMessage:  "testuser!~test@hostname.example.com has joined #test",
			wantUserAdded:    "testuser",
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
				channel = &Channel{
					Window: &Window{
						name:     name,
						users:    make([]*User, 0),
						messages: make([]*Message, 0),
						hasUsers: true,
					},
				}
				return channel, nil
			}

			handler := HandleOtherJoin(tt.args.timestampFormat, setPendingUpdate, tt.args.currentNick, getChannelByName)
			handler(tt.message)

			assert.True(t, pendingUpdateCalled, "setPendingUpdate should have been called")

			if tt.wantSelfJoin || tt.wantNoParams || tt.wantChannelError {
				// For self joins, invalid messages, or channel errors, no channel operations should occur
				return
			}

			assert.NotNil(t, channel, "Channel should have been retrieved")
			assert.Equal(t, tt.wantChannelName, channel.GetName(), "Channel name should match")

			if tt.wantUserAdded != "" {
				users := channel.GetUsers()
				assert.NotEmpty(t, users, "At least one user should have been added to channel")

				var foundUser *User
				for _, user := range users {
					if user.GetNickListDisplay() == tt.wantUserAdded {
						foundUser = user
						break
					}
				}
				assert.NotNil(t, foundUser, "The expected user should have been added to channel")
				assert.Equal(t, tt.wantUserAdded, foundUser.GetNickListDisplay(), "Added user nickname should match")
				assert.Empty(t, foundUser.GetNickListModes(), "User should have no modes initially")
			}

			messages := channel.GetMessages()
			assert.NotEmpty(t, messages, "At least one message should have been added to the channel")
			lastMessage := messages[len(messages)-1]
			assert.Equal(t, tt.wantJoinMessage, lastMessage.GetMessage(), "Join message should match expected")
			assert.Equal(t, MessageType(Event), lastMessage.GetType(), "Message type should be Event")
		})
	}
}
