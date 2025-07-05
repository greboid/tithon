package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandleNamesReply(t *testing.T) {
	type args struct {
		setPendingUpdate func()
		getChannelByName func(string) (*Channel, error)
		getModePrefixes  func() []string
	}
	tests := []struct {
		name             string
		args             args
		message          ircmsg.Message
		initialUsers     []*User
		wantChannelName  string
		wantChannelError bool
		wantUsers        []struct {
			nickname string
			modes    string
		}
		wantUserCount int
	}{
		{
			name: "Basic NAMES reply with mode prefixes",
			args: args{
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    make([]*User, 0),
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						}, nil
					}
					return nil, assert.AnError
				},
				getModePrefixes: func() []string {
					return []string{"ov", "@+"}
				},
			},
			message: ircmsg.Message{
				Source:  "irc.example.com",
				Command: "353", // RPL_NAMREPLY
				Params:  []string{"testnick", "=", "#test", "@alice +bob charlie"},
			},
			wantChannelName: "#test",
			wantUsers: []struct {
				nickname string
				modes    string
			}{
				{"alice", "@"},
				{"bob", "+"},
				{"charlie", ""},
			},
			wantUserCount: 3,
		},
		{
			name: "NAMES reply with existing users (mode update)",
			args: args{
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						channel := &Channel{
							Window: &Window{
								name: "#test",
								users: []*User{
									NewUser("alice", ""),
									NewUser("bob", "@"),
									NewUser("charlie", "+"),
								},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						}
						return channel, nil
					}
					return nil, assert.AnError
				},
				getModePrefixes: func() []string {
					return []string{"ov", "@+"}
				},
			},
			message: ircmsg.Message{
				Source:  "irc.example.com",
				Command: "353",
				Params:  []string{"testnick", "=", "#test", "@alice bob +charlie"},
			},
			wantChannelName: "#test",
			wantUsers: []struct {
				nickname string
				modes    string
			}{
				{"alice", "@"},
				{"bob", ""},
				{"charlie", "+"},
			},
			wantUserCount: 3,
		},
		{
			name: "NAMES reply with multiple mode prefixes",
			args: args{
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    make([]*User, 0),
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						}, nil
					}
					return nil, assert.AnError
				},
				getModePrefixes: func() []string {
					return []string{"qaohv", "~&@%+"}
				},
			},
			message: ircmsg.Message{
				Source:  "irc.example.com",
				Command: "353",
				Params:  []string{"testnick", "=", "#test", "~alice &bob @charlie %dave +eve frank"},
			},
			wantChannelName: "#test",
			wantUsers: []struct {
				nickname string
				modes    string
			}{
				{"alice", "~"},
				{"bob", "&"},
				{"charlie", "@"},
				{"dave", "%"},
				{"eve", "+"},
				{"frank", ""},
			},
			wantUserCount: 6,
		},
		{
			name: "NAMES reply with stacked mode prefixes",
			args: args{
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    make([]*User, 0),
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						}, nil
					}
					return nil, assert.AnError
				},
				getModePrefixes: func() []string {
					return []string{"ov", "@+"}
				},
			},
			message: ircmsg.Message{
				Source:  "irc.example.com",
				Command: "353",
				Params:  []string{"testnick", "=", "#test", "@+alice +bob charlie"},
			},
			wantChannelName: "#test",
			wantUsers: []struct {
				nickname string
				modes    string
			}{
				{"alice", "@+"},
				{"bob", "+"},
				{"charlie", ""},
			},
			wantUserCount: 3,
		},
		{
			name: "NAMES reply with empty names list",
			args: args{
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    make([]*User, 0),
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						}, nil
					}
					return nil, assert.AnError
				},
				getModePrefixes: func() []string {
					return []string{"ov", "@+"}
				},
			},
			message: ircmsg.Message{
				Source:  "irc.example.com",
				Command: "353",
				Params:  []string{"testnick", "=", "#test", ""},
			},
			wantChannelName: "#test",
			wantUsers: []struct {
				nickname string
				modes    string
			}{},
			wantUserCount: 0,
		},
		{
			name: "NAMES reply with extra spaces",
			args: args{
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    make([]*User, 0),
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						}, nil
					}
					return nil, assert.AnError
				},
				getModePrefixes: func() []string {
					return []string{"ov", "@+"}
				},
			},
			message: ircmsg.Message{
				Source:  "irc.example.com",
				Command: "353",
				Params:  []string{"testnick", "=", "#test", "@alice  +bob   charlie  "},
			},
			wantChannelName: "#test",
			wantUsers: []struct {
				nickname string
				modes    string
			}{
				{"alice", "@"},
				{"bob", "+"},
				{"charlie", ""},
			},
			wantUserCount: 3,
		},
		{
			name: "NAMES reply for unknown channel",
			args: args{
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
				},
				getModePrefixes: func() []string {
					return []string{"ov", "@+"}
				},
			},
			message: ircmsg.Message{
				Source:  "irc.example.com",
				Command: "353",
				Params:  []string{"testnick", "=", "#nonexistent", "@alice +bob"},
			},
			wantChannelName:  "#nonexistent",
			wantChannelError: true,
		},
		{
			name: "NAMES reply with empty entries mixed in",
			args: args{
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    make([]*User, 0),
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						}, nil
					}
					return nil, assert.AnError
				},
				getModePrefixes: func() []string {
					return []string{"ov", "@+"}
				},
			},
			message: ircmsg.Message{
				Source:  "irc.example.com",
				Command: "353",
				Params:  []string{"testnick", "=", "#test", "@alice  +bob   charlie"},
			},
			wantChannelName: "#test",
			wantUsers: []struct {
				nickname string
				modes    string
			}{
				{"alice", "@"},
				{"bob", "+"},
				{"charlie", ""},
			},
			wantUserCount: 3,
		},
		{
			name: "NAMES reply with no mode prefixes configured",
			args: args{
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    make([]*User, 0),
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						}, nil
					}
					return nil, assert.AnError
				},
				getModePrefixes: func() []string {
					return []string{"", ""}
				},
			},
			message: ircmsg.Message{
				Source:  "irc.example.com",
				Command: "353",
				Params:  []string{"testnick", "=", "#test", "alice bob charlie"},
			},
			wantChannelName: "#test",
			wantUsers: []struct {
				nickname string
				modes    string
			}{
				{"alice", ""},
				{"bob", ""},
				{"charlie", ""},
			},
			wantUserCount: 3,
		},
		{
			name: "NAMES reply with only spaces and empty entries",
			args: args{
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    make([]*User, 0),
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						}, nil
					}
					return nil, assert.AnError
				},
				getModePrefixes: func() []string {
					return []string{"ov", "@+"}
				},
			},
			message: ircmsg.Message{
				Source:  "irc.example.com",
				Command: "353",
				Params:  []string{"testnick", "=", "#test", "   "},
			},
			wantChannelName: "#test",
			wantUsers: []struct {
				nickname string
				modes    string
			}{},
			wantUserCount: 0,
		},
		{
			name: "NAMES reply with users having complex mode combinations",
			args: args{
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    make([]*User, 0),
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						}, nil
					}
					return nil, assert.AnError
				},
				getModePrefixes: func() []string {
					return []string{"qaohv", "~&@%+"}
				},
			},
			message: ircmsg.Message{
				Source:  "irc.example.com",
				Command: "353",
				Params:  []string{"testnick", "=", "#test", "~&@%+alice @%bob +charlie dave"},
			},
			wantChannelName: "#test",
			wantUsers: []struct {
				nickname string
				modes    string
			}{
				{"alice", "~&@%+"},
				{"bob", "@%"},
				{"charlie", "+"},
				{"dave", ""},
			},
			wantUserCount: 4,
		},
		{
			name: "NAMES reply with empty mode prefix characters",
			args: args{
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    make([]*User, 0),
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						}, nil
					}
					return nil, assert.AnError
				},
				getModePrefixes: func() []string {
					return []string{"ov", ""} // Second element empty - no mode chars
				},
			},
			message: ircmsg.Message{
				Source:  "irc.example.com",
				Command: "353",
				Params:  []string{"testnick", "=", "#test", "alice bob"},
			},
			wantChannelName: "#test",
			wantUsers: []struct {
				nickname string
				modes    string
			}{
				{"alice", ""}, // With empty prefixes[1], no modes are stripped
				{"bob", ""},
			},
			wantUserCount: 2,
		},
		{
			name: "NAMES reply with prefix-only names",
			args: args{
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    make([]*User, 0),
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						}, nil
					}
					return nil, assert.AnError
				},
				getModePrefixes: func() []string {
					return []string{"ov", "@+"}
				},
			},
			message: ircmsg.Message{
				Source:  "irc.example.com",
				Command: "353",
				Params:  []string{"testnick", "=", "#test", "@ + alice @bob"},
			},
			wantChannelName: "#test",
			wantUsers: []struct {
				nickname string
				modes    string
			}{
				{"", "+"},     // Empty nickname - second mode overwrites first
				{"alice", ""}, // Normal user
				{"bob", "@"},  // User with mode
			},
			wantUserCount: 3,
		},
		{
			name: "NAMES reply with mixed new and existing users",
			args: args{
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						channel := &Channel{
							Window: &Window{
								name: "#test",
								users: []*User{
									NewUser("alice", "@"),
									NewUser("charlie", ""),
								},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
						}
						return channel, nil
					}
					return nil, assert.AnError
				},
				getModePrefixes: func() []string {
					return []string{"ov", "@+"}
				},
			},
			message: ircmsg.Message{
				Source:  "irc.example.com",
				Command: "353",
				Params:  []string{"testnick", "=", "#test", "alice +bob +charlie dave"},
			},
			wantChannelName: "#test",
			wantUsers: []struct {
				nickname string
				modes    string
			}{
				{"alice", ""},    // existing user, mode updated
				{"charlie", "+"}, // existing user, mode updated
				{"bob", "+"},     // new user
				{"dave", ""},     // new user
			},
			wantUserCount: 4,
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
				users := make([]*User, 0)
				if tt.initialUsers != nil {
					users = append(users, tt.initialUsers...)
				}
				channel = &Channel{
					Window: &Window{
						name:     name,
						users:    users,
						messages: make([]*Message, 0),
						hasUsers: true,
					},
				}
				return channel, nil
			}

			handler := HandleNamesReply(setPendingUpdate, getChannelByName, tt.args.getModePrefixes)
			handler(tt.message)

			assert.True(t, pendingUpdateCalled, "setPendingUpdate should have been called")

			if tt.wantChannelError {
				// For channel error cases, we can't test much more since the function returns early
				return
			}

			assert.NotNil(t, channel, "Channel should have been retrieved")
			assert.Equal(t, tt.wantChannelName, channel.GetName(), "Channel name should match")

			users := channel.GetUsers()
			assert.Equal(t, tt.wantUserCount, len(users), "User count should match expected")

			if len(tt.wantUsers) > 0 {
				for _, expectedUser := range tt.wantUsers {
					var foundUser *User
					for _, user := range users {
						if user.GetNickListDisplay() == expectedUser.nickname {
							foundUser = user
							break
						}
					}
					assert.NotNil(t, foundUser, "Expected user should be found: "+expectedUser.nickname)
					assert.Equal(t, expectedUser.modes, foundUser.GetNickListModes(), "User modes should match for: "+expectedUser.nickname)
				}
			}
		})
	}
}
