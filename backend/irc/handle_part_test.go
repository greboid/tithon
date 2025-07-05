package irc

import (
	"errors"
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestHandlePart(t *testing.T) {
	linkRegex := regexp.MustCompile(`https?://[^\s]+`)
	timestampFormat := "15:04:05"

	tests := []struct {
		name                 string
		message              ircmsg.Message
		currentNick          string
		channelToReturn      *Channel
		channelError         error
		expectedUserCount    int
		expectChannelRemoval bool
		expectUpdate         bool
		expectMessage        bool
	}{
		{
			name: "Current user parts channel",
			message: ircmsg.Message{
				Source:  "testnick!user@example.com",
				Command: "PART",
				Params:  []string{"#general", "Goodbye!"},
			},
			currentNick: "testnick",
			channelToReturn: &Channel{
				Window: &Window{
					id:       "channel1",
					name:     "#general",
					users:    []*User{NewUser("testnick", ""), NewUser("other", "")},
					messages: []*Message{},
					hasUsers: true,
				},
			},
			channelError:         nil,
			expectedUserCount:    2, // Handler returns early, so user list isn't modified
			expectChannelRemoval: true,
			expectUpdate:         true,
			expectMessage:        false,
		},
		{
			name: "Other user parts channel",
			message: ircmsg.Message{
				Source:  "othernick!other@example.com",
				Command: "PART",
				Params:  []string{"#general", "See you later"},
			},
			currentNick: "testnick",
			channelToReturn: &Channel{
				Window: &Window{
					id:       "channel1",
					name:     "#general",
					users:    []*User{NewUser("testnick", ""), NewUser("othernick", ""), NewUser("someone", "")},
					messages: []*Message{},
					hasUsers: true,
				},
			},
			channelError:         nil,
			expectedUserCount:    2,
			expectChannelRemoval: false,
			expectUpdate:         true,
			expectMessage:        true,
		},
		{
			name: "User parts from unknown channel",
			message: ircmsg.Message{
				Source:  "testnick!user@example.com",
				Command: "PART",
				Params:  []string{"#nonexistent"},
			},
			currentNick:          "testnick",
			channelToReturn:      nil,
			channelError:         errors.New("channel not found"),
			expectedUserCount:    0,
			expectChannelRemoval: false,
			expectUpdate:         true,
			expectMessage:        false,
		},
		{
			name: "Part without reason",
			message: ircmsg.Message{
				Source:  "someone!user@host.com",
				Command: "PART",
				Params:  []string{"#test"},
			},
			currentNick: "testnick",
			channelToReturn: &Channel{
				Window: &Window{
					id:       "channel2",
					name:     "#test",
					users:    []*User{NewUser("testnick", ""), NewUser("someone", "")},
					messages: []*Message{},
					hasUsers: true,
				},
			},
			channelError:         nil,
			expectedUserCount:    1,
			expectChannelRemoval: false,
			expectUpdate:         true,
			expectMessage:        true,
		},
		{
			name: "User not in channel user list",
			message: ircmsg.Message{
				Source:  "stranger!user@unknown.com",
				Command: "PART",
				Params:  []string{"#test"},
			},
			currentNick: "testnick",
			channelToReturn: &Channel{
				Window: &Window{
					id:       "channel3",
					name:     "#test",
					users:    []*User{NewUser("testnick", ""), NewUser("regular", "")},
					messages: []*Message{},
					hasUsers: true,
				},
			},
			channelError:         nil,
			expectedUserCount:    2,
			expectChannelRemoval: false,
			expectUpdate:         true,
			expectMessage:        true,
		},
		{
			name: "Current user parts channel with long nick",
			message: ircmsg.Message{
				Source:  "verylongnickname!user@example.com",
				Command: "PART",
				Params:  []string{"#longtest"},
			},
			currentNick: "verylongnickname",
			channelToReturn: &Channel{
				Window: &Window{
					id:       "channel4",
					name:     "#longtest",
					users:    []*User{NewUser("verylongnickname", "")},
					messages: []*Message{},
					hasUsers: true,
				},
			},
			channelError:         nil,
			expectedUserCount:    1, // Handler returns early, so user list isn't modified
			expectChannelRemoval: true,
			expectUpdate:         true,
			expectMessage:        false,
		},
		{
			name: "Case sensitive nick comparison",
			message: ircmsg.Message{
				Source:  "TestNick!user@example.com",
				Command: "PART",
				Params:  []string{"#case"},
			},
			currentNick: "testnick",
			channelToReturn: &Channel{
				Window: &Window{
					id:       "channel5",
					name:     "#case",
					users:    []*User{NewUser("testnick", ""), NewUser("TestNick", "")},
					messages: []*Message{},
					hasUsers: true,
				},
			},
			channelError:         nil,
			expectedUserCount:    1,
			expectChannelRemoval: false,
			expectUpdate:         true,
			expectMessage:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pendingUpdateCalled := false
			channelRemoved := ""

			setPendingUpdate := func() {
				pendingUpdateCalled = true
			}

			currentNick := func() string {
				return tt.currentNick
			}

			getChannelByName := func(name string) (*Channel, error) {
				if tt.channelError != nil {
					return nil, tt.channelError
				}
				return tt.channelToReturn, nil
			}

			removeChannel := func(id string) {
				channelRemoved = id
			}

			handler := HandlePart(
				linkRegex,
				timestampFormat,
				setPendingUpdate,
				currentNick,
				getChannelByName,
				removeChannel,
			)

			handler(tt.message)

			assert.Equal(t, tt.expectUpdate, pendingUpdateCalled, "setPendingUpdate should be called")

			if tt.expectChannelRemoval {
				assert.Equal(t, tt.channelToReturn.GetID(), channelRemoved, "Channel should be removed")
			} else {
				assert.Equal(t, "", channelRemoved, "Channel should not be removed")
			}

			if tt.channelToReturn != nil && tt.channelError == nil {
				users := tt.channelToReturn.GetUsers()
				assert.Len(t, users, tt.expectedUserCount, "User count should match expected")

				// Check that the right user was removed (only if channel wasn't removed)
				if !tt.expectChannelRemoval {
					for _, user := range users {
						assert.NotEqual(t, tt.message.Nick(), user.GetNickListDisplay(), "Parted user should be removed from user list")
					}
				}

				if tt.expectMessage {
					messages := tt.channelToReturn.GetMessages()
					assert.Len(t, messages, 1, "Should add part message to channel")
					assert.Contains(t, messages[0].GetMessage(), "has parted", "Message should indicate part")
					assert.Contains(t, messages[0].GetMessage(), tt.message.Source, "Message should contain source")
					assert.Equal(t, MessageType(Event), messages[0].GetType(), "Should be event message type")
				} else {
					messages := tt.channelToReturn.GetMessages()
					assert.Len(t, messages, 0, "Should not add message when current user parts")
				}
			}
		})
	}
}
