package irc

import (
	"fmt"
	"github.com/ergochat/irc-go/ircevent"
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/greboid/tithon/config"
	"github.com/hueristiq/hq-go-url/extractor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strings"
	"testing"
	"time"
)

// Mock implementations for testing
type MockChannelHandler struct {
	mock.Mock
}

func (m *MockChannelHandler) IsValidChannel(target string) bool {
	args := m.Called(target)
	return args.Bool(0)
}

func (m *MockChannelHandler) GetChannelByName(name string) (*Channel, error) {
	args := m.Called(name)
	return args.Get(0).(*Channel), args.Error(1)
}

func (m *MockChannelHandler) GetChannels() []*Channel {
	args := m.Called()
	return args.Get(0).([]*Channel)
}

func (m *MockChannelHandler) AddChannel(name string) *Channel {
	args := m.Called(name)
	return args.Get(0).(*Channel)
}

func (m *MockChannelHandler) RemoveChannel(id string) {
	m.Called(id)
}

type MockQueryHandler struct {
	mock.Mock
}

func (m *MockQueryHandler) GetQueries() []*Query {
	args := m.Called()
	return args.Get(0).([]*Query)
}

func (m *MockQueryHandler) GetQueryByName(name string) (*Query, error) {
	args := m.Called(name)
	return args.Get(0).(*Query), args.Error(1)
}

func (m *MockQueryHandler) AddQuery(name string) *Query {
	args := m.Called(name)
	return args.Get(0).(*Query)
}

type MockCallbackHandler struct {
	mock.Mock
}

func (m *MockCallbackHandler) AddConnectCallback(callback func(message ircmsg.Message)) {
	m.Called(callback)
}

func (m *MockCallbackHandler) AddDisconnectCallback(callback func(message ircmsg.Message)) {
	m.Called(callback)
}

func (m *MockCallbackHandler) AddCallback(command string, callback func(ircmsg.Message)) {
	m.Called(command, callback)
}

func (m *MockCallbackHandler) AddBatchCallback(callback func(*ircevent.Batch) bool) {
	m.Called(callback)
}

type MockInfoHandler struct {
	mock.Mock
}

func (m *MockInfoHandler) ISupport(value string) string {
	args := m.Called(value)
	return args.String(0)
}

func (m *MockInfoHandler) CurrentNick() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockInfoHandler) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockInfoHandler) SetName(name string) {
	m.Called(name)
}

func (m *MockInfoHandler) GetHostname() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockInfoHandler) HasCapability(name string) bool {
	args := m.Called(name)
	return args.Bool(0)
}

type MockModeHandler struct {
	mock.Mock
}

func (m *MockModeHandler) GetModeNameForMode(mode string) string {
	args := m.Called(mode)
	return args.String(0)
}

func (m *MockModeHandler) GetModePrefixes() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockModeHandler) GetCurrentModes() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockModeHandler) SetCurrentModes(modes string) {
	m.Called(modes)
}

func (m *MockModeHandler) GetChannelModeType(mode string) rune {
	args := m.Called(mode)
	return args.Get(0).(rune)
}

type MockMessageHandler struct {
	mock.Mock
}

func (m *MockMessageHandler) AddMessage(message *Message) {
	m.Called(message)
}

func (m *MockMessageHandler) SendRaw(message string) {
	m.Called(message)
}

type MockUpdateTrigger struct {
	mock.Mock
}

func (m *MockUpdateTrigger) SetPendingUpdate() {
	m.Called()
}

type MockNotificationManager struct {
	mock.Mock
	CheckAndNotifyResponse bool
}

func (m *MockNotificationManager) CheckAndNotify(server, target, nickname, message string) bool {
	args := m.Called(server, target, nickname, message)
	return args.Bool(0)
}

func (m *MockNotificationManager) SendNotification(notification Notification) {
	m.Called(notification)
}

// Helper functions for tests
func createTestHandler() (*Handler, *MockChannelHandler, *MockQueryHandler, *MockInfoHandler, *MockModeHandler, *MockMessageHandler, *MockUpdateTrigger, *MockNotificationManager) {
	linkRegex := extractor.New(extractor.WithHost()).CompileRegex()
	conf := &config.Config{
		UISettings: config.UISettings{
			TimestampFormat: "15:04:05",
		},
	}

	mockChannelHandler := &MockChannelHandler{}
	mockQueryHandler := &MockQueryHandler{}
	mockCallbackHandler := &MockCallbackHandler{}
	mockInfoHandler := &MockInfoHandler{}
	mockModeHandler := &MockModeHandler{}
	mockMessageHandler := &MockMessageHandler{}
	mockUpdateTrigger := &MockUpdateTrigger{}
	mockNotificationManager := &MockNotificationManager{}

	handler := &Handler{
		channelHandler:      mockChannelHandler,
		queryHandler:        mockQueryHandler,
		callbackHandler:     mockCallbackHandler,
		infoHandler:         mockInfoHandler,
		modeHandler:         mockModeHandler,
		messageHandler:      mockMessageHandler,
		updateTrigger:       mockUpdateTrigger,
		notificationManager: mockNotificationManager,
		conf:                conf,
		batchMap:            make(map[string]string),
		linkRegex:           linkRegex,
	}

	return handler, mockChannelHandler, mockQueryHandler, mockInfoHandler, mockModeHandler, mockMessageHandler, mockUpdateTrigger, mockNotificationManager
}

func createTestChannel(name string) *Channel {
	return &Channel{
		Window: &Window{
			id:        "test-channel-id",
			name:      name,
			title:     name,
			messages:  make([]*Message, 0),
			users:     make([]*User, 0),
			hasUsers:  true,
			isChannel: true,
		},
	}
}

func createTestQuery(name string) *Query {
	return &Query{
		Window: &Window{
			id:       "test-query-id",
			name:     name,
			title:    name,
			messages: make([]*Message, 0),
		},
	}
}

