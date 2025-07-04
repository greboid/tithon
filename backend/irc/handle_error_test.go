package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestHandleError(t *testing.T) {
	linkRegex := regexp.MustCompile(`https?://[^\s]+`)
	timestampFormat := "15:04:05"

	tests := []struct {
		name            string
		message         ircmsg.Message
		expectedMessage string
	}{
		{
			name: "Single error parameter",
			message: ircmsg.Message{
				Command: "ERROR",
				Params:  []string{"Connection failed"},
			},
			expectedMessage: "Connection failed",
		},
		{
			name: "Multiple error parameters",
			message: ircmsg.Message{
				Command: "ERROR",
				Params:  []string{"Closing", "Link:", "Banned"},
			},
			expectedMessage: "Closing Link: Banned",
		},
		{
			name: "Empty error parameters",
			message: ircmsg.Message{
				Command: "ERROR",
				Params:  []string{},
			},
			expectedMessage: "",
		},
		{
			name: "Single empty parameter",
			message: ircmsg.Message{
				Command: "ERROR",
				Params:  []string{""},
			},
			expectedMessage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pendingUpdateCalled := false
			messagesAdded := []*Message{}

			setPendingUpdate := func() {
				pendingUpdateCalled = true
			}

			addMessage := func(msg *Message) {
				messagesAdded = append(messagesAdded, msg)
			}

			handler := HandleError(linkRegex, timestampFormat, setPendingUpdate, addMessage)
			handler(tt.message)

			assert.True(t, pendingUpdateCalled, "setPendingUpdate should be called")
			assert.Len(t, messagesAdded, 1, "Should add exactly one error message")
			assert.Equal(t, tt.expectedMessage, messagesAdded[0].GetMessage(), "Error message text should match")
			assert.Equal(t, MessageType(Error), messagesAdded[0].GetType(), "Should be error message type")
		})
	}
}

func TestHandleNickInUse(t *testing.T) {
	linkRegex := regexp.MustCompile(`https?://[^\s]+`)
	timestampFormat := "15:04:05"

	tests := []struct {
		name            string
		message         ircmsg.Message
		expectedMessage string
	}{
		{
			name: "Basic nick in use",
			message: ircmsg.Message{
				Command: "433",
				Params:  []string{"*", "testnick", "Nickname is already in use"},
			},
			expectedMessage: "Nickname in use: testnick",
		},
		{
			name: "Nick in use with different params",
			message: ircmsg.Message{
				Command: "433",
				Params:  []string{"currentnick", "wantednick", "Nick already in use"},
			},
			expectedMessage: "Nickname in use: wantednick",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pendingUpdateCalled := false
			messagesAdded := []*Message{}

			setPendingUpdate := func() {
				pendingUpdateCalled = true
			}

			addMessage := func(msg *Message) {
				messagesAdded = append(messagesAdded, msg)
			}

			handler := HandleNickInUse(linkRegex, timestampFormat, setPendingUpdate, addMessage)
			handler(tt.message)

			assert.True(t, pendingUpdateCalled, "setPendingUpdate should be called")
			assert.Len(t, messagesAdded, 1, "Should add exactly one error message")
			assert.Equal(t, tt.expectedMessage, messagesAdded[0].GetMessage(), "Error message text should match")
			assert.Equal(t, MessageType(Error), messagesAdded[0].GetType(), "Should be error message type")
		})
	}
}

func TestHandlePasswordMismatch(t *testing.T) {
	linkRegex := regexp.MustCompile(`https?://[^\s]+`)
	timestampFormat := "15:04:05"

	tests := []struct {
		name            string
		message         ircmsg.Message
		expectedMessage string
	}{
		{
			name: "Basic password mismatch",
			message: ircmsg.Message{
				Command: "464",
				Params:  []string{"Password incorrect"},
			},
			expectedMessage: "Password Mismatch: Password incorrect",
		},
		{
			name: "Password mismatch with multiple params",
			message: ircmsg.Message{
				Command: "464",
				Params:  []string{"Access", "denied:", "Bad", "password"},
			},
			expectedMessage: "Password Mismatch: Access denied: Bad password",
		},
		{
			name: "Empty password mismatch",
			message: ircmsg.Message{
				Command: "464",
				Params:  []string{},
			},
			expectedMessage: "Password Mismatch: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pendingUpdateCalled := false
			messagesAdded := []*Message{}

			setPendingUpdate := func() {
				pendingUpdateCalled = true
			}

			addMessage := func(msg *Message) {
				messagesAdded = append(messagesAdded, msg)
			}

			handler := HandlePasswordMismatch(linkRegex, timestampFormat, setPendingUpdate, addMessage)
			handler(tt.message)

			assert.True(t, pendingUpdateCalled, "setPendingUpdate should be called")
			assert.Len(t, messagesAdded, 1, "Should add exactly one error message")
			assert.Equal(t, tt.expectedMessage, messagesAdded[0].GetMessage(), "Error message text should match")
			assert.Equal(t, MessageType(Error), messagesAdded[0].GetType(), "Should be error message type")
		})
	}
}