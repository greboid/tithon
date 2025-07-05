package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandleUserModes(t *testing.T) {
	timestampFormat := "15:04:05"

	tests := []struct {
		name             string
		message          ircmsg.Message
		currentModes     string
		isValidChannel   bool
		expectedModes    string
		expectedMessage  string
		expectUpdate     bool
		expectAddMessage bool
	}{
		{
			name: "Add single mode",
			message: ircmsg.Message{
				Command: "MODE",
				Params:  []string{"testnick", "+i"},
			},
			currentModes:     "",
			isValidChannel:   false,
			expectedModes:    "i",
			expectedMessage:  "Your modes changed: +i",
			expectUpdate:     true,
			expectAddMessage: true,
		},
		{
			name: "Remove single mode",
			message: ircmsg.Message{
				Command: "MODE",
				Params:  []string{"testnick", "-w"},
			},
			currentModes:     "iw",
			isValidChannel:   false,
			expectedModes:    "i",
			expectedMessage:  "Your modes changed: -w",
			expectUpdate:     true,
			expectAddMessage: true,
		},
		{
			name: "Add multiple modes",
			message: ircmsg.Message{
				Command: "MODE",
				Params:  []string{"user", "+iwx"},
			},
			currentModes:     "",
			isValidChannel:   false,
			expectedModes:    "iwx",
			expectedMessage:  "Your modes changed: +iwx",
			expectUpdate:     true,
			expectAddMessage: true,
		},
		{
			name: "Mixed add and remove modes",
			message: ircmsg.Message{
				Command: "MODE",
				Params:  []string{"nick", "+i-w+x"},
			},
			currentModes:     "w",
			isValidChannel:   false,
			expectedModes:    "ix",
			expectedMessage:  "Your modes changed: +i-w+x",
			expectUpdate:     true,
			expectAddMessage: true,
		},
		{
			name: "Mode without +/- prefix (defaults to +)",
			message: ircmsg.Message{
				Command: "MODE",
				Params:  []string{"nick", "i"},
			},
			currentModes:     "",
			isValidChannel:   false,
			expectedModes:    "i",
			expectedMessage:  "Your modes changed: +i",
			expectUpdate:     true,
			expectAddMessage: true,
		},
		{
			name: "Add duplicate mode (should not duplicate)",
			message: ircmsg.Message{
				Command: "MODE",
				Params:  []string{"nick", "+i"},
			},
			currentModes:     "iw",
			isValidChannel:   false,
			expectedModes:    "iw",
			expectedMessage:  "Your modes changed: +i",
			expectUpdate:     true,
			expectAddMessage: true,
		},
		{
			name: "Remove non-existent mode",
			message: ircmsg.Message{
				Command: "MODE",
				Params:  []string{"nick", "-z"},
			},
			currentModes:     "iw",
			isValidChannel:   false,
			expectedModes:    "iw",
			expectedMessage:  "Your modes changed: -z",
			expectUpdate:     true,
			expectAddMessage: true,
		},
		{
			name: "Channel target (should be ignored)",
			message: ircmsg.Message{
				Command: "MODE",
				Params:  []string{"#channel", "+o", "someone"},
			},
			currentModes:     "i",
			isValidChannel:   true,
			expectedModes:    "i",
			expectedMessage:  "",
			expectUpdate:     false,
			expectAddMessage: false,
		},
		{
			name: "Invalid message (insufficient params)",
			message: ircmsg.Message{
				Command: "MODE",
				Params:  []string{"nick"},
			},
			currentModes:     "i",
			isValidChannel:   false,
			expectedModes:    "i",
			expectedMessage:  "",
			expectUpdate:     true,
			expectAddMessage: false,
		},
		{
			name: "Empty mode string",
			message: ircmsg.Message{
				Command: "MODE",
				Params:  []string{"nick", ""},
			},
			currentModes:     "iw",
			isValidChannel:   false,
			expectedModes:    "iw",
			expectedMessage:  "Your modes changed: +",
			expectUpdate:     true,
			expectAddMessage: true,
		},
		{
			name: "Complex mode changes",
			message: ircmsg.Message{
				Command: "MODE",
				Params:  []string{"user", "+abc-def+ghi"},
			},
			currentModes:     "defxyz",
			isValidChannel:   false,
			expectedModes:    "xyzabcghi",
			expectedMessage:  "Your modes changed: +abc-def+ghi",
			expectUpdate:     true,
			expectAddMessage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pendingUpdateCalled := false
			finalModes := ""
			messagesAdded := []*Message{}

			isValidChannel := func(target string) bool {
				return tt.isValidChannel
			}

			setPendingUpdate := func() {
				pendingUpdateCalled = true
			}

			getCurrentModes := func() string {
				return tt.currentModes
			}

			setCurrentModes := func(modes string) {
				finalModes = modes
			}

			addMessage := func(msg *Message) {
				messagesAdded = append(messagesAdded, msg)
			}

			handler := HandleUserModes(
				timestampFormat,
				isValidChannel,
				setPendingUpdate,
				getCurrentModes,
				setCurrentModes,
				addMessage,
			)

			handler(tt.message)

			assert.Equal(t, tt.expectUpdate, pendingUpdateCalled, "setPendingUpdate call should match expectation")

			if tt.expectAddMessage {
				assert.Len(t, messagesAdded, 1, "Should add exactly one message")
				assert.Equal(t, tt.expectedMessage, messagesAdded[0].GetMessage(), "Message text should match")
				assert.Equal(t, MessageType(Event), messagesAdded[0].GetType(), "Should be event message type")
				assert.Equal(t, tt.expectedModes, finalModes, "Final modes should match expected")
			} else {
				assert.Len(t, messagesAdded, 0, "Should not add any messages")
				if tt.isValidChannel {
					assert.Equal(t, tt.currentModes, tt.expectedModes, "Modes should not change for channel targets")
				}
			}
		})
	}
}