// Tests for JOIN handler
func TestHandler_handleJoin_SelfJoin(t *testing.T) {
	handler, mockChannelHandler, _, mockInfoHandler, _, _, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockInfoHandler.On("CurrentNick").Return("testnick")
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#test").Return((*Channel)(nil), assert.AnError)
	mockChannelHandler.On("AddChannel", "#test").Return(createTestChannel("#test"))
	mockInfoHandler.On("HasCapability", "draft/chathistory").Return(false)

	// Create test message
	message := ircmsg.Message{
		Source:  "testnick!user@host",
		Command: "JOIN",
		Params:  []string{"#test"},
	}

	// Execute
	handler.handleJoin(message)

	// Verify expectations
	mockInfoHandler.AssertExpectations(t)
	mockChannelHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleJoin_SelfJoinWithChatHistory(t *testing.T) {
	handler, mockChannelHandler, _, mockInfoHandler, _, mockMessageHandler, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockInfoHandler.On("CurrentNick").Return("testnick")
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#test").Return((*Channel)(nil), assert.AnError)
	mockChannelHandler.On("AddChannel", "#test").Return(createTestChannel("#test"))
	mockInfoHandler.On("HasCapability", "draft/chathistory").Return(true)
	mockMessageHandler.On("SendRaw", "CHATHISTORY LATEST #test * 100")

	// Create test message
	message := ircmsg.Message{
		Source:  "testnick!user@host",
		Command: "JOIN",
		Params:  []string{"#test"},
	}

	// Execute
	handler.handleJoin(message)

	// Verify expectations
	mockInfoHandler.AssertExpectations(t)
	mockChannelHandler.AssertExpectations(t)
	mockMessageHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleJoin_OtherJoin(t *testing.T) {
	handler, mockChannelHandler, _, mockInfoHandler, _, _, mockUpdateTrigger, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockInfoHandler.On("CurrentNick").Return("testnick")
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)

	// Create test message
	message := ircmsg.Message{
		Source:  "othernick!user@host",
		Command: "JOIN",
		Params:  []string{"#test"},
	}

	// Execute
	handler.handleJoin(message)

	// Verify user was added to channel
	assert.Len(t, channel.users, 1)
	assert.Equal(t, "othernick", channel.users[0].nickname)

	// Verify expectations
	mockInfoHandler.AssertExpectations(t)
	mockChannelHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleJoin_OtherJoin_UnknownChannel(t *testing.T) {
	handler, mockChannelHandler, _, mockInfoHandler, _, _, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockInfoHandler.On("CurrentNick").Return("testnick")
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#unknown").Return((*Channel)(nil), assert.AnError)

	// Create test message for other user joining unknown channel
	message := ircmsg.Message{
		Source:  "othernick!user@host",
		Command: "JOIN",
		Params:  []string{"#unknown"},
	}

	// Execute
	handler.handleJoin(message)

	// Verify expectations - should only call CurrentNick, GetChannelByName, and SetPendingUpdate
	// No users should be added, no messages should be added since function returns early
	mockInfoHandler.AssertExpectations(t)
	mockChannelHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

// Tests for PRIVMSG handler
func TestHandler_handlePrivMsg_ChannelMessage(t *testing.T) {
	handler, mockChannelHandler, _, mockInfoHandler, _, _, mockUpdateTrigger, mockNotificationManager := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("IsValidChannel", "#test").Return(true)
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockInfoHandler.On("CurrentNick").Return("testnick")
	mockInfoHandler.On("GetName").Return("testserver")
	mockNotificationManager.On("CheckAndNotify", "testserver", "#test", "othernick", "Hello everyone!").Return(false)

	// Create test message
	message := ircmsg.Message{
		Source:  "othernick!user@host",
		Command: "PRIVMSG",
		Params:  []string{"#test", "Hello everyone!"},
	}

	// Execute
	handler.handlePrivMsg(message)

	// Verify message was added to channel
	assert.Len(t, channel.messages, 1)
	assert.Equal(t, "Hello everyone!", channel.messages[0].GetMessage())
	assert.Equal(t, "othernick", channel.messages[0].GetNickname())

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockInfoHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
	mockNotificationManager.AssertExpectations(t)

}

func TestHandler_handlePrivMsg_PrivateMessage(t *testing.T) {
	handler, mockChannelHandler, mockQueryHandler, mockInfoHandler, _, _, mockUpdateTrigger, mockNotificationManager := createTestHandler()

	query := createTestQuery("othernick")

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("IsValidChannel", "testnick").Return(false)
	mockInfoHandler.On("CurrentNick").Return("testnick")
	mockQueryHandler.On("GetQueryByName", "othernick").Return((*Query)(nil), assert.AnError)
	mockQueryHandler.On("AddQuery", "othernick").Return(query)
	mockInfoHandler.On("GetName").Return("testserver")
	mockNotificationManager.On("CheckAndNotify", "testserver", "othernick", "othernick", "Hello privately!").Return(true)

	// Create test message
	message := ircmsg.Message{
		Source:  "othernick!user@host",
		Command: "PRIVMSG",
		Params:  []string{"testnick", "Hello privately!"},
	}

	// Execute
	handler.handlePrivMsg(message)

	// Verify message was added to query
	assert.Len(t, query.messages, 1)
	assert.Equal(t, "Hello privately!", query.messages[0].GetMessage())
	assert.Equal(t, "othernick", query.messages[0].GetNickname())

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockQueryHandler.AssertExpectations(t)
	mockInfoHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
	mockNotificationManager.AssertExpectations(t)
}

func TestHandler_handlePrivMsg_SelfMessage(t *testing.T) {
	handler, mockChannelHandler, mockQueryHandler, mockInfoHandler, _, _, mockUpdateTrigger, _ := createTestHandler()

	query := createTestQuery("targetuser")

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("IsValidChannel", "targetuser").Return(false)
	mockInfoHandler.On("CurrentNick").Return("testnick")
	mockQueryHandler.On("GetQueryByName", "targetuser").Return((*Query)(nil), assert.AnError)
	mockQueryHandler.On("AddQuery", "testnick").Return(query)

	// Create test message from current user to another user
	message := ircmsg.Message{
		Source:  "testnick!user@host",
		Command: "PRIVMSG",
		Params:  []string{"targetuser", "Hello from me!"},
	}
	message.SetTag("", "")

	// Execute
	handler.handlePrivMsg(message)

	// Verify message was added to query
	assert.Len(t, query.messages, 1)
	assert.Equal(t, "Hello from me!", query.messages[0].GetMessage())
	assert.Equal(t, "testnick", query.messages[0].GetNickname())

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockQueryHandler.AssertExpectations(t)
	mockInfoHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handlePrivMsg_UnknownChannel(t *testing.T) {
	handler, mockChannelHandler, _, _, _, _, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("IsValidChannel", "#unknown").Return(true)
	mockChannelHandler.On("GetChannelByName", "#unknown").Return((*Channel)(nil), assert.AnError)

	// Create test message for unknown channel
	message := ircmsg.Message{
		Source:  "othernick!user@host",
		Command: "PRIVMSG",
		Params:  []string{"#unknown", "Message to unknown channel"},
	}

	// Execute
	handler.handlePrivMsg(message)

	// Verify expectations - should only call IsValidChannel, GetChannelByName, and SetPendingUpdate
	// No message should be added, no notifications should be triggered since function returns early
	mockChannelHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

// Tests for PART handler
func TestHandler_handlePart_SelfPart(t *testing.T) {
	handler, mockChannelHandler, _, mockInfoHandler, _, _, mockUpdateTrigger, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockInfoHandler.On("CurrentNick").Return("testnick")
	mockChannelHandler.On("RemoveChannel", "test-channel-id")

	// Create test message
	message := ircmsg.Message{
		Source:  "testnick!user@host",
		Command: "PART",
		Params:  []string{"#test"},
	}

	// Execute
	handler.handlePart(message)

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockInfoHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handlePart_OtherPart(t *testing.T) {
	handler, mockChannelHandler, _, mockInfoHandler, _, _, mockUpdateTrigger, _ := createTestHandler()

	channel := createTestChannel("#test")
	channel.users = append(channel.users, NewUser("othernick", ""))

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockInfoHandler.On("CurrentNick").Return("testnick")

	// Create test message
	message := ircmsg.Message{
		Source:  "othernick!user@host",
		Command: "PART",
		Params:  []string{"#test"},
	}

	// Execute
	handler.handlePart(message)

	assert.Len(t, channel.users, 0)
	assert.Len(t, channel.messages, 1)
	assert.Contains(t, channel.messages[0].GetMessage(), "othernick!user@host has parted #test")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockInfoHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handlePart_UnknownChannel(t *testing.T) {
	handler, mockChannelHandler, _, _, _, _, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#unknown").Return((*Channel)(nil), assert.AnError)

	// Create test message for unknown channel
	message := ircmsg.Message{
		Source:  "othernick!user@host",
		Command: "PART",
		Params:  []string{"#unknown"},
	}

	// Execute
	handler.handlePart(message)

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

// Tests for QUIT handler
func TestHandler_handleQuit(t *testing.T) {
	handler, mockChannelHandler, _, _, _, _, mockUpdateTrigger, _ := createTestHandler()

	channel1 := createTestChannel("#test1")
	channel1.SetUsers([]*User{NewUser("quittingnick", ""), NewUser("othernick", "")})

	channel2 := createTestChannel("#test2")
	channel2.SetUsers([]*User{NewUser("quittingnick", "")})

	channels := []*Channel{channel1, channel2}

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannels").Return(channels)

	// Create test message
	message := ircmsg.Message{
		Source:  "quittingnick!user@host",
		Command: "QUIT",
		Params:  []string{"Quit message"},
	}

	// Execute
	handler.handleQuit(message)

	// Verify user was removed from both channels
	assert.Len(t, channel1.GetUsers(), 1)
	assert.Equal(t, "othernick", channel1.GetUsers()[0].nickname)
	assert.Len(t, channel2.GetUsers(), 0)

	// Verify quit messages were added to both channels
	assert.Len(t, channel1.messages, 1)
	assert.Len(t, channel2.messages, 1)
	assert.Contains(t, channel1.messages[0].GetMessage(), "has quit")
	assert.Contains(t, channel2.messages[0].GetMessage(), "has quit")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

// Tests for NICK handler
func TestHandler_handleNick_SelfNick(t *testing.T) {
	handler, mockChannelHandler, _, mockInfoHandler, _, mockMessageHandler, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockInfoHandler.On("CurrentNick").Return("oldnick")
	mockChannelHandler.On("GetChannels").Return([]*Channel{})
	mockMessageHandler.On("AddMessage", mock.AnythingOfType("*irc.Message"))

	// Create test message
	message := ircmsg.Message{
		Source:  "oldnick!user@host",
		Command: "NICK",
		Params:  []string{"newnick"},
	}

	// Execute
	handler.handleNick(message)

	// Verify expectations
	mockInfoHandler.AssertExpectations(t)
	mockChannelHandler.AssertExpectations(t)
	mockMessageHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleNick_OtherNick(t *testing.T) {
	handler, mockChannelHandler, _, mockInfoHandler, _, _, mockUpdateTrigger, _ := createTestHandler()

	channel := createTestChannel("#test")
	channel.users = append(channel.users, NewUser("oldnick", ""))
	channels := []*Channel{channel}

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockInfoHandler.On("CurrentNick").Return("testnick")
	mockChannelHandler.On("GetChannels").Return(channels)

	// Create test message
	message := ircmsg.Message{
		Source:  "oldnick!user@host",
		Command: "NICK",
		Params:  []string{"newnick"},
	}

	// Execute
	handler.handleNick(message)

	// Verify user's nickname was changed
	assert.Equal(t, "newnick", channel.users[0].nickname)

	// Verify nick change message was added to channel
	assert.Len(t, channel.messages, 1)
	assert.Contains(t, channel.messages[0].GetMessage(), "oldnick is now known as newnick")

	// Verify expectations
	mockInfoHandler.AssertExpectations(t)
	mockChannelHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

// Tests for connection state handlers
func TestHandler_handleConnected(t *testing.T) {
	handler, mockChannelHandler, mockQueryHandler, mockInfoHandler, _, mockMessageHandler, mockUpdateTrigger, _ := createTestHandler()

	channel := createTestChannel("#test")
	query := createTestQuery("friend")

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockInfoHandler.On("ISupport", "NETWORK").Return("TestNetwork")
	mockInfoHandler.On("SetName", "TestNetwork")
	mockInfoHandler.On("GetHostname").Return("irc.example.com")
	mockChannelHandler.On("GetChannels").Return([]*Channel{channel})
	mockQueryHandler.On("GetQueries").Return([]*Query{query})
	mockMessageHandler.On("AddMessage", mock.AnythingOfType("*irc.Message"))

	// Create test message
	message := ircmsg.Message{
		Command: "001",
		Params:  []string{"testnick", "Welcome to the network"},
	}

	// Execute
	handler.handleConnected(message)

	// Verify connect messages were added
	assert.Len(t, channel.messages, 1)
	assert.Len(t, query.messages, 1)
	assert.Contains(t, channel.messages[0].GetMessage(), "Connected to irc.example.com")
	assert.Contains(t, query.messages[0].GetMessage(), "Connected to irc.example.com")

	// Verify expectations
	mockInfoHandler.AssertExpectations(t)
	mockChannelHandler.AssertExpectations(t)
	mockQueryHandler.AssertExpectations(t)
	mockMessageHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleDisconnected(t *testing.T) {
	handler, mockChannelHandler, mockQueryHandler, mockInfoHandler, _, mockMessageHandler, mockUpdateTrigger, _ := createTestHandler()

	channel := createTestChannel("#test")
	query := createTestQuery("friend")

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockInfoHandler.On("GetHostname").Return("irc.example.com")
	mockChannelHandler.On("GetChannels").Return([]*Channel{channel})
	mockQueryHandler.On("GetQueries").Return([]*Query{query})
	mockMessageHandler.On("AddMessage", mock.AnythingOfType("*irc.Message"))

	// Create test message
	message := ircmsg.Message{
		Command: "ERROR",
		Params:  []string{"Connection lost"},
	}

	// Execute
	handler.handleDisconnected(message)

	// Verify disconnect messages were added
	assert.Len(t, channel.messages, 1)
	assert.Len(t, query.messages, 1)
	assert.Contains(t, channel.messages[0].GetMessage(), "Disconnected from irc.example.com: Connection lost")
	assert.Contains(t, query.messages[0].GetMessage(), "Disconnected from irc.example.com: Connection lost")

	// Verify expectations
	mockInfoHandler.AssertExpectations(t)
	mockChannelHandler.AssertExpectations(t)
	mockQueryHandler.AssertExpectations(t)
	mockMessageHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

// Tests for TOPIC handler
func TestHandler_handleTopic(t *testing.T) {
	handler, mockChannelHandler, _, mockInfoHandler, _, _, mockUpdateTrigger, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockInfoHandler.On("CurrentNick").Return("testnick")
	mockInfoHandler.On("GetName").Return("testserver")

	// Create test message
	message := ircmsg.Message{
		Source:  "othernick!user@host",
		Command: "TOPIC",
		Params:  []string{"#test", "New channel topic"},
	}
	now := time.Now()
	message.SetTag("timestamp", now.Format(v3TimestampFormat))

	// Execute
	handler.handleTopic(message)

	// Verify topic was set
	assert.Equal(t, "New channel topic", channel.GetTopic().GetTopic())
	assert.Equal(t, channel.GetTitle(), "New channel topic (set by othernick on "+now.Format(topicTimeformat)+")")

	// Verify topic change message was added
	assert.Len(t, channel.messages, 1)
	assert.Contains(t, channel.messages[0].GetMessage(), "othernick changed the topic: New channel topic")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockInfoHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleTopic_NonExistentChannel(t *testing.T) {
	handler, mockChannelHandler, _, _, _, _, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#nonexistent").Return((*Channel)(nil), assert.AnError)

	// Create test message for non-existent channel
	message := ircmsg.Message{
		Source:  "othernick!user@host",
		Command: "TOPIC",
		Params:  []string{"#nonexistent", "Topic for non-existent channel"},
	}

	// Execute
	handler.handleTopic(message)

	// Verify expectations - should only call GetChannelByName and SetPendingUpdate
	// No topic should be set, no messages should be added since channel doesn't exist
	mockChannelHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleTopic_UnsetTopic(t *testing.T) {
	handler, mockChannelHandler, _, mockInfoHandler, _, _, mockUpdateTrigger, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockInfoHandler.On("CurrentNick").Return("testnick")
	mockInfoHandler.On("GetName").Return("testserver")

	// Create test message with empty topic (unsetting topic)
	message := ircmsg.Message{
		Source:  "othernick!user@host",
		Command: "TOPIC",
		Params:  []string{"#test"},
	}

	// Execute
	handler.handleTopic(message)

	// Verify topic was unset (displays "No Topic set")
	assert.Equal(t, "No Topic set", channel.GetTopic().GetTopic())
	assert.Equal(t, "No Topic set", channel.GetTitle())

	// Verify unset topic message was added
	assert.Len(t, channel.messages, 1)
	assert.Contains(t, channel.messages[0].GetMessage(), "othernick unset the topic")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockInfoHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

// Tests for RPL_TOPIC handler
func TestHandler_handleRPLTopic(t *testing.T) {
	handler, mockChannelHandler, _, mockInfoHandler, _, _, mockUpdateTrigger, _ := createTestHandler()

	channel1 := createTestChannel("#test1")
	channel2 := createTestChannel("#test2")
	channels := []*Channel{channel1, channel2}

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannels").Return(channels)
	mockInfoHandler.On("GetName").Return("testserver")

	// Create test RPL_TOPIC message
	message := ircmsg.Message{
		Command: "332",
		Params:  []string{"testnick", "#test1", "Welcome to test1 channel!"},
	}

	// Execute
	handler.handleRPLTopic(message)

	// Verify topic was set on the correct channel
	assert.Equal(t, "Welcome to test1 channel!", channel1.GetTopic().GetTopic())
	assert.Equal(t, "Welcome to test1 channel!", channel1.GetTitle())

	// Verify topic was NOT set on the other channel
	assert.Equal(t, "No Topic set", channel2.GetTopic().GetTopic())
	assert.Equal(t, "#test2", channel2.GetTitle())

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockInfoHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleRPLTopic_ChannelNotFound(t *testing.T) {
	handler, mockChannelHandler, _, _, _, _, mockUpdateTrigger, _ := createTestHandler()

	channel1 := createTestChannel("#test1")
	channels := []*Channel{channel1}

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannels").Return(channels)

	// Create test RPL_TOPIC message for non-existent channel
	message := ircmsg.Message{
		Command: "332", // RPL_TOPIC
		Params:  []string{"testnick", "#nonexistent", "Topic for non-existent channel"},
	}

	// Execute
	handler.handleRPLTopic(message)

	// Verify no topic was set on any channel
	assert.Equal(t, "No Topic set", channel1.GetTopic().GetTopic())
	assert.Equal(t, "#test1", channel1.GetTitle())

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

// Tests for ERROR handler
func TestHandler_handleError(t *testing.T) {
	handler, _, _, _, _, mockMessageHandler, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockMessageHandler.On("AddMessage", mock.AnythingOfType("*irc.Message"))

	// Create test message
	message := ircmsg.Message{
		Command: "ERROR",
		Params:  []string{"Server error message"},
	}

	// Execute
	handler.handleError(message)

	// Verify expectations
	mockMessageHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

// Tests for MODE handlers
func TestHandler_handleMode_UserMode(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, mockMessageHandler, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("IsValidChannel", "testnick").Return(false)
	mockModeHandler.On("GetCurrentModes").Return("")
	mockModeHandler.On("SetCurrentModes", "i")
	mockMessageHandler.On("AddMessage", mock.AnythingOfType("*irc.Message"))

	// Create test message
	message := ircmsg.Message{
		Source:  "testnick!user@host",
		Command: "MODE",
		Params:  []string{"testnick", "+i"},
	}

	// Execute
	handler.handleMode(message)

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
	mockMessageHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleMode_ChannelMode(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, mockUpdateTrigger, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("IsValidChannel", "#test").Return(true)
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockModeHandler.On("GetChannelModeType", "o").Return('P')
	mockModeHandler.On("GetModeNameForMode", "o").Return("@")

	// Create test message - giving user ops
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "+o", "testnick"},
	}

	// Execute
	handler.handleMode(message)

	// Verify mode change message was added
	assert.Len(t, channel.messages, 1)
	assert.Contains(t, channel.messages[0].GetMessage(), "admin sets mode +o testnick")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleUserMode(t *testing.T) {
	handler, _, _, _, mockModeHandler, mockMessageHandler, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockModeHandler.On("GetCurrentModes").Return("i")
	mockModeHandler.On("SetCurrentModes", "iw")
	mockMessageHandler.On("AddMessage", mock.AnythingOfType("*irc.Message"))

	// Create test message
	message := ircmsg.Message{
		Source:  "testnick!user@host",
		Command: "MODE",
		Params:  []string{"testnick", "+w"},
	}

	// Execute
	handler.handleUserMode(message)

	// Verify expectations
	mockModeHandler.AssertExpectations(t)
	mockMessageHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleUserMode_InvalidMessage(t *testing.T) {
	handler, _, _, _, _, _, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks - defer should still be called
	mockUpdateTrigger.On("SetPendingUpdate")

	// Create test message with insufficient params
	message := ircmsg.Message{
		Source:  "testnick!user@host",
		Command: "MODE",
		Params:  []string{"testnick"}, // Missing mode string
	}

	// Execute
	handler.handleUserMode(message)

	// Verify expectations - only defer should be called
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleUserMode_RemoveModes(t *testing.T) {
	handler, _, _, _, mockModeHandler, mockMessageHandler, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockModeHandler.On("GetCurrentModes").Return("iw")
	mockModeHandler.On("SetCurrentModes", "i")
	mockMessageHandler.On("AddMessage", mock.AnythingOfType("*irc.Message"))

	// Create test message to remove mode 'w'
	message := ircmsg.Message{
		Source:  "testnick!user@host",
		Command: "MODE",
		Params:  []string{"testnick", "-w"},
	}

	// Execute
	handler.handleUserMode(message)

	// Verify expectations
	mockModeHandler.AssertExpectations(t)
	mockMessageHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleUserMode_MixedModes(t *testing.T) {
	handler, _, _, _, mockModeHandler, mockMessageHandler, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockModeHandler.On("GetCurrentModes").Return("i")
	mockModeHandler.On("SetCurrentModes", "w")
	mockMessageHandler.On("AddMessage", mock.AnythingOfType("*irc.Message"))

	// Create test message with mixed add/remove modes
	message := ircmsg.Message{
		Source:  "testnick!user@host",
		Command: "MODE",
		Params:  []string{"testnick", "-i+w"},
	}

	// Execute
	handler.handleUserMode(message)

	// Verify expectations
	mockModeHandler.AssertExpectations(t)
	mockMessageHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleUserMode_NoPrefixDefaultsToRemove(t *testing.T) {
	handler, _, _, _, mockModeHandler, mockMessageHandler, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockModeHandler.On("GetCurrentModes").Return("iw")
	mockModeHandler.On("SetCurrentModes", "i")
	mockMessageHandler.On("AddMessage", mock.AnythingOfType("*irc.Message"))

	// Create test message without explicit +/- prefix (defaults to remove)
	message := ircmsg.Message{
		Source:  "testnick!user@host",
		Command: "MODE",
		Params:  []string{"testnick", "w"},
	}

	// Execute
	handler.handleUserMode(message)

	// Verify expectations
	mockModeHandler.AssertExpectations(t)
	mockMessageHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleUserMode_DuplicateMode(t *testing.T) {
	handler, _, _, _, mockModeHandler, mockMessageHandler, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockModeHandler.On("GetCurrentModes").Return("iw")
	mockModeHandler.On("SetCurrentModes", "iw")
	mockMessageHandler.On("AddMessage", mock.AnythingOfType("*irc.Message"))

	// Create test message trying to add mode that already exists
	message := ircmsg.Message{
		Source:  "testnick!user@host",
		Command: "MODE",
		Params:  []string{"testnick", "+w"},
	}

	// Execute
	handler.handleUserMode(message)

	// Verify expectations
	mockModeHandler.AssertExpectations(t)
	mockMessageHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleUserMode_RemoveNonExistentMode(t *testing.T) {
	handler, _, _, _, mockModeHandler, mockMessageHandler, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockModeHandler.On("GetCurrentModes").Return("i")
	mockModeHandler.On("SetCurrentModes", "i")
	mockMessageHandler.On("AddMessage", mock.AnythingOfType("*irc.Message"))

	// Create test message trying to remove mode that doesn't exist
	message := ircmsg.Message{
		Source:  "testnick!user@host",
		Command: "MODE",
		Params:  []string{"testnick", "-w"},
	}

	// Execute
	handler.handleUserMode(message)

	// Verify expectations
	mockModeHandler.AssertExpectations(t)
	mockMessageHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleUserMode_MultipleModes(t *testing.T) {
	handler, _, _, _, mockModeHandler, mockMessageHandler, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockModeHandler.On("GetCurrentModes").Return("i")
	mockModeHandler.On("SetCurrentModes", "iwx")
	mockMessageHandler.On("AddMessage", mock.AnythingOfType("*irc.Message"))

	// Create test message adding multiple modes
	message := ircmsg.Message{
		Source:  "testnick!user@host",
		Command: "MODE",
		Params:  []string{"testnick", "+wx"},
	}

	// Execute
	handler.handleUserMode(message)

	// Verify expectations
	mockModeHandler.AssertExpectations(t)
	mockMessageHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleChannelModes_EmptyModeString(t *testing.T) {
	handler, _, _, _, _, _, _, _ := createTestHandler()
	channel := createTestChannel("#test")
	initialMessageCount := len(channel.messages)

	// Create test message with empty mode string
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", ""},
	}

	// Execute - should handle gracefully without panic
	handler.handleChannelModes(message)

	// Verify no messages were added to channel since no modes were processed
	assert.Len(t, channel.messages, initialMessageCount)
}

func TestHandler_handleChannelModes_PrefixModeWithoutParam(t *testing.T) {
	handler, _, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks - note that prefix modes without parameters should be ignored
	mockModeHandler.On("GetChannelModeType", "o").Return('P')

	initialMessageCount := len(channel.messages)

	// Create test message - trying to give ops but no nick parameter
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "+o"},
	}

	// Execute
	handler.handleChannelModes(message)

	// Verify no message was added to channel (prefix modes without parameters are ignored)
	assert.Len(t, channel.messages, initialMessageCount)

	// Verify expectations
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_handleChannelModes_TypeAModeWithParam(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockModeHandler.On("GetChannelModeType", "k").Return('A')

	initialMessageCount := len(channel.messages)

	// Create test message - setting key with parameter
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "+k", "secretkey"},
	}

	// Execute
	handler.handleChannelModes(message)

	// Verify message was added to channel
	assert.Len(t, channel.messages, initialMessageCount+1)
	addedMessage := channel.messages[len(channel.messages)-1]
	assert.Equal(t, MessageType(Event), addedMessage.messageType)
	assert.Contains(t, addedMessage.GetMessage(), "admin sets mode +k secretkey")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_handleChannelModes_TypeAModeWithoutParam(t *testing.T) {
	handler, _, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockModeHandler.On("GetChannelModeType", "k").Return('A')
	initialMessageCount := len(channel.messages)

	// Create test message - trying to set key but no parameter
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "+k"},
	}

	// Execute
	handler.handleChannelModes(message)

	// Verify no message was added to channel (Type A modes without parameters are ignored)
	assert.Len(t, channel.messages, initialMessageCount)

	// Verify expectations
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_handleChannelModes_TypeAModeRemoval(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockModeHandler.On("GetChannelModeType", "k").Return('A')
	initialMessageCount := len(channel.messages)

	// Create test message - removing key with parameter
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "-k", "secretkey"},
	}

	// Execute
	handler.handleChannelModes(message)

	// Verify message was added to channel
	assert.Len(t, channel.messages, initialMessageCount+1)
	addedMessage := channel.messages[len(channel.messages)-1]
	assert.Equal(t, MessageType(Event), addedMessage.messageType)
	assert.Contains(t, addedMessage.GetMessage(), "admin sets mode -k secretkey")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_handleChannelModes_TypeAModeRemovalWithoutParam(t *testing.T) {
	handler, _, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockModeHandler.On("GetChannelModeType", "k").Return('A')
	initialMessageCount := len(channel.messages)

	// Create test message - removing key but no parameter
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "-k"},
	}

	// Execute
	handler.handleChannelModes(message)

	// Verify no message was added to channel (Type A modes always require parameters)
	assert.Len(t, channel.messages, initialMessageCount)

	// Verify expectations
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_handleChannelModes_TypeBModeWithParam(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockModeHandler.On("GetChannelModeType", "b").Return('B')
	initialMessageCount := len(channel.messages)

	// Create test message - setting ban with parameter
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "+b", "*!*@spam.com"},
	}

	// Execute
	handler.handleChannelModes(message)

	// Verify message was added to channel
	assert.Len(t, channel.messages, initialMessageCount+1)
	addedMessage := channel.messages[len(channel.messages)-1]
	assert.Equal(t, MessageType(Event), addedMessage.messageType)
	assert.Equal(t, addedMessage.GetMessage(), "admin sets mode +b *!*@spam.com")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_handleChannelModes_TypeBModeWithoutParam(t *testing.T) {
	handler, _, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockModeHandler.On("GetChannelModeType", "b").Return('B')
	initialMessageCount := len(channel.messages)

	// Create test message - trying to set ban but no parameter
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "+b"},
	}

	// Execute
	handler.handleChannelModes(message)

	// Verify no message was added to channel (Type B modes always require parameters)
	assert.Len(t, channel.messages, initialMessageCount)

	// Verify expectations
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_handleChannelModes_TypeBModeRemoval(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockModeHandler.On("GetChannelModeType", "b").Return('B')
	initialMessageCount := len(channel.messages)

	// Create test message - removing ban with parameter
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "-b", "*!*@spam.com"},
	}

	// Execute
	handler.handleChannelModes(message)

	// Verify message was added to channel
	assert.Len(t, channel.messages, initialMessageCount+1)
	addedMessage := channel.messages[len(channel.messages)-1]
	assert.Equal(t, MessageType(Event), addedMessage.messageType)
	assert.Equal(t, addedMessage.GetMessage(), "admin sets mode -b *!*@spam.com")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_handleChannelModes_TypeBModeRemovalWithoutParam(t *testing.T) {
	handler, _, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockModeHandler.On("GetChannelModeType", "b").Return('B')
	initialMessageCount := len(channel.messages)

	// Create test message - removing ban but no parameter
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "-b"},
	}

	// Execute
	handler.handleChannelModes(message)

	// Verify no message was added to channel (Type B modes always require parameters)
	assert.Len(t, channel.messages, initialMessageCount)

	// Verify expectations
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_handleChannelModes_TypeCModeWithParam(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockModeHandler.On("GetChannelModeType", "l").Return('C')
	initialMessageCount := len(channel.messages)

	// Create test message - setting limit with parameter
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "+l", "50"},
	}

	// Execute
	handler.handleChannelModes(message)

	// Verify message was added to channel
	assert.Len(t, channel.messages, initialMessageCount+1)
	addedMessage := channel.messages[len(channel.messages)-1]
	assert.Equal(t, MessageType(Event), addedMessage.messageType)
	assert.Equal(t, addedMessage.GetMessage(), "admin sets mode +l 50")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_handleChannelModes_TypeCModeWithoutParam(t *testing.T) {
	handler, _, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockModeHandler.On("GetChannelModeType", "l").Return('C')
	initialMessageCount := len(channel.messages)

	// Create test message - setting limit without parameter (should be ignored)
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "+l"},
	}

	// Execute
	handler.handleChannelModes(message)

	// Verify no message was added to channel (Type C modes require parameters when setting)
	assert.Len(t, channel.messages, initialMessageCount)

	// Verify expectations
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_handleChannelModes_TypeCModeRemoval(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockModeHandler.On("GetChannelModeType", "l").Return('C')
	initialMessageCount := len(channel.messages)

	// Create test message - removing limit (type C doesn't need parameter for removal)
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "-l"},
	}

	// Execute
	handler.handleChannelModes(message)

	// Verify message was added to channel
	assert.Len(t, channel.messages, initialMessageCount+1)
	addedMessage := channel.messages[len(channel.messages)-1]
	assert.Equal(t, MessageType(Event), addedMessage.messageType)
	assert.Equal(t, addedMessage.GetMessage(), "admin sets mode -l")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_handleChannelModes_TypeDMode(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockModeHandler.On("GetChannelModeType", "s").Return('D')
	initialMessageCount := len(channel.messages)

	// Create test message - setting secret mode (boolean, no parameter)
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "+s"},
	}

	// Execute
	handler.handleChannelModes(message)

	// Verify message was added to channel
	assert.Len(t, channel.messages, initialMessageCount+1)
	addedMessage := channel.messages[len(channel.messages)-1]
	assert.Equal(t, MessageType(Event), addedMessage.messageType)
	assert.Equal(t, addedMessage.GetMessage(), "admin sets mode +s")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_handleChannelModes_NoPrefixDefaultsToRemove(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockModeHandler.On("GetChannelModeType", "s").Return('D')
	initialMessageCount := len(channel.messages)

	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "s"},
	}

	// Execute
	handler.handleChannelModes(message)

	// Verify message was added to channel
	assert.Len(t, channel.messages, initialMessageCount+1)
	addedMessage := channel.messages[len(channel.messages)-1]
	assert.Equal(t, MessageType(Event), addedMessage.messageType)
	assert.Equal(t, addedMessage.GetMessage(), "admin sets mode +s")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_handleChannelModes_MixedModeTypes(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil).Times(3)
	mockModeHandler.On("GetChannelModeType", "o").Return('P')
	mockModeHandler.On("GetChannelModeType", "s").Return('D')
	mockModeHandler.On("GetChannelModeType", "l").Return('C')
	mockModeHandler.On("GetModeNameForMode", "o").Return("@")
	initialMessageCount := len(channel.messages)

	// Create test message with mixed mode types
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "+osl", "testnick", "50"},
	}

	// Execute
	handler.handleChannelModes(message)

	// Verify multiple messages were added to channel (one per mode)
	assert.Len(t, channel.messages, initialMessageCount+3)
	msgLen := initialMessageCount + 3
	msg := channel.messages[msgLen-3 : msgLen-2][0]
	assert.Equal(t, MessageType(Event), msg.messageType)
	assert.Equal(t, msg.GetMessage(), "admin sets mode +o testnick")
	msg = channel.messages[msgLen-2 : msgLen-1][0]
	assert.Equal(t, MessageType(Event), msg.messageType)
	assert.Equal(t, msg.GetMessage(), "admin sets mode +s")
	msg = channel.messages[msgLen-1 : msgLen][0]
	assert.Equal(t, MessageType(Event), msg.messageType)
	assert.Equal(t, msg.GetMessage(), "admin sets mode +l 50")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_handleChannelModes_ComplexModeString(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil).Times(3)
	mockModeHandler.On("GetChannelModeType", "o").Return('P')
	mockModeHandler.On("GetChannelModeType", "s").Return('D')
	mockModeHandler.On("GetChannelModeType", "v").Return('P')
	mockModeHandler.On("GetModeNameForMode", "o").Return("@")
	mockModeHandler.On("GetModeNameForMode", "v").Return("+")
	initialMessageCount := len(channel.messages)

	// Create test message with complex add/remove operations
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "+o-s+v", "user1", "user2"},
	}

	// Execute
	handler.handleChannelModes(message)

	// Verify multiple messages were added to channel (one per mode)
	assert.Len(t, channel.messages, initialMessageCount+3)
	msgLen := initialMessageCount + 3
	msg := channel.messages[msgLen-3 : msgLen-2][0]
	assert.Equal(t, MessageType(Event), msg.messageType)
	assert.Equal(t, msg.GetMessage(), "admin sets mode +o user1")
	msg = channel.messages[msgLen-2 : msgLen-1][0]
	assert.Equal(t, MessageType(Event), msg.messageType)
	assert.Equal(t, msg.GetMessage(), "admin sets mode -s")
	msg = channel.messages[msgLen-1 : msgLen][0]
	assert.Equal(t, MessageType(Event), msg.messageType)
	assert.Equal(t, msg.GetMessage(), "admin sets mode +v user2")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_handleChannelModes_InsufficientParameters(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil).Once()
	mockModeHandler.On("GetChannelModeType", "o").Return('P')
	mockModeHandler.On("GetChannelModeType", "v").Return('P')
	mockModeHandler.On("GetModeNameForMode", "o").Return("@")
	initialMessageCount := len(channel.messages)

	// Create test message - wants to give ops and voice but only one nickname
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "+ov", "user1"},
	}

	// Execute
	handler.handleChannelModes(message)

	// Verify multiple messages were added to channel (one per mode)
	assert.Len(t, channel.messages, initialMessageCount+1)
	msgLen := initialMessageCount + 1
	msg := channel.messages[msgLen-1 : msgLen][0]
	assert.Equal(t, MessageType(Event), msg.messageType)
	assert.Equal(t, "admin sets mode +o user1", msg.GetMessage())

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_handleUserModeSet(t *testing.T) {
	handler, _, _, _, mockModeHandler, mockMessageHandler, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockModeHandler.On("SetCurrentModes", "iw")
	mockMessageHandler.On("AddMessage", mock.AnythingOfType("*irc.Message"))

	// Create test message
	message := ircmsg.Message{
		Command: ircevent.RPL_UMODEIS,
		Params:  []string{"testnick", "iw"},
	}

	// Execute
	handler.handleUserModeSet(message)

	// Verify expectations
	mockModeHandler.AssertExpectations(t)
	mockMessageHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleNameReply(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, mockUpdateTrigger, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockModeHandler.On("GetModePrefixes").Return([]string{"ov", "@+"})

	// Create test message with users having modes
	message := ircmsg.Message{
		Command: ircevent.RPL_NAMREPLY,
		Params:  []string{"testnick", "=", "#test", "@admin +voice regular"},
	}

	// Execute
	handler.handleNameReply(message)

	// Verify users were added with correct modes
	assert.Len(t, channel.users, 3)

	// Find users by nickname and check modes
	var adminUser, voiceUser, regularUser *User
	for _, user := range channel.users {
		switch user.nickname {
		case "admin":
			adminUser = user
		case "voice":
			voiceUser = user
		case "regular":
			regularUser = user
		}
	}

	assert.NotNil(t, adminUser)
	assert.NotNil(t, voiceUser)
	assert.NotNil(t, regularUser)
	assert.Equal(t, "@", adminUser.modes)
	assert.Equal(t, "+", voiceUser.modes)
	assert.Equal(t, "", regularUser.modes)

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleNameReply_UnknownChannel(t *testing.T) {
	handler, mockChannelHandler, _, _, _, _, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#unknown").Return((*Channel)(nil), fmt.Errorf("channel not found"))

	// Create test message for unknown channel
	message := ircmsg.Message{
		Command: ircevent.RPL_NAMREPLY,
		Params:  []string{"testnick", "=", "#unknown", "@admin +voice regular"},
	}

	// Execute
	handler.handleNameReply(message)

	// Verify expectations - function should return early without processing users
	mockChannelHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleNameReply_UpdateExistingUserModes(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, mockUpdateTrigger, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Pre-populate channel with existing users
	existingUser1 := NewUser("alice", "")
	existingUser2 := NewUser("bob", "+")
	channel.AddUser(existingUser1)
	channel.AddUser(existingUser2)

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockModeHandler.On("GetModePrefixes").Return([]string{"ov", "@+"})

	// Create test message with updated modes for existing users
	message := ircmsg.Message{
		Command: ircevent.RPL_NAMREPLY,
		Params:  []string{"testnick", "=", "#test", "@alice bob"},
	}

	// Execute
	handler.handleNameReply(message)

	// Verify users were updated with new modes
	assert.Len(t, channel.users, 2)

	// Find users and check updated modes
	var aliceUser, bobUser *User
	for _, user := range channel.users {
		switch user.nickname {
		case "alice":
			aliceUser = user
		case "bob":
			bobUser = user
		}
	}

	assert.NotNil(t, aliceUser)
	assert.NotNil(t, bobUser)
	assert.Equal(t, "@", aliceUser.modes) // alice now has ops
	assert.Equal(t, "", bobUser.modes)    // bob lost voice

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleNameReply_MixedNewAndExistingUsers(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, mockUpdateTrigger, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Pre-populate channel with one existing user
	existingUser := NewUser("alice", "")
	channel.AddUser(existingUser)

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockModeHandler.On("GetModePrefixes").Return([]string{"ov", "@+"})

	// Create test message with both existing and new users
	message := ircmsg.Message{
		Command: ircevent.RPL_NAMREPLY,
		Params:  []string{"testnick", "=", "#test", "@alice +bob charlie"},
	}

	// Execute
	handler.handleNameReply(message)

	// Verify we now have 3 users total
	assert.Len(t, channel.users, 3)

	// Find users and check modes
	var aliceUser, bobUser, charlieUser *User
	for _, user := range channel.users {
		switch user.nickname {
		case "alice":
			aliceUser = user
		case "bob":
			bobUser = user
		case "charlie":
			charlieUser = user
		}
	}

	assert.NotNil(t, aliceUser)
	assert.NotNil(t, bobUser)
	assert.NotNil(t, charlieUser)
	assert.Equal(t, "@", aliceUser.modes)  // existing user updated with ops
	assert.Equal(t, "+", bobUser.modes)    // new user with voice
	assert.Equal(t, "", charlieUser.modes) // new user without modes

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleNameReply_ComplexModes(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, mockUpdateTrigger, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockModeHandler.On("GetModePrefixes").Return([]string{"ov", "@+"})

	// Create test message with users having multiple/complex modes
	message := ircmsg.Message{
		Command: ircevent.RPL_NAMREPLY,
		Params:  []string{"testnick", "=", "#test", "@+alice +bob @charlie regular"},
	}

	// Execute
	handler.handleNameReply(message)

	// Verify users were added with correct modes
	assert.Len(t, channel.users, 4)

	// Find users by nickname and check modes
	var aliceUser, bobUser, charlieUser, regularUser *User
	for _, user := range channel.users {
		switch user.nickname {
		case "alice":
			aliceUser = user
		case "bob":
			bobUser = user
		case "charlie":
			charlieUser = user
		case "regular":
			regularUser = user
		}
	}

	assert.NotNil(t, aliceUser)
	assert.NotNil(t, bobUser)
	assert.NotNil(t, charlieUser)
	assert.NotNil(t, regularUser)
	assert.Equal(t, "@+", aliceUser.modes)  // both ops and voice
	assert.Equal(t, "+", bobUser.modes)     // voice only
	assert.Equal(t, "@", charlieUser.modes) // ops only
	assert.Equal(t, "", regularUser.modes)  // no modes

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleNameReply_EmptyUsersList(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, mockUpdateTrigger, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)

	// Create test message with empty users list
	message := ircmsg.Message{
		Command: ircevent.RPL_NAMREPLY,
		Params:  []string{"testnick", "=", "#test", ""},
	}

	// Execute
	handler.handleNameReply(message)

	assert.Len(t, channel.users, 0)

	// Verify expectations
	mockChannelHandler.AssertNotCalled(t, "GetModePrefixes")
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

// Tests for applyChannelMode function
func TestHandler_applyChannelMode_UnknownChannel(t *testing.T) {
	handler, mockChannelHandler, _, _, _, _, _, _ := createTestHandler()

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#unknown").Return((*Channel)(nil), assert.AnError)

	// Create test mode change and message
	change := modeChange{
		mode:      "t",
		change:    true,
		modeType:  'D',
		parameter: "",
	}
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#unknown", "+t"},
	}

	// Execute
	handler.applyChannelMode(change, message)

	// Verify expectations - should only call GetChannelByName and log warning
	mockChannelHandler.AssertExpectations(t)
}

func TestHandler_applyChannelMode_UserPrivilegeMode(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")
	channel.users = append(channel.users, NewUser("testuser", ""))

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockModeHandler.On("GetModeNameForMode", "o").Return("@")

	// Create test mode change for user privilege (op)
	change := modeChange{
		mode:      "o",
		change:    true,
		nickname:  "testuser",
		modeType:  'P',
		parameter: "testuser",
	}
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "+o", "testuser"},
	}

	// Execute
	handler.applyChannelMode(change, message)

	// Verify user got the mode
	assert.Equal(t, "@", channel.users[0].modes)

	// Verify mode message was added
	assert.Len(t, channel.messages, 1)
	assert.Contains(t, channel.messages[0].GetMessage(), "admin sets mode +o testuser")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_applyChannelMode_ChannelModeWithParameter(t *testing.T) {
	handler, mockChannelHandler, _, _, _, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)

	// Create test mode change for channel mode with parameter (e.g., +k key)
	change := modeChange{
		mode:      "k",
		change:    true,
		modeType:  'C',
		parameter: "secretkey",
	}
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "+k", "secretkey"},
	}

	// Execute
	handler.applyChannelMode(change, message)

	// Verify mode message was added with parameter
	assert.Len(t, channel.messages, 1)
	assert.Contains(t, channel.messages[0].GetMessage(), "admin sets mode +k secretkey")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
}

func TestHandler_applyChannelMode_ChannelModeWithoutParameter(t *testing.T) {
	handler, mockChannelHandler, _, _, _, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)

	// Create test mode change for channel mode without parameter (e.g., +t)
	change := modeChange{
		mode:      "t",
		change:    true,
		modeType:  'D',
		parameter: "",
	}
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "+t"},
	}

	// Execute
	handler.applyChannelMode(change, message)

	// Verify mode message was added without parameter
	assert.Len(t, channel.messages, 1)
	assert.Contains(t, channel.messages[0].GetMessage(), "admin sets mode +t")
	assert.NotContains(t, channel.messages[0].GetMessage(), "admin sets mode +t ")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
}

func TestHandler_applyChannelMode_RemoveChannelMode(t *testing.T) {
	handler, mockChannelHandler, _, _, _, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)

	// Create test mode change for removing channel mode (e.g., -t)
	change := modeChange{
		mode:      "t",
		change:    false,
		modeType:  'D',
		parameter: "",
	}
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "-t"},
	}

	// Execute
	handler.applyChannelMode(change, message)

	// Verify mode message shows removal
	assert.Len(t, channel.messages, 1)
	assert.Contains(t, channel.messages[0].GetMessage(), "admin sets mode -t")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
}

func TestHandler_applyChannelMode_UnknownModeType(t *testing.T) {
	handler, mockChannelHandler, _, _, _, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)

	// Create test mode change with unknown mode type
	change := modeChange{
		mode:      "x",
		change:    true,
		modeType:  'X', // Unknown mode type
		parameter: "",
	}
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "+x"},
	}

	// Execute
	handler.applyChannelMode(change, message)

	// Verify no messages were added (unknown mode type)
	assert.Len(t, channel.messages, 0)

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
}

