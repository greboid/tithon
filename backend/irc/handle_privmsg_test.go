package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestHandlePrivMsg(t *testing.T) {
	type args struct {
		linkRegex        *regexp.Regexp
		timestampFormat  string
		setPendingUpdate func()
		isValidChannel   func(string) bool
		getChannelByName func(string) (*Channel, error)
		currentNick      func() string
		getServerName    func() string
		checkAndNotify   func(string, string, string, string) bool
		getQueryByName   func(string) (*Query, error)
		addQuery         func(string) *Query
	}
	tests := []struct {
		name                   string
		args                   args
		message                ircmsg.Message
		wantChannelName        string
		wantQueryName          string
		wantMessage            string
		wantChannelError       bool
		wantQueryError         bool
		wantCreateQuery        bool
		wantNotificationCalled bool
		wantErrorLog           bool
	}{
		{
			name: "Channel message from other user",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				isValidChannel: func(name string) bool {
					return name == "#test"
				},
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
				currentNick:   func() string { return "testnick" },
				getServerName: func() string { return "irc.example.com" },
				checkAndNotify: func(server, target, nick, message string) bool {
					return true
				},
				getQueryByName: func(name string) (*Query, error) {
					return nil, assert.AnError
				},
				addQuery: func(name string) *Query {
					return &Query{
						Window: &Window{
							name: name,
						},
					}
				},
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "PRIVMSG",
				Params:  []string{"#test", "Hello", "everyone!"},
			},
			wantChannelName:        "#test",
			wantMessage:            "Hello everyone!",
			wantChannelError:       false,
			wantNotificationCalled: true,
		},
		{
			name: "Channel message from current user",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				isValidChannel: func(name string) bool {
					return name == "#test"
				},
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
				currentNick:   func() string { return "testnick" },
				getServerName: func() string { return "irc.example.com" },
				checkAndNotify: func(server, target, nick, message string) bool {
					return true
				},
				getQueryByName: func(name string) (*Query, error) {
					return nil, assert.AnError
				},
				addQuery: func(name string) *Query {
					return &Query{
						Window: &Window{
							name: name,
						},
					}
				},
			},
			message: ircmsg.Message{
				Source:  "testnick!nick@example.com",
				Command: "PRIVMSG",
				Params:  []string{"#test", "Hello", "from", "me!"},
			},
			wantChannelName:        "#test",
			wantMessage:            "Hello from me!",
			wantChannelError:       false,
			wantNotificationCalled: false,
		},
		{
			name: "Private message to current user",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				isValidChannel: func(name string) bool {
					return name[0] == '#'
				},
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
				},
				currentNick:   func() string { return "testnick" },
				getServerName: func() string { return "irc.example.com" },
				checkAndNotify: func(server, target, nick, message string) bool {
					return true
				},
				getQueryByName: func(name string) (*Query, error) {
					return nil, assert.AnError
				},
				addQuery: func(name string) *Query {
					return &Query{
						Window: &Window{
							name: name,
						},
					}
				},
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "PRIVMSG",
				Params:  []string{"testnick", "Hello", "privately!"},
			},
			wantQueryName:          "user1",
			wantMessage:            "Hello privately!",
			wantQueryError:         true,
			wantCreateQuery:        true,
			wantNotificationCalled: true,
		},
		{
			name: "Private message from current user",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				isValidChannel: func(name string) bool {
					return name[0] == '#'
				},
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
				},
				currentNick:   func() string { return "testnick" },
				getServerName: func() string { return "irc.example.com" },
				checkAndNotify: func(server, target, nick, message string) bool {
					return true
				},
				getQueryByName: func(name string) (*Query, error) {
					return nil, assert.AnError
				},
				addQuery: func(name string) *Query {
					return &Query{
						Window: &Window{
							name: name,
						},
					}
				},
			},
			message: ircmsg.Message{
				Source:  "testnick!nick@example.com",
				Command: "PRIVMSG",
				Params:  []string{"user1", "Hello", "back!"},
			},
			wantQueryName:          "testnick",
			wantMessage:            "Hello back!",
			wantQueryError:         true,
			wantCreateQuery:        true,
			wantNotificationCalled: false,
		},
		{
			name: "Message with chathistory tag (no notification)",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				isValidChannel: func(name string) bool {
					return name == "#test"
				},
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
				currentNick:   func() string { return "testnick" },
				getServerName: func() string { return "irc.example.com" },
				checkAndNotify: func(server, target, nick, message string) bool {
					return true
				},
				getQueryByName: func(name string) (*Query, error) {
					return nil, assert.AnError
				},
				addQuery: func(name string) *Query {
					return &Query{
						Window: &Window{
							name: name,
						},
					}
				},
			},
			message:                ircmsg.MakeMessage(map[string]string{"chathistory": "true"}, "user1!user@example.com", "PRIVMSG", "#test", "Historical", "message"),
			wantChannelName:        "#test",
			wantMessage:            "Historical message",
			wantChannelError:       false,
			wantNotificationCalled: false,
		},
		{
			name: "Message to unknown channel",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				isValidChannel: func(name string) bool {
					return name[0] == '#'
				},
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
				},
				currentNick:   func() string { return "testnick" },
				getServerName: func() string { return "irc.example.com" },
				checkAndNotify: func(server, target, nick, message string) bool {
					return true
				},
				getQueryByName: func(name string) (*Query, error) {
					return nil, assert.AnError
				},
				addQuery: func(name string) *Query {
					return &Query{
						Window: &Window{
							name: name,
						},
					}
				},
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "PRIVMSG",
				Params:  []string{"#nonexistent", "Hello"},
			},
			wantChannelName:  "#nonexistent",
			wantChannelError: true,
			wantErrorLog:     true,
		},
		{
			name: "Unsupported message target",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				isValidChannel: func(name string) bool {
					return name[0] == '#'
				},
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
				},
				currentNick:   func() string { return "testnick" },
				getServerName: func() string { return "irc.example.com" },
				checkAndNotify: func(server, target, nick, message string) bool {
					return true
				},
				getQueryByName: func(name string) (*Query, error) {
					return nil, assert.AnError
				},
				addQuery: func(name string) *Query {
					return &Query{
						Window: &Window{
							name: name,
						},
					}
				},
			},
			message: ircmsg.Message{
				Source:  "user1!user@example.com",
				Command: "PRIVMSG",
				Params:  []string{"someservice", "Hello"},
			},
			wantErrorLog: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pendingUpdateCalled bool
			var channel *Channel
			var query *Query
			var notificationCalled bool
			var queryCreated bool

			setPendingUpdate := func() {
				pendingUpdateCalled = true
			}
			getChannelByName := func(name string) (*Channel, error) {
				if tt.wantChannelError {
					return nil, assert.AnError
				}
				channel = &Channel{
					Window: &Window{
						name: name,
					},
				}
				return channel, nil
			}
			getQueryByName := func(name string) (*Query, error) {
				if tt.wantQueryError {
					return nil, assert.AnError
				}
				query = &Query{
					Window: &Window{
						name: name,
					},
				}
				return query, nil
			}
			addQuery := func(name string) *Query {
				queryCreated = true
				query = &Query{
					Window: &Window{
						name: name,
					},
				}
				return query
			}
			checkAndNotify := func(server, target, nick, message string) bool {
				notificationCalled = true
				return true
			}

			handler := HandlePrivMsg(tt.args.linkRegex, tt.args.timestampFormat, setPendingUpdate, tt.args.isValidChannel, getChannelByName, tt.args.currentNick, tt.args.getServerName, checkAndNotify, getQueryByName, addQuery)
			handler(tt.message)

			assert.True(t, pendingUpdateCalled, "setPendingUpdate should have been called")

			if tt.wantChannelError && tt.wantChannelName != "" {
				// For channel error cases, we can't test much more since the function returns early
				return
			}

			if tt.wantErrorLog && tt.wantChannelName == "" && tt.wantQueryName == "" {
				// For unsupported target cases, nothing should happen
				return
			}

			if tt.wantChannelName != "" && !tt.wantChannelError {
				assert.NotNil(t, channel, "Channel should have been retrieved")
				assert.Equal(t, tt.wantChannelName, channel.GetName(), "Channel name should match")

				messages := channel.GetMessages()
				assert.NotEmpty(t, messages, "At least one message should have been added to the channel")
				lastMessage := messages[len(messages)-1]
				assert.Equal(t, tt.wantMessage, lastMessage.GetMessage(), "Channel message should match expected")
			}

			if tt.wantQueryName != "" {
				if tt.wantCreateQuery {
					assert.True(t, queryCreated, "Query should have been created")
				}
				assert.NotNil(t, query, "Query should exist")

				messages := query.GetMessages()
				assert.NotEmpty(t, messages, "At least one message should have been added to the query")
				lastMessage := messages[len(messages)-1]
				assert.Equal(t, tt.wantMessage, lastMessage.GetMessage(), "Query message should match expected")
			}

			assert.Equal(t, tt.wantNotificationCalled, notificationCalled, "Notification call state should match expected")
		})
	}
}
