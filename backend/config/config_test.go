package config

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockProvider struct {
	loadCalled bool
	saveCalled bool
	loadError  error
	saveError  error
	loadData   *Config
}

func (m *MockProvider) Load(target interface{}) error {
	m.loadCalled = true
	if m.loadError != nil {
		return m.loadError
	}
	if m.loadData != nil {
		if config, ok := target.(*Config); ok {
			config.Servers = m.loadData.Servers
			config.UISettings = m.loadData.UISettings
			config.Notifications = m.loadData.Notifications
		}
	}
	return nil
}

func (m *MockProvider) Save(_ interface{}) error {
	m.saveCalled = true
	return m.saveError
}

func TestNewConfig(t *testing.T) {
	provider := &MockProvider{}
	config := NewConfig(provider)

	require.NotNil(t, config, "NewConfig returned nil")
	assert.Equal(t, provider, config.instance, "NewConfig did not set the provider correctly")
}

func TestConfig_Load_BlankID(t *testing.T) {
	c := NewConfig(&MockProvider{
		loadData: &Config{
			Servers: []Server{
				{
					ID: "",
				},
			},
		},
	})
	err := c.Load()
	assert.NoError(t, err, "Unexpected error")
	assert.NotNil(t, c, "Config is nil")
	assert.NotEmptyf(t, c.Servers[0].ID, "Server ID is blank")
}

func TestConfig_Load(t *testing.T) {
	tests := []struct {
		name           string
		provider       *MockProvider
		wantErr        bool
		expectedErr    error
		expectedConfig *Config
	}{
		{
			name: "Successful load",
			provider: &MockProvider{
				loadData: &Config{
					Servers: []Server{
						{
							ID:       "test-id",
							Hostname: "irc.example.com",
							Port:     6667,
							TLS:      true,
						},
					},
					UISettings: UISettings{
						TimestampFormat: "15:04:05",
					},
				},
			},
			wantErr: false,
			expectedConfig: &Config{
				Servers: []Server{
					{
						ID:       "test-id",
						Hostname: "irc.example.com",
						Port:     6667,
						TLS:      true,
					},
				},
				UISettings: UISettings{
					TimestampFormat: "15:04:05",
				},
			},
		},
		{
			name: "Load error",
			provider: &MockProvider{
				loadError: errors.New("load error"),
			},
			wantErr:     true,
			expectedErr: errors.New("load error"),
		},
		{
			name: "Empty timestamp format",
			provider: &MockProvider{
				loadData: &Config{
					UISettings: UISettings{
						TimestampFormat: "",
					},
				},
			},
			wantErr: false,
			expectedConfig: &Config{
				UISettings: UISettings{
					TimestampFormat: time.TimeOnly,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfig(tt.provider)
			err := c.Load()

			assert.True(t, tt.provider.loadCalled, "Provider's Load method was not called")

			if tt.wantErr {
				assert.Error(t, err, "Expected an error but got none")
				assert.Equal(t, tt.expectedErr.Error(), err.Error(), "Error messages don't match")
				return
			}

			assert.NoError(t, err, "Unexpected error")
			assert.NotNil(t, c, "Config is nil")
			assert.Equal(t, tt.expectedConfig.Servers, c.Servers, "Servers don't match")
			assert.Equal(t, tt.expectedConfig.UISettings, c.UISettings, "UISettings don't match")
			assert.Equal(t, tt.expectedConfig.Notifications, c.Notifications, "Notifications don't match")
		})
	}
}

func TestConfig_Save(t *testing.T) {
	tests := []struct {
		name        string
		provider    *MockProvider
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "Successful save",
			provider: &MockProvider{},
			wantErr:  false,
		},
		{
			name: "Save error",
			provider: &MockProvider{
				saveError: errors.New("save error"),
			},
			wantErr:     true,
			expectedErr: errors.New("save error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewConfig(tt.provider)
			err := c.Save()

			assert.True(t, tt.provider.saveCalled, "Provider's Save method was not called")

			if tt.wantErr {
				assert.Error(t, err, "Expected an error but got none")
				assert.Equal(t, tt.expectedErr.Error(), err.Error(), "Error messages don't match")
				return
			}

			assert.NoError(t, err, "Unexpected error")
		})
	}
}
