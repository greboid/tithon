package irc

import (
	"errors"
	"github.com/greboid/tithon/config"
	"github.com/hueristiq/hq-go-url/extractor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"regexp"
	"testing"
)

type MockCommand struct {
	mock.Mock
}

func (m *MockCommand) GetArgSpecs() []Argument {
	return []Argument{}
}

func (m *MockCommand) GetFlagSpecs() []Flag {
	return []Flag{}
}

func (m *MockCommand) GetUsage() string {
	return GenerateDetailedHelp(m)
}

func (m *MockCommand) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCommand) GetHelp() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCommand) Execute(cm *ServerManager, w *Window, input string) error {
	args := m.Called(cm, w, input)
	return args.Error(0)
}

var regex *regexp.Regexp

func init() {
	regex = extractor.New(extractor.WithHost()).CompileRegex()
}

func createTestWindow() *Window {
	return &Window{
		id:       "test-id",
		name:     "test-window",
		title:    "Test Window",
		messages: make([]*Message, 0),
	}
}

func TestNewCommandManager(t *testing.T) {
	conf := getCommandManagerTestConfig()
	cm := NewCommandManager(conf, make(chan bool, 1))

	assert.NotNil(t, cm, "CommandManager should not be nil")
	assert.Equal(t, conf, cm.conf, "Config should be set correctly")
	assert.NotNil(t, cm.commands, "Commands should be initialized")
}

func TestCommandManager_Execute(t *testing.T) {
	cm := NewCommandManager(getCommandManagerTestConfig(), make(chan bool, 1))

	mockCmd := new(MockCommand)
	mockCmd.On("GetName").Return("test")
	mockCmd.On("Execute", mock.Anything, mock.Anything, "test input").Return(nil)
	cm.commands = []Command{mockCmd}

	window := &Window{}
	connections := &ServerManager{}

	cm.Execute(connections, window, "/test test input")

	mockCmd.AssertExpectations(t)
}

func TestCommandManager_Execute_NoArguments(t *testing.T) {
	cm := NewCommandManager(getCommandManagerTestConfig(), make(chan bool, 1))

	mockCmd := new(MockCommand)
	mockCmd.On("GetName").Return("test")
	mockCmd.On("Execute", mock.Anything, mock.Anything, "").Return(nil)
	cm.commands = []Command{mockCmd}

	window := &Window{}
	connections := &ServerManager{}

	cm.Execute(connections, window, "/test")

	mockCmd.AssertExpectations(t)
}

func TestCommandManager_Execute_NoMatch(t *testing.T) {
	cm := NewCommandManager(getCommandManagerTestConfig(), make(chan bool, 1))

	window := createTestWindow()

	cm.Execute(nil, window, "/nonexistent command")

	messages := window.GetMessages()
	assert.Len(t, messages, 1, "Should have added one error message")
	assert.Contains(t, messages[0].GetMessage(),
		"Command &#39;nonexistent command&#39; not found. Use /help to see all available commands.",
		"Should state command not found")
}

func TestCommandManager_Execute_Error(t *testing.T) {
	cm := NewCommandManager(getCommandManagerTestConfig(), make(chan bool, 1))

	mockCmd := new(MockCommand)
	mockCmd.On("GetName").Return("test")
	mockCmd.On("Execute", mock.Anything, mock.Anything, "test input").Return(errors.New("test error"))
	cm.commands = []Command{mockCmd}

	window := createTestWindow()

	cm.Execute(nil, window, "/test test input")

	mockCmd.AssertExpectations(t)
	messages := window.GetMessages()
	assert.Len(t, messages, 1, "Should have added one error message")
	assert.Contains(t, messages[0].GetMessage(), "Command Error: test: test error", "Error message should contain command name and error")
}

func TestCommandManager_Execute_InputNoSlash(t *testing.T) {
	cm := NewCommandManager(getCommandManagerTestConfig(), make(chan bool, 1))

	mockCmd := new(MockCommand)
	mockCmd.On("GetName").Return("test")
	mockCmd.AssertNotCalled(t, "Execute")
	msgCmd := new(MockCommand)
	msgCmd.On("GetName").Return("msg")
	msgCmd.On("Execute", mock.Anything, mock.Anything, "Hello world").Return(nil)
	cm.commands = []Command{mockCmd, msgCmd}

	window := &Window{}
	connections := &ServerManager{}

	cm.Execute(connections, window, "Hello world")

	mockCmd.AssertExpectations(t)
}

func TestCommandManager_SetNotificationManager(t *testing.T) {
	cm := NewCommandManager(getCommandManagerTestConfig(), make(chan bool, 1))
	nm := DesktopNotificationManager{}

	cm.SetNotificationManager(nm)

	assert.Equal(t, nm, cm.nm, "NotificationManager should be set")
}

func TestCommandManager_showNotification(t *testing.T) {
	cm := NewCommandManager(getCommandManagerTestConfig(), make(chan bool, 1))

	nm := &DesktopNotificationManager{
		pendingNotifications: make(chan Notification, 1),
	}
	cm.SetNotificationManager(nm)

	notification := Notification{
		Title: "Test Title",
		Text:  "Test Text",
		Sound: true,
		Popup: true,
	}

	cm.showNotification(notification)

	select {
	case received := <-nm.pendingNotifications:
		assert.Equal(t, notification, received, "Notification should be sent")
	default:
		t.Error("No notification was sent")
	}
}

func TestCommandManager_showError(t *testing.T) {
	cm := NewCommandManager(getCommandManagerTestConfig(), make(chan bool, 1))

	window := createTestWindow()

	cm.showError(window, "Test error message")

	messages := window.GetMessages()
	assert.Len(t, messages, 1, "Should have added one error message")
	assert.Contains(t, messages[0].GetMessage(), "Test error message", "Error message should be correct")
}

func getCommandManagerTestConfig() *config.Config {
	return &config.Config{
		UISettings: config.UISettings{
			TimestampFormat: "15:04:05",
		},
	}
}
