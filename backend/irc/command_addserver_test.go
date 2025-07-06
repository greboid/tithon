package irc

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Mock UpdateTrigger for testing
type mockUpdateTrigger struct{}

func (m *mockUpdateTrigger) SetPendingUpdate() {
	// Do nothing for testing
}

func TestAddServer_Execute(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantError     bool
		errorContains string
	}{
		{
			name:      "basic hostname",
			input:     "irc.example.com",
			wantError: false,
		},
		{
			name:      "hostname with port",
			input:     "irc.example.com:6667",
			wantError: false,
		},
		{
			name:      "hostname with nickname",
			input:     "irc.example.com testuser",
			wantError: false,
		},
		{
			name:      "hostname with flags",
			input:     "irc.example.com --notls --password=secret",
			wantError: false,
		},
		{
			name:      "hostname with sasl",
			input:     "irc.example.com --sasl=user:pass",
			wantError: false,
		},
		{
			name:      "complex example",
			input:     "irc.example.com:6697 mynick --password=serverpass --sasl=user:pass",
			wantError: false,
		},
		{
			name:          "missing hostname",
			input:         "",
			wantError:     true,
			errorContains: "required argument hostname is missing",
		},
		{
			name:          "invalid nickname",
			input:         "irc.example.com invalid#nick",
			wantError:     true,
			errorContains: "invalid nickname format",
		},
		{
			name:          "invalid sasl format",
			input:         "irc.example.com --sasl=invalidformat",
			wantError:     true,
			errorContains: "SASL credentials must be in format username:password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &AddServer{}
			sm := NewServerManager("15:04:05", nil)
			sm.SetUpdateTrigger(&mockUpdateTrigger{})

			err := cmd.Execute(sm, nil, tt.input)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSplitSASLCredentials(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "normal credentials",
			input: "user:pass",
			want:  []string{"user", "pass"},
		},
		{
			name:  "password with colon",
			input: "user:pass:word",
			want:  []string{"user", "pass:word"},
		},
		{
			name:  "empty password",
			input: "user:",
			want:  []string{"user", ""},
		},
		{
			name:  "no colon",
			input: "userpass",
			want:  []string{"userpass"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitSASLCredentials(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}
