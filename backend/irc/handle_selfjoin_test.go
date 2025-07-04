package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestHandleSelfJoin(t *testing.T) {
	type args struct {
		linkRegex        *regexp.Regexp
		timestampFormat  string
		setPendingUpdate setPendingUpdate
		currentNick      currentNick
		getChannelByName getChannelByName
		addChannel       addChannel
		hasCapability    hasCapability
		sendRaw          sendRaw
	}
	tests := []struct {
		name                   string
		args                   args
		message                ircmsg.Message
		wantChannelName        string
		wantChannelError       bool
		wantCreateChannel      bool
		wantChathistoryCapable bool
		wantChathistoryCommand string
		wantJoinMessage        string
		wantOtherNick          bool
		wantNoParams           bool
	}{
		{
			name: "Join new channel with chathistory capability",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
				},
				addChannel: func(name string) *Channel {
					return &Channel{
						Window: &Window{
							name: name,
						},
					}
				},
				hasCapability: func(cap string) bool {
					return cap == "draft/chathistory"
				},
				sendRaw: func(command string) {},
			},
			message: ircmsg.Message{
				Source:  "testnick!nick@example.com",
				Command: "JOIN",
				Params:  []string{"#test"},
			},
			wantChannelName:        "#test",
			wantChannelError:       true,
			wantCreateChannel:      true,
			wantChathistoryCapable: true,
			wantChathistoryCommand: "CHATHISTORY LATEST #test * 100",
			wantJoinMessage:        "You have joined #test",
		},
		{
			name: "Join new channel without chathistory capability",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
				},
				addChannel: func(name string) *Channel {
					return &Channel{
						Window: &Window{
							name: name,
						},
					}
				},
				hasCapability: func(cap string) bool {
					return false
				},
				sendRaw: func(command string) {},
			},
			message: ircmsg.Message{
				Source:  "testnick!nick@example.com",
				Command: "JOIN",
				Params:  []string{"#test"},
			},
			wantChannelName:        "#test",
			wantChannelError:       true,
			wantCreateChannel:      true,
			wantChathistoryCapable: false,
			wantJoinMessage:        "You have joined #test",
		},
		{
			name: "Join existing channel",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#existing" {
						return &Channel{
							Window: &Window{
								name: "#existing",
							},
						}, nil
					}
					return nil, assert.AnError
				},
				addChannel: func(name string) *Channel {
					return &Channel{
						Window: &Window{
							name: name,
						},
					}
				},
				hasCapability: func(cap string) bool {
					return true
				},
				sendRaw: func(command string) {},
			},
			message: ircmsg.Message{
				Source:  "testnick!nick@example.com",
				Command: "JOIN",
				Params:  []string{"#existing"},
			},
			wantChannelName:        "#existing",
			wantChannelError:       false,
			wantCreateChannel:      false,
			wantChathistoryCapable: false,
			wantJoinMessage:        "You have joined #existing",
		},
		{
			name: "Join by other user (should be ignored)",
			args: args{
				linkRegex:        regexp.MustCompile(`https?://\S+`),
				timestampFormat:  "15:04:05",
				setPendingUpdate: func() {},
				currentNick:      func() string { return "testnick" },
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
				},
				addChannel: func(name string) *Channel {
					return &Channel{
						Window: &Window{
							name: name,
						},
					}
				},
				hasCapability: func(cap string) bool {
					return false
				},
				sendRaw: func(command string) {},
			},
			message: ircmsg.Message{
				Source:  "otheruser!other@example.com",
				Command: "JOIN",
				Params:  []string{"#test"},
			},
			wantOtherNick: true,
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
				addChannel: func(name string) *Channel {
					return &Channel{
						Window: &Window{
							name: name,
						},
					}
				},
				hasCapability: func(cap string) bool {
					return false
				},
				sendRaw: func(command string) {},
			},
			message: ircmsg.Message{
				Source:  "testnick!nick@example.com",
				Command: "JOIN",
				Params:  []string{},
			},
			wantNoParams: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pendingUpdateCalled bool
			var channel *Channel
			var channelCreated bool
			var chathistoryCommandSent string

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
			addChannel := func(name string) *Channel {
				channelCreated = true
				channel = &Channel{
					Window: &Window{
						name: name,
					},
				}
				return channel
			}
			sendRaw := func(command string) {
				chathistoryCommandSent = command
			}

			handler := HandleSelfJoin(tt.args.linkRegex, tt.args.timestampFormat, setPendingUpdate, tt.args.currentNick, getChannelByName, addChannel, tt.args.hasCapability, sendRaw)
			handler(tt.message)

			assert.True(t, pendingUpdateCalled, "setPendingUpdate should have been called")

			if tt.wantOtherNick || tt.wantNoParams {
				assert.False(t, channelCreated, "No channel should have been created")
				assert.Empty(t, chathistoryCommandSent, "No chathistory command should have been sent")
				return
			}

			if tt.wantCreateChannel {
				assert.True(t, channelCreated, "Channel should have been created")
			} else {
				assert.False(t, channelCreated, "Channel should not have been created")
			}

			assert.NotNil(t, channel, "Channel should exist")
			assert.Equal(t, tt.wantChannelName, channel.GetName(), "Channel name should match")

			if tt.wantChathistoryCapable {
				assert.Equal(t, tt.wantChathistoryCommand, chathistoryCommandSent, "Chathistory command should match expected")
			} else {
				assert.Empty(t, chathistoryCommandSent, "No chathistory command should have been sent")
			}

			messages := channel.GetMessages()
			assert.NotEmpty(t, messages, "At least one message should have been added to the channel")
			lastMessage := messages[len(messages)-1]
			assert.Equal(t, tt.wantJoinMessage, lastMessage.GetMessage(), "Join message should match expected")
			assert.Equal(t, MessageType(Event), lastMessage.GetType(), "Message type should be Event")
		})
	}
}
