package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestHandleNotice(t *testing.T) {
	type args struct {
		linkRegex        *regexp.Regexp
		timestampFormat  string
		setPendingUpdate func()
		currentNick      func() string
		addMessage       func(*Message)
		isValidChannel   func(string) bool
		getChannelByName func(string) (*Channel, error)
		getQueryByName   func(string) (*Query, error)
		addQuery         func(string) *Query
	}
	tests := []struct {
		name              string
		args              args
		message           ircmsg.Message
		wantChannelName   string
		wantQueryName     string
		wantMessage       string
		wantChannelError  bool
		wantQueryError    bool
		wantCreateQuery   bool
		wantServerMessage bool
		wantErrorLog      bool
	}{
		{
			name: "Server notice (empty source)",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				addMessage:       func(msg *Message) {},
				isValidChannel: func(name string) bool {
					return name[0] == '#'
				},
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
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
				Source:  "",
				Command: "NOTICE",
				Params:  []string{"testnick", "Welcome", "to", "the", "network"},
			},
			wantMessage:       "Welcome to the network",
			wantServerMessage: true,
		},
		{
			name: "Server notice (domain source)",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				addMessage:       func(msg *Message) {},
				isValidChannel: func(name string) bool {
					return name[0] == '#'
				},
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
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
				Source:  "irc.example.com",
				Command: "NOTICE",
				Params:  []string{"testnick", "Server", "maintenance", "in", "5", "minutes"},
			},
			wantMessage:       "Server maintenance in 5 minutes",
			wantServerMessage: true,
		},
		{
			name: "Channel notice",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				addMessage:       func(msg *Message) {},
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
				Command: "NOTICE",
				Params:  []string{"#test", "Channel", "notice", "message"},
			},
			wantChannelName: "#test",
			wantMessage:     "Channel notice message",
		},
		{
			name: "Private notice to current user (exact match)",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				addMessage:       func(msg *Message) {},
				isValidChannel: func(name string) bool {
					return name[0] == '#'
				},
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
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
				Command: "NOTICE",
				Params:  []string{"testnick", "Private", "notice", "for", "you"},
			},
			wantQueryName:   "user1",
			wantMessage:     "Private notice for you",
			wantQueryError:  true,
			wantCreateQuery: true,
		},
		{
			name: "Private notice to current user (case insensitive)",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				addMessage:       func(msg *Message) {},
				isValidChannel: func(name string) bool {
					return name[0] == '#'
				},
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
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
				Command: "NOTICE",
				Params:  []string{"TESTNICK", "Case", "insensitive", "notice"},
			},
			wantQueryName:   "user1",
			wantMessage:     "Case insensitive notice",
			wantQueryError:  true,
			wantCreateQuery: true,
		},
		{
			name: "Notice to unknown channel",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				addMessage:       func(msg *Message) {},
				isValidChannel: func(name string) bool {
					return name[0] == '#'
				},
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
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
				Command: "NOTICE",
				Params:  []string{"#nonexistent", "Notice", "to", "unknown", "channel"},
			},
			wantChannelName:  "#nonexistent",
			wantChannelError: true,
			wantErrorLog:     true,
		},
		{
			name: "Notice to existing query",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				addMessage:       func(msg *Message) {},
				isValidChannel: func(name string) bool {
					return name[0] == '#'
				},
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
				},
				getQueryByName: func(name string) (*Query, error) {
					if name == "user1" {
						return &Query{
							Window: &Window{
								name: name,
							},
						}, nil
					}
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
				Command: "NOTICE",
				Params:  []string{"testnick", "Notice", "to", "existing", "query"},
			},
			wantQueryName:   "user1",
			wantMessage:     "Notice to existing query",
			wantQueryError:  false,
			wantCreateQuery: false,
		},
		{
			name: "Unsupported notice target",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				addMessage:       func(msg *Message) {},
				isValidChannel: func(name string) bool {
					return name[0] == '#'
				},
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
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
				Command: "NOTICE",
				Params:  []string{"someothernick", "Notice", "to", "someone", "else"},
			},
			wantErrorLog: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pendingUpdateCalled bool
			var channel *Channel
			var query *Query
			var serverMessage *Message
			var queryCreated bool

			setPendingUpdate := func() {
				pendingUpdateCalled = true
			}
			addMessage := func(msg *Message) {
				serverMessage = msg
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

			handler := HandleNotice(tt.args.tt.args.timestampFormat, setPendingUpdate, tt.args.currentNick, addMessage, tt.args.isValidChannel, getChannelByName, getQueryByName, addQuery)
			handler(tt.message)

			assert.True(t, pendingUpdateCalled, "setPendingUpdate should have been called")

			if tt.wantChannelError && tt.wantChannelName != "" {
				// For channel error cases, we can't test much more since the function returns early
				return
			}

			if tt.wantErrorLog && tt.wantChannelName == "" && tt.wantQueryName == "" && !tt.wantServerMessage {
				// For unsupported target cases, nothing should happen
				return
			}

			if tt.wantServerMessage {
				assert.NotNil(t, serverMessage, "Server message should have been added")
				assert.Equal(t, tt.wantMessage, serverMessage.GetMessage(), "Server message should match expected")
				assert.Equal(t, MessageType(Notice), serverMessage.GetType(), "Message type should be Notice")
				return
			}

			if tt.wantChannelName != "" && !tt.wantChannelError {
				assert.NotNil(t, channel, "Channel should have been retrieved")
				assert.Equal(t, tt.wantChannelName, channel.GetName(), "Channel name should match")

				messages := channel.GetMessages()
				assert.NotEmpty(t, messages, "At least one message should have been added to the channel")
				lastMessage := messages[len(messages)-1]
				assert.Equal(t, tt.wantMessage, lastMessage.GetMessage(), "Channel message should match expected")
				assert.Equal(t, MessageType(Notice), lastMessage.GetType(), "Message type should be Notice")
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
				assert.Equal(t, MessageType(Notice), lastMessage.GetType(), "Message type should be Notice")
			}
		})
	}
}