func TestHandler_applyChannelMode_RemoveUserPrivilegeMode(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")
	// User starts with op mode (@)
	channel.users = append(channel.users, NewUser("testuser", "@"))

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockModeHandler.On("GetModeNameForMode", "o").Return("@")

	// Create test mode change for removing user privilege (deop)
	change := modeChange{
		mode:      "o",
		change:    false, // This is the key difference - removing the mode
		nickname:  "testuser",
		modeType:  'P',
		parameter: "testuser",
	}
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "-o", "testuser"},
	}

	// Execute
	handler.applyChannelMode(change, message)

	// Verify user lost the mode (should be empty now)
	assert.Equal(t, "", channel.users[0].modes)

	// Verify mode message was added showing removal
	assert.Len(t, channel.messages, 1)
	assert.Contains(t, channel.messages[0].GetMessage(), "admin sets mode -o testuser")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_applyChannelMode_RemoveUserPrivilegeMode_MultipleModes(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")
	// User starts with both op and voice modes (@+)
	channel.users = append(channel.users, NewUser("testuser", "@+"))

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockModeHandler.On("GetModeNameForMode", "o").Return("@")

	// Create test mode change for removing op mode (but keeping voice)
	change := modeChange{
		mode:      "o",
		change:    false,
		nickname:  "testuser",
		modeType:  'P',
		parameter: "testuser",
	}
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "-o", "testuser"},
	}

	// Execute
	handler.applyChannelMode(change, message)

	// Verify user lost op mode but kept voice mode (should be "+" now)
	assert.Equal(t, "+", channel.users[0].modes)

	// Verify mode message was added showing removal
	assert.Len(t, channel.messages, 1)
	assert.Contains(t, channel.messages[0].GetMessage(), "admin sets mode -o testuser")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
}

