package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestHandleUserModeSet(t *testing.T) {
	linkRegex := regexp.MustCompile(`https?://[^\s]+`)
	timestampFormat := "15:04:05"

	tests := []struct {
		name            string
		message         ircmsg.Message
		expectedModes   string
		expectedMessage string
	}{
		{
			name: "Set basic user modes",
			message: ircmsg.Message{
				Command: "221",
				Params:  []string{"testnick", "+iwx"},
			},
			expectedModes:   "+iwx",
			expectedMessage: "Your modes changed: +iwx",
		},
		{
			name: "Set modes without prefix",
			message: ircmsg.Message{
				Command: "221",
				Params:  []string{"user", "i"},
			},
			expectedModes:   "i",
			expectedMessage: "Your modes changed: i",
		},
		{
			name: "Set empty modes",
			message: ircmsg.Message{
				Command: "221",
				Params:  []string{"nick", ""},
			},
			expectedModes:   "",
			expectedMessage: "Your modes changed: ",
		},
		{
			name: "Set complex mode string",
			message: ircmsg.Message{
				Command: "221",
				Params:  []string{"someone", "+iws-o"},
			},
			expectedModes:   "+iws-o",
			expectedMessage: "Your modes changed: +iws-o",
		},
		{
			name: "Set modes with special characters",
			message: ircmsg.Message{
				Command: "221",
				Params:  []string{"nick", "+Zr"},
			},
			expectedModes:   "+Zr",
			expectedMessage: "Your modes changed: +Zr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pendingUpdateCalled := false
			finalModes := ""
			messagesAdded := []*Message{}

			setPendingUpdate := func() {
				pendingUpdateCalled = true
			}

			setCurrentModes := func(modes string) {
				finalModes = modes
			}

			addMessage := func(msg *Message) {
				messagesAdded = append(messagesAdded, msg)
			}

			handler := HandleUserModeSet(
				linkRegex,
				timestampFormat,
				setPendingUpdate,
				setCurrentModes,
				addMessage,
			)

			handler(tt.message)

			assert.True(t, pendingUpdateCalled, "setPendingUpdate should be called")
			assert.Len(t, messagesAdded, 1, "Should add exactly one message")
			assert.Equal(t, tt.expectedMessage, messagesAdded[0].GetMessage(), "Message text should match")
			assert.Equal(t, MessageType(Event), messagesAdded[0].GetType(), "Should be event message type")
			assert.Equal(t, tt.expectedModes, finalModes, "Final modes should match expected")
		})
	}
}