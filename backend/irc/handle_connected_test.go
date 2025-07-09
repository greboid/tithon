package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"testing"
)

// createTestServer creates a mock server for testing
func createTestServer() *Server {
	cm := &CommandManager{commands: []Command{}}
	return &Server{
		cm: cm,
		Window: &Window{
			id:           "test",
			name:         "test",
			title:        "test",
			messages:     make([]*Message, 0),
			tabCompleter: &NoopTabCompleter{},
		},
	}
}

// createTestChannel creates a test channel with a mock server
func createTestChannel(name string) *Channel {
	server := createTestServer()
	return NewChannel(server, name)
}

// createTestQuery creates a test query with a mock server
func createTestQuery(name string) *Query {
	server := createTestServer()
	return NewQuery(server, name)
}

func TestHandleConnected(t *testing.T) {
	timestampFormat := "15:04:05"

	tests := []struct {
		name                string
		message             ircmsg.Message
		networkSupport      string
		serverHostname      string
		channels            []*Channel
		queries             []*Query
		expectedMessage     string
		expectedServerName  string
		expectPendingUpdate bool
	}{
		{
			name: "Basic connection with network name",
			message: ircmsg.Message{
				Command: "001",
				Params:  []string{"testnick", "Welcome to the Example Network"},
			},
			networkSupport:      "ExampleNet",
			serverHostname:      "irc.example.com",
			channels:            []*Channel{},
			queries:             []*Query{},
			expectedMessage:     "Connected to irc.example.com",
			expectedServerName:  "ExampleNet",
			expectPendingUpdate: true,
		},
		{
			name: "Connection without network name",
			message: ircmsg.Message{
				Command: "001",
				Params:  []string{"testnick", "Welcome to the IRC Network"},
			},
			networkSupport:      "",
			serverHostname:      "irc.test.net",
			channels:            []*Channel{},
			queries:             []*Query{},
			expectedMessage:     "Connected to irc.test.net",
			expectedServerName:  "",
			expectPendingUpdate: true,
		},
		{
			name: "Connection with empty hostname",
			message: ircmsg.Message{
				Command: "001",
				Params:  []string{"user", "Welcome"},
			},
			networkSupport:      "TestNet",
			serverHostname:      "",
			channels:            []*Channel{},
			queries:             []*Query{},
			expectedMessage:     "Connected to ",
			expectedServerName:  "TestNet",
			expectPendingUpdate: true,
		},
		{
			name: "Connection with many channels and queries",
			message: ircmsg.Message{
				Command: "001",
				Params:  []string{"nick", "Welcome to BigNetwork"},
			},
			networkSupport: "BigNetwork",
			serverHostname: "irc.big.net",
			channels: []*Channel{
				createTestChannel("#general"),
				createTestChannel("#random"),
				createTestChannel("#dev"),
			},
			queries: []*Query{
				createTestQuery("friend1"),
				createTestQuery("friend2"),
			},
			expectedMessage:     "Connected to irc.big.net",
			expectedServerName:  "BigNetwork",
			expectPendingUpdate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track function calls
			pendingUpdateCalled := false
			messagesAdded := []*Message{}
			serverNameSet := ""

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

			iSupport := func(capability string) string {
				if capability == "NETWORK" {
					return tt.networkSupport
				}
				return ""
			}

			setServerName := func(name string) {
				serverNameSet = name
			}

			getChannels := func() []*Channel {
				return tt.channels
			}

			addMessage := func(msg *Message) {
				messagesAdded = append(messagesAdded, msg)
			}

			// Create handler
			handler := HandleConnected(
				timestampFormat,
				setPendingUpdate,
				getQueries,
				getServerHostname,
				iSupport,
				setServerName,
				getChannels,
				addMessage,
			)

			// Execute handler
			handler(tt.message)

			// Verify pending update was called
			assert.Equal(t, tt.expectPendingUpdate, pendingUpdateCalled, "setPendingUpdate should be called")

			// Verify server name was set correctly
			assert.Equal(t, tt.expectedServerName, serverNameSet, "Server name should be set correctly")

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

func TestHandleConnected_NetworkSupportVariations(t *testing.T) {
	timestampFormat := "15:04:05"

	tests := []struct {
		name               string
		supportResponse    map[string]string
		expectedServerName string
	}{
		{
			name: "Network support returns value",
			supportResponse: map[string]string{
				"NETWORK": "CustomNetwork",
			},
			expectedServerName: "CustomNetwork",
		},
		{
			name: "Network support returns empty string",
			supportResponse: map[string]string{
				"NETWORK": "",
			},
			expectedServerName: "",
		},
		{
			name: "Network support returns nothing (undefined capability)",
			supportResponse: map[string]string{
				"OTHER": "value",
			},
			expectedServerName: "",
		},
		{
			name: "Network support returns whitespace",
			supportResponse: map[string]string{
				"NETWORK": "   \t\n   ",
			},
			expectedServerName: "   \t\n   ",
		},
		{
			name: "Network support returns special characters",
			supportResponse: map[string]string{
				"NETWORK": "Net-Work_123",
			},
			expectedServerName: "Net-Work_123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverNameSet := ""

			// Mock functions
			setPendingUpdate := func() {}
			getQueries := func() []*Query { return []*Query{} }
			getServerHostname := func() string { return "test.server.com" }
			getChannels := func() []*Channel { return []*Channel{} }
			addMessage := func(msg *Message) {}

			iSupport := func(capability string) string {
				return tt.supportResponse[capability]
			}

			setServerName := func(name string) {
				serverNameSet = name
			}

			// Create handler
			handler := HandleConnected(
				timestampFormat,
				setPendingUpdate,
				getQueries,
				getServerHostname,
				iSupport,
				setServerName,
				getChannels,
				addMessage,
			)

			// Execute handler
			message := ircmsg.Message{
				Command: "001",
				Params:  []string{"testnick", "Welcome"},
			}
			handler(message)

			// Verify server name was set correctly
			assert.Equal(t, tt.expectedServerName, serverNameSet, "Server name should be set correctly")
		})
	}
}

func TestHandleConnected_MessageDistribution(t *testing.T) {
	timestampFormat := "15:04:05"

	// Create channels and queries
	channels := []*Channel{
		createTestChannel("#channel1"),
		createTestChannel("#channel2"),
	}
	queries := []*Query{
		createTestQuery("user1"),
		createTestQuery("user2"),
	}

	messagesAdded := []*Message{}

	// Mock functions
	setPendingUpdate := func() {}
	getQueries := func() []*Query { return queries }
	getServerHostname := func() string { return "test.example.com" }
	iSupport := func(capability string) string { return "TestNetwork" }
	setServerName := func(name string) {}
	getChannels := func() []*Channel { return channels }
	addMessage := func(msg *Message) {
		messagesAdded = append(messagesAdded, msg)
	}

	// Create handler
	handler := HandleConnected(
		timestampFormat,
		setPendingUpdate,
		getQueries,
		getServerHostname,
		iSupport,
		setServerName,
		getChannels,
		addMessage,
	)

	// Execute handler
	message := ircmsg.Message{
		Command: "001",
		Params:  []string{"testnick", "Welcome to TestNetwork"},
	}
	handler(message)

	// Verify main message
	assert.Len(t, messagesAdded, 1, "Should add exactly one message to main")
	assert.Equal(t, "Connected to test.example.com", messagesAdded[0].GetMessage())

	// Verify each channel received the message
	for i, channel := range channels {
		messages := channel.GetMessages()
		assert.Len(t, messages, 1, "Channel %d should have exactly one message", i)
		assert.Equal(t, "Connected to test.example.com", messages[0].GetMessage())
	}

	// Verify each query received the message
	for i, query := range queries {
		messages := query.GetMessages()
		assert.Len(t, messages, 1, "Query %d should have exactly one message", i)
		assert.Equal(t, "Connected to test.example.com", messages[0].GetMessage())
	}
}