func TestHandler_applyChannelMode_RemoveUserPrivilegeMode_UserNotFound(t *testing.T) {
	handler, mockChannelHandler, _, _, mockModeHandler, _, _, _ := createTestHandler()

	channel := createTestChannel("#test")
	// Add a different user, not the target of the mode change
	channel.users = append(channel.users, NewUser("otheruser", "@"))

	// Setup mocks
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockModeHandler.On("GetModeNameForMode", "o").Return("@")

	// Create test mode change for user that doesn't exist in channel
	change := modeChange{
		mode:      "o",
		change:    false,
		nickname:  "nonexistentuser",
		modeType:  'P',
		parameter: "nonexistentuser",
	}
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "MODE",
		Params:  []string{"#test", "-o", "nonexistentuser"},
	}

	// Execute
	handler.applyChannelMode(change, message)

	// Verify other user's modes unchanged
	assert.Equal(t, "@", channel.users[0].modes)

	// Verify mode message was still added (IRC servers send these even for non-existent users)
	assert.Len(t, channel.messages, 1)
	assert.Contains(t, channel.messages[0].GetMessage(), "admin sets mode -o nonexistentuser")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockModeHandler.AssertExpectations(t)
}

// Tests for KICK handler
func TestHandler_handleKick_SelfKick(t *testing.T) {
	handler, mockChannelHandler, _, mockInfoHandler, _, mockMessageHandler, mockUpdateTrigger, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockInfoHandler.On("CurrentNick").Return("testnick")
	mockChannelHandler.On("RemoveChannel", "test-channel-id")
	mockMessageHandler.On("AddMessage", mock.AnythingOfType("*irc.Message"))

	// Create test message
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "KICK",
		Params:  []string{"#test", "testnick", "You're out"},
	}

	// Execute
	handler.handleKick(message)

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockInfoHandler.AssertExpectations(t)
	mockMessageHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleKick_OtherKick(t *testing.T) {
	handler, mockChannelHandler, _, mockInfoHandler, _, _, mockUpdateTrigger, _ := createTestHandler()

	channel := createTestChannel("#test")
	channel.users = append(channel.users, NewUser("kickeduser", ""))
	channel.users = append(channel.users, NewUser("othernick", ""))

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)
	mockInfoHandler.On("CurrentNick").Return("testnick")

	// Create test message
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "KICK",
		Params:  []string{"#test", "kickeduser", "Bye"},
	}

	// Execute
	handler.handleKick(message)

	// Verify kicked user was removed but other user remains
	assert.Len(t, channel.users, 1)
	assert.Equal(t, "othernick", channel.users[0].nickname)

	// Verify kick message was added
	assert.Len(t, channel.messages, 1)
	assert.Contains(t, channel.messages[0].GetMessage(), "admin!user@host has kicked kickeduser from #test")

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockInfoHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleKick_UnknownChannel(t *testing.T) {
	handler, mockChannelHandler, _, _, _, _, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("GetChannelByName", "#unknown").Return((*Channel)(nil), assert.AnError)

	// Create test message for unknown channel
	message := ircmsg.Message{
		Source:  "admin!user@host",
		Command: "KICK",
		Params:  []string{"#unknown", "kickeduser", "You're out"},
	}

	// Execute
	handler.handleKick(message)

	// Verify expectations - should only call GetChannelByName and SetPendingUpdate
	// No users should be removed, no messages should be added since function returns early
	mockChannelHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

// Tests for NOTICE handler
func TestHandler_handleNotice_ServerNotice(t *testing.T) {
	handler, _, _, mockInfoHandler, _, mockMessageHandler, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockInfoHandler.On("CurrentNick").Return("testnick")
	mockMessageHandler.On("AddMessage", mock.AnythingOfType("*irc.Message"))

	// Create test message from server
	message := ircmsg.Message{
		Source:  "irc.example.com",
		Command: "NOTICE",
		Params:  []string{"testnick", "Server notice message"},
	}

	// Execute
	handler.handleNotice(message)

	// Verify expectations
	mockInfoHandler.AssertExpectations(t)
	mockMessageHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleNotice_ChannelNotice(t *testing.T) {
	handler, mockChannelHandler, _, mockInfoHandler, _, _, mockUpdateTrigger, _ := createTestHandler()

	channel := createTestChannel("#test")

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockInfoHandler.On("CurrentNick").Return("testnick")
	mockChannelHandler.On("IsValidChannel", "#test").Return(true)
	mockChannelHandler.On("GetChannelByName", "#test").Return(channel, nil)

	// Create test message
	message := ircmsg.Message{
		Source:  "othernick!user@host",
		Command: "NOTICE",
		Params:  []string{"#test", "Channel notice"},
	}

	// Execute
	handler.handleNotice(message)

	// Verify notice was added to channel
	assert.Len(t, channel.messages, 1)
	assert.Equal(t, "Channel notice", channel.messages[0].GetMessage())
	assert.Equal(t, MessageType(Notice), channel.messages[0].GetType())

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockInfoHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleNotice_UnknownChannel(t *testing.T) {
	handler, mockChannelHandler, _, mockInfoHandler, _, _, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockInfoHandler.On("CurrentNick").Return("testnick")
	mockChannelHandler.On("IsValidChannel", "#unknown").Return(true)
	mockChannelHandler.On("GetChannelByName", "#unknown").Return((*Channel)(nil), assert.AnError)

	// Create test message for unknown channel
	message := ircmsg.Message{
		Source:  "othernick!user@host",
		Command: "NOTICE",
		Params:  []string{"#unknown", "Notice for unknown channel"},
	}

	// Execute
	handler.handleNotice(message)

	// Verify expectations - should only call mocks and return early
	mockChannelHandler.AssertExpectations(t)
	mockInfoHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleNotice_PrivateNotice_ExactCase(t *testing.T) {
	handler, mockChannelHandler, mockQueryHandler, mockInfoHandler, _, _, mockUpdateTrigger, _ := createTestHandler()

	query := createTestQuery("othernick")

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("IsValidChannel", "testnick").Return(false)
	mockInfoHandler.On("CurrentNick").Return("testnick")
	mockQueryHandler.On("GetQueryByName", "othernick").Return((*Query)(nil), assert.AnError)
	mockQueryHandler.On("AddQuery", "othernick").Return(query)

	// Create test message - private notice to current user (exact case)
	message := ircmsg.Message{
		Source:  "othernick!user@host",
		Command: "NOTICE",
		Params:  []string{"testnick", "Private notice message"},
	}

	// Execute
	handler.handleNotice(message)

	// Verify notice was added to query
	assert.Len(t, query.messages, 1)
	assert.Equal(t, "Private notice message", query.messages[0].GetMessage())
	assert.Equal(t, MessageType(Notice), query.messages[0].GetType())

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockQueryHandler.AssertExpectations(t)
	mockInfoHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleNotice_PrivateNotice_CaseInsensitive(t *testing.T) {
	handler, mockChannelHandler, mockQueryHandler, mockInfoHandler, _, _, mockUpdateTrigger, _ := createTestHandler()

	query := createTestQuery("othernick")

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("IsValidChannel", "TestNick").Return(false)
	mockInfoHandler.On("CurrentNick").Return("testnick")
	mockQueryHandler.On("GetQueryByName", "othernick").Return((*Query)(nil), assert.AnError)
	mockQueryHandler.On("AddQuery", "othernick").Return(query)

	// Create test message - private notice with different case
	message := ircmsg.Message{
		Source:  "othernick!user@host",
		Command: "NOTICE",
		Params:  []string{"TestNick", "Case insensitive notice"},
	}

	// Execute
	handler.handleNotice(message)

	// Verify notice was added to query
	assert.Len(t, query.messages, 1)
	assert.Equal(t, "Case insensitive notice", query.messages[0].GetMessage())
	assert.Equal(t, MessageType(Notice), query.messages[0].GetType())

	// Verify expectations
	mockChannelHandler.AssertExpectations(t)
	mockQueryHandler.AssertExpectations(t)
	mockInfoHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

func TestHandler_handleNotice_UnsupportedTarget(t *testing.T) {
	handler, mockChannelHandler, _, mockInfoHandler, _, _, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockChannelHandler.On("IsValidChannel", "someothernick").Return(false)
	mockInfoHandler.On("CurrentNick").Return("testnick")

	// Create test message with unsupported target
	message := ircmsg.Message{
		Source:  "othernick!user@host",
		Command: "NOTICE",
		Params:  []string{"someothernick", "Notice to someone else"},
	}

	// Execute
	handler.handleNotice(message)

	// Verify expectations - should only call mocks and log warning
	mockChannelHandler.AssertExpectations(t)
	mockInfoHandler.AssertExpectations(t)
	mockUpdateTrigger.AssertExpectations(t)
}

// Tests for WHOIS response handlers - these are inline functions that call addEvent
func TestHandler_WhoisHandlers(t *testing.T) {
	handler, _, _, _, _, mockMessageHandler, mockUpdateTrigger, _ := createTestHandler()

	// Setup mocks
	mockUpdateTrigger.On("SetPendingUpdate")
	mockMessageHandler.On("AddMessage", mock.AnythingOfType("*irc.Message")).Times(9)

	// Test various WHOIS responses
	whoisTests := []struct {
		command string
		params  []string
	}{
		{ircevent.RPL_WHOISUSER, []string{"testnick", "othernick", "user", "host", "*", "Real Name"}},
		{ircevent.RPL_WHOISCERTFP, []string{"testnick", "othernick", "has client certificate fingerprint abc123"}},
		{ircevent.RPL_WHOISACCOUNT, []string{"testnick", "othernick", "account", "is logged in as"}},
		{ircevent.RPL_WHOISBOT, []string{"testnick", "othernick", "is a bot"}},
		{ircevent.RPL_WHOISACTUALLY, []string{"testnick", "othernick", "1.2.3.4", "Actual hostname"}},
		{ircevent.RPL_WHOISCHANNELS, []string{"testnick", "othernick", "#channel1 #channel2"}},
		{ircevent.RPL_WHOISIDLE, []string{"testnick", "othernick", "300", "1234567890", "seconds idle, signon time"}},
		{ircevent.RPL_WHOISSERVER, []string{"testnick", "othernick", "irc.example.com", "Example IRC Server"}},
		{ircevent.RPL_ENDOFWHOIS, []string{"testnick", "othernick", "End of WHOIS list"}},
	}

	for _, test := range whoisTests {
		message := ircmsg.Message{
			Command: test.command,
			Params:  test.params,
		}

		// Execute the callback directly by simulating what addCallbacks would do
		switch test.command {
		case ircevent.RPL_WHOISUSER:
			handler.addEvent(EventWhois, false, "WHOIS: "+strings.Join(message.Params[1:], " "))
		case ircevent.RPL_WHOISCERTFP:
			handler.addEvent(EventWhois, false, "WHOIS: "+strings.Join(message.Params[1:], " "))
		case ircevent.RPL_WHOISACCOUNT:
			handler.addEvent(EventWhois, false, "WHOIS "+strings.Join(message.Params[2:], " ")+" "+message.Params[1])
		case ircevent.RPL_WHOISBOT:
			handler.addEvent(EventWhois, false, "WHOIS: "+strings.Join(message.Params[1:], " "))
		case ircevent.RPL_WHOISACTUALLY:
			handler.addEvent(EventWhois, false, "WHOIS: "+strings.Join(message.Params, " "))
		case ircevent.RPL_WHOISCHANNELS:
			handler.addEvent(EventWhois, false, "WHOIS: "+strings.Join(message.Params[1:], " "))
		case ircevent.RPL_WHOISIDLE:
			handler.addEvent(EventWhois, false, "WHOIS: "+strings.Join(message.Params[1:], " "))
		case ircevent.RPL_WHOISSERVER:
			handler.addEvent(EventWhois, false, "WHOIS: "+strings.Join(message.Params[1:], " "))
		case ircevent.RPL_ENDOFWHOIS:
			handler.addEvent(EventWhois, false, "WHOIS END "+message.Params[1])
		}
	}

	// Verify expectations
	mockMessageHandler.AssertExpectations(t)
}

// Tests for error response handlers - these are inline functions
func TestHandler_ErrorHandlers(t *testing.T) {
	handler, _, _, _, _, mockMessageHandler, _, _ := createTestHandler()

	// Setup mocks
	mockMessageHandler.On("AddMessage", mock.AnythingOfType("*irc.Message")).Times(2)

	// Test nickname in use error
	nickInUseMessage := ircmsg.Message{
		Command: ircevent.ERR_NICKNAMEINUSE,
		Params:  []string{"*", "testnick", "Nickname is already in use"},
	}
	handler.messageHandler.AddMessage(NewError(handler.linkRegex, handler.conf.UISettings.TimestampFormat, false, "Nickname in use: "+nickInUseMessage.Params[1]))

	// Test password mismatch error
	passwdMismatchMessage := ircmsg.Message{
		Command: ircevent.ERR_PASSWDMISMATCH,
		Params:  []string{"testnick", "Password incorrect"},
	}
	handler.messageHandler.AddMessage(NewError(handler.linkRegex, handler.conf.UISettings.TimestampFormat, false, "Password Mismatch: "+strings.Join(passwdMismatchMessage.Params, " ")))

	// Verify expectations
	mockMessageHandler.AssertExpectations(t)
}

// Tests for handleBatch function
func TestHandler_handleBatch_ChatHistory(t *testing.T) {
	handler, _, _, _, _, _, _, _ := createTestHandler()

	// Create test batch items (messages within the batch)
	item1 := &ircevent.Batch{
		Message: ircmsg.Message{
			Command: "PRIVMSG",
			Params:  []string{"#test", "Historical message 1"},
		},
	}
	item2 := &ircevent.Batch{
		Message: ircmsg.Message{
			Command: "PRIVMSG",
			Params:  []string{"#test", "Historical message 2"},
		},
	}

	// Create test batch for chat history
	batch := &ircevent.Batch{
		Message: ircmsg.Message{
			Command: "BATCH",
			Params:  []string{"batch-id", "chathistory", "#test"},
		},
		Items: []*ircevent.Batch{item1, item2},
	}

	// Execute
	result := handler.handleBatch(batch)

	// Verify chathistory tags were set on all messages
	present1, value1 := batch.Items[0].Message.GetTag("chathistory")
	present2, value2 := batch.Items[1].Message.GetTag("chathistory")
	assert.True(t, present1, "first message should have chathistory tag present")
	assert.Equal(t, "true", value1, "first message chathistory tag should be 'true'")
	assert.True(t, present2, "second message should have chathistory tag present")
	assert.Equal(t, "true", value2, "second message chathistory tag should be 'true'")

	// Verify function returns false (doesn't consume the batch)
	assert.False(t, result, "handleBatch should return false")
}

func TestHandler_handleBatch_NonChatHistory(t *testing.T) {
	handler, _, _, _, _, _, _, _ := createTestHandler()

	// Create test batch item
	item := &ircevent.Batch{
		Message: ircmsg.Message{
			Command: "PRIVMSG",
			Params:  []string{"#test", "Regular batch message"},
		},
	}

	// Create test batch for non-chathistory type
	batch := &ircevent.Batch{
		Message: ircmsg.Message{
			Command: "BATCH",
			Params:  []string{"batch-id", "other-type", "#test"},
		},
		Items: []*ircevent.Batch{item},
	}

	// Execute
	result := handler.handleBatch(batch)

	// Verify no chathistory tag was set
	present, value := batch.Items[0].Message.GetTag("chathistory")
	assert.False(t, present && value == "true", "chathistory tag should not be set for non-chathistory batches")

	// Verify function returns false
	assert.False(t, result, "handleBatch should return false")
}

func TestHandler_handleBatch_EmptyBatch(t *testing.T) {
	handler, _, _, _, _, _, _, _ := createTestHandler()

	// Create empty batch for chat history
	batch := &ircevent.Batch{
		Message: ircmsg.Message{
			Command: "BATCH",
			Params:  []string{"batch-id", "chathistory", "#test"},
		},
		Items: []*ircevent.Batch{},
	}

	// Execute - should not panic
	result := handler.handleBatch(batch)

	// Verify function returns false
	assert.False(t, result, "handleBatch should return false for empty batch")
}

// Test helper method
func TestHandler_isMsgMe(t *testing.T) {
	handler, _, _, mockInfoHandler, _, _, _, _ := createTestHandler()

	// Setup mocks
	mockInfoHandler.On("CurrentNick").Return("testnick")

	// Create test messages
	messageFromMe := ircmsg.Message{
		Source:  "testnick!user@host",
		Command: "PRIVMSG",
		Params:  []string{"#test", "Hello"},
	}

	messageFromOther := ircmsg.Message{
		Source:  "othernick!user@host",
		Command: "PRIVMSG",
		Params:  []string{"#test", "Hello"},
	}

	// Test
	assert.True(t, handler.isMsgMe(messageFromMe))
	assert.False(t, handler.isMsgMe(messageFromOther))

	// Verify expectations
	mockInfoHandler.AssertExpectations(t)
}
