package services

import (
	"testing"

	"github.com/greboid/tithon/config"
	"github.com/stretchr/testify/assert"
)

func TestNewSettingsService(t *testing.T) {
	mockConfig := createMockConfig()

	service := NewSettingsService(mockConfig)

	assert.NotNil(t, service)
	assert.NotNil(t, service.conf)
	assert.NotNil(t, service.settingsData)

	data := service.GetSettingsData()
	assert.Equal(t, mockConfig.UISettings.TimestampFormat, data.TimestampFormat)
	assert.Equal(t, mockConfig.UISettings.ShowNicklist, data.ShowNicklist)
	assert.Equal(t, mockConfig.UISettings.Theme, data.Theme)
	assert.Equal(t, len(mockConfig.Servers), len(data.Servers))
	assert.Equal(t, len(mockConfig.Notifications.Triggers), len(data.Notifications))
	assert.NotEmpty(t, data.Version)
}

func TestSettingsService_GetFromConfig(t *testing.T) {
	mockConfig := createMockConfig()
	service := NewSettingsService(mockConfig)
	originalData := service.GetSettingsData()

	mockConfig.UISettings.TimestampFormat = "15:04"
	mockConfig.UISettings.ShowNicklist = false
	mockConfig.UISettings.Theme = "dark"
	newServer := config.Server{
		Hostname: "irc.example.com",
		Port:     6667,
		TLS:      false,
		Profile: config.Profile{
			Nickname: "testnick",
		},
		ID:          "test-server",
		AutoConnect: true,
	}
	mockConfig.Servers = append(mockConfig.Servers, newServer)
	newTrigger := config.NotificationTrigger{
		Network: "freenode",
		Source:  "#test",
		Nick:    "testnick",
		Message: "hello",
		Sound:   true,
		Popup:   false,
	}
	mockConfig.Notifications.Triggers = append(mockConfig.Notifications.Triggers, newTrigger)

	updatedData := service.GetFromConfig()

	assert.NotSame(t, originalData, updatedData)
	assert.Equal(t, "15:04", updatedData.TimestampFormat)
	assert.False(t, updatedData.ShowNicklist)
	assert.Equal(t, "dark", updatedData.Theme)
	assert.Equal(t, 2, len(updatedData.Servers))
	assert.Equal(t, 2, len(updatedData.Notifications))
	assert.NotEmpty(t, updatedData.Version)

	assert.Equal(t, newServer.Hostname, updatedData.Servers[1].Hostname)
	assert.Equal(t, newServer.Port, updatedData.Servers[1].Port)
	assert.Equal(t, newServer.Profile.Nickname, updatedData.Servers[1].Profile.Nickname)

	assert.Equal(t, newTrigger.Network, updatedData.Notifications[1].Network)
	assert.Equal(t, newTrigger.Source, updatedData.Notifications[1].Source)
	assert.Equal(t, newTrigger.Nick, updatedData.Notifications[1].Nick)

	currentData := service.GetSettingsData()
	assert.Same(t, updatedData, currentData)
}

func TestSettingsService_GetSettingsData(t *testing.T) {
	mockConfig := createMockConfig()
	service := NewSettingsService(mockConfig)

	data := service.GetSettingsData()

	assert.NotNil(t, data)
	assert.Equal(t, mockConfig.UISettings.TimestampFormat, data.TimestampFormat)
	assert.Equal(t, mockConfig.UISettings.ShowNicklist, data.ShowNicklist)
	assert.Equal(t, mockConfig.UISettings.Theme, data.Theme)

	data2 := service.GetSettingsData()
	assert.Same(t, data, data2)
}

