package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHandleDisconnected(t *testing.T) {
	timestampFormat := "15:04:05"

	tests := []struct {
		name                string
		message             ircmsg.Message
		serverHostname      string
		channels            []*Channel
		queries             []*Query
		expectedMessage     string
		expectPendingUpdate bool
	}{
		{
			name: "Basic disconnect message",
			message: ircmsg.Message{
				Command: "ERROR",
				Params:  []string{"Connection reset"},
			},
			serverHostname:      "irc.example.com",
			channels:            []*Channel{},
			queries:             []*Query{},
			expectedMessage:     "Disconnected from irc.example.com: Connection reset",
			expectPendingUpdate: true,
		},
		{
			name: "Disconnect with multiple params",
			message: ircmsg.Message{
				Command: "ERROR",
				Params:  []string{"Closing", "Link:", "Connection", "timeout"},
			},
			serverHostname:      "irc.network.org",
			channels:            []*Channel{},
			queries:             []*Query{},
			expectedMessage:     "Disconnected from irc.network.org: Closing Link: Connection timeout",
			expectPendingUpdate: true,
		},
		{
			name: "Disconnect with channels and queries",
			message: ircmsg.Message{
				Command: "ERROR",
				Params:  []string{"Server shutdown"},
			},
			serverHostname: "irc.test.net",
			channels: []*Channel{
				createTestChannel("#test"),
				createTestChannel("#general"),
			},
			queries: []*Query{
				createTestQuery("friend"),
			},
			expectedMessage:     "Disconnected from irc.test.net: Server shutdown",
			expectPendingUpdate: true,
		},
		{
			name: "Disconnect with no params",
			message: ircmsg.Message{
				Command: "ERROR",
				Params:  []string{},
			},
			serverHostname:      "irc.empty.com",
			channels:            []*Channel{},
			queries:             []*Query{},
			expectedMessage:     "Disconnected from irc.empty.com: ",
			expectPendingUpdate: true,
		},
		{
			name: "Disconnect with single empty param",
			message: ircmsg.Message{
				Command: "ERROR",
				Params:  []string{""},
			},
			serverHostname:      "irc.blank.net",
			channels:            []*Channel{},
			queries:             []*Query{},
			expectedMessage:     "Disconnected from irc.blank.net: ",
			expectPendingUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track function calls
			pendingUpdateCalled := false
			messagesAdded := []*Message{}

			// Mock functions
			setPendingUpdate := func() {
				pendingUpdateCalled = true
			}

			getQueries := func() []*Query {
				return tt.queries
			}

			getServerHostname := func() string {
				return tt.serverHostname
			}

			getChannels := func() []*Channel {
				return tt.channels
			}

			addMessage := func(msg *Message) {
				messagesAdded = append(messagesAdded, msg)
			}

			// Create handler
			handler := HandleDisconnected(
				timestampFormat,
				setPendingUpdate,
				getQueries,
				getServerHostname,
				getChannels,
				addMessage,
			)

			// Execute handler
			handler(tt.message)

			// Verify pending update was called
			assert.Equal(t, tt.expectPendingUpdate, pendingUpdateCalled, "setPendingUpdate should be called")

			// Verify main message was added
			assert.Len(t, messagesAdded, 1, "Should add exactly one message to main")
			assert.Equal(t, tt.expectedMessage, messagesAdded[0].GetMessage(), "Main message text should match")
			assert.Equal(t, MessageType(Event), messagesAdded[0].GetType(), "Should be event message type")

			// Verify messages were added to all channels
			for _, channel := range tt.channels {
				messages := channel.GetMessages()
				assert.Len(t, messages, 1, "Should add message to channel %s", channel.GetName())
				assert.Equal(t, tt.expectedMessage, messages[0].GetMessage(), "Channel message should match")
				assert.Equal(t, MessageType(Event), messages[0].GetType(), "Should be event message type")
			}

			// Verify messages were added to all queries
			for _, query := range tt.queries {
				messages := query.GetMessages()
				assert.Len(t, messages, 1, "Should add message to query %s", query.GetName())
				assert.Equal(t, tt.expectedMessage, messages[0].GetMessage(), "Query message should match")
				assert.Equal(t, MessageType(Event), messages[0].GetType(), "Should be event message type")
			}
		})
	}
}