func TestSettingsService_SaveSettingsToConfig(t *testing.T) {
	provider := &MockProvider{}
	mockConfig := config.NewConfig(provider)

	mockConfig.UISettings = config.UISettings{
		TimestampFormat: "15:04:05",
		ShowNicklist:    true,
		Theme:           "light",
	}
	mockConfig.Servers = []config.Server{
		{
			Hostname: "irc.test.com",
			Port:     6667,
			TLS:      false,
			Profile: config.Profile{
				Nickname: "testnick",
			},
			ID:          "test-server",
			AutoConnect: false,
		},
	}
	mockConfig.Notifications = config.Notifications{
		Triggers: []config.NotificationTrigger{
			{
				Network: "testnet",
				Source:  "#test",
				Nick:    "testnick",
				Message: "test",
				Sound:   false,
				Popup:   true,
			},
		},
	}

	service := NewSettingsService(mockConfig)

	settingsData := service.GetSettingsData()
	settingsData.TimestampFormat = "15:04"
	settingsData.ShowNicklist = false
	settingsData.Theme = "dark"
	newServer := config.Server{
		Hostname: "irc.example.com",
		Port:     6697,
		TLS:      true,
		Profile: config.Profile{
			Nickname: "newnick",
		},
		ID:          "new-server",
		AutoConnect: true,
	}
	settingsData.Servers = append(settingsData.Servers, newServer)
	newTrigger := config.NotificationTrigger{
		Network: "newnet",
		Source:  "#new",
		Nick:    "newnick",
		Message: "new",
		Sound:   true,
		Popup:   false,
	}
	settingsData.Notifications = append(settingsData.Notifications, newTrigger)

	err := service.SaveSettingsToConfig()
	assert.NoError(t, err)

	assert.Equal(t, "15:04", mockConfig.UISettings.TimestampFormat)
	assert.False(t, mockConfig.UISettings.ShowNicklist)
	assert.Equal(t, "dark", mockConfig.UISettings.Theme)
	assert.Equal(t, 2, len(mockConfig.Servers))
	assert.Equal(t, "irc.test.com", mockConfig.Servers[0].Hostname)
	assert.Equal(t, "irc.example.com", mockConfig.Servers[1].Hostname)
	assert.Equal(t, 6697, mockConfig.Servers[1].Port)
	assert.True(t, mockConfig.Servers[1].TLS)
	assert.Equal(t, "newnick", mockConfig.Servers[1].Profile.Nickname)
	assert.Equal(t, 2, len(mockConfig.Notifications.Triggers))
	assert.Equal(t, "testnet", mockConfig.Notifications.Triggers[0].Network)
	assert.Equal(t, "newnet", mockConfig.Notifications.Triggers[1].Network)
	assert.Equal(t, "new", mockConfig.Notifications.Triggers[1].Message)
	assert.True(t, mockConfig.Notifications.Triggers[1].Sound)
	assert.False(t, mockConfig.Notifications.Triggers[1].Popup)

	assert.True(t, provider.saveCalled)
}

func TestSettingsService_DataIntegrity(t *testing.T) {
	mockConfig := createMockConfig()
	service := NewSettingsService(mockConfig)

	initialData := service.GetSettingsData()
	initialData.Servers[0].Hostname = "modified.hostname.com"
	initialData.Servers[0].Port = 9999

	assert.Equal(t, "modified.hostname.com", mockConfig.Servers[0].Hostname)
	assert.Equal(t, 9999, mockConfig.Servers[0].Port)

	freshData := service.GetFromConfig()

	assert.NotSame(t, initialData, freshData)
	assert.Equal(t, "modified.hostname.com", freshData.Servers[0].Hostname)
	assert.Equal(t, 9999, freshData.Servers[0].Port)

	freshData.Servers[0].Hostname = "second.modification.com"
	freshData.Servers[0].Port = 1234

	assert.Equal(t, "modified.hostname.com", mockConfig.Servers[0].Hostname)
	assert.Equal(t, 9999, mockConfig.Servers[0].Port)

	currentData := service.GetSettingsData()
	assert.Same(t, freshData, currentData)
	assert.Equal(t, "second.modification.com", currentData.Servers[0].Hostname)
	assert.Equal(t, 1234, currentData.Servers[0].Port)
}

func TestSettingsService_GetFromConfig_ProperCopying(t *testing.T) {
	mockConfig := createMockConfig()
	service := NewSettingsService(mockConfig)

	data := service.GetFromConfig()

	mockConfig.Servers[0].Hostname = "config.modified.com"
	mockConfig.Notifications.Triggers[0].Network = "modified-network"

	assert.Equal(t, "irc.libera.chat", data.Servers[0].Hostname)
	assert.Equal(t, "libera", data.Notifications[0].Network)

	data.Servers[0].Hostname = "data.modified.com"
	data.Notifications[0].Network = "data-network"

	assert.Equal(t, "config.modified.com", mockConfig.Servers[0].Hostname)
	assert.Equal(t, "modified-network", mockConfig.Notifications.Triggers[0].Network)
}

type MockProvider struct {
	saveCalled bool
	saveError  error
}

func (m *MockProvider) Load(_ interface{}) error {
	return nil
}

func (m *MockProvider) Save(_ interface{}) error {
	m.saveCalled = true
	return m.saveError
}

func createMockConfig() *config.Config {
	provider := &MockProvider{}
	mockConfig := config.NewConfig(provider)

	mockConfig.UISettings = config.UISettings{
		TimestampFormat: "15:04:05",
		ShowNicklist:    true,
		Theme:           "light",
	}

	mockConfig.Servers = []config.Server{
		{
			Hostname: "irc.libera.chat",
			Port:     6697,
			TLS:      true,
			Profile: config.Profile{
				Nickname: "testnick",
			},
			ID:          "libera",
			AutoConnect: true,
		},
	}

	mockConfig.Notifications = config.Notifications{
		Triggers: []config.NotificationTrigger{
			{
				Network: "libera",
				Source:  "#test",
				Nick:    "testnick",
				Message: "hello",
				Sound:   true,
				Popup:   false,
			},
		},
	}

	return mockConfig
}
