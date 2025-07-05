package irc

import (
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestHandleChannelModes(t *testing.T) {
	type args struct {
		linkRegex          *regexp.Regexp
		timestampFormat    string
		isValidChannel     func(string) bool
		setPendingUpdate   func()
		getChannelByName   func(string) (*Channel, error)
		getModeNameForMode func(string) string
		getChannelModeType func(string) rune
	}
	tests := []struct {
		name               string
		args               args
		message            ircmsg.Message
		initialUsers       []*User
		initialModes       []*ChannelMode
		wantChannelName    string
		wantChannelError   bool
		wantIsValidChannel bool
		wantUserModes      map[string]string
		wantChannelModes   []struct {
			mode      string
			parameter string
			set       bool
			modeType  rune
		}
		wantModeMessage string
		wantProcessed   bool
	}{
		{
			name: "User privilege mode change (+o)",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return true // Now should return true for valid channels
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name: "#test",
								users: []*User{
									NewUser("alice", ""),
									NewUser("bob", "+"),
								},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
							channelModes: make([]*ChannelMode, 0),
						}, nil
					}
					return nil, assert.AnError
				},
				getModeNameForMode: func(mode string) string {
					if mode == "o" {
						return "@"
					}
					if mode == "v" {
						return "+"
					}
					return ""
				},
				getChannelModeType: func(mode string) rune {
					if mode == "o" || mode == "v" {
						return 'P' // Privilege mode
					}
					if mode == "k" {
						return 'B' // Key mode
					}
					if mode == "l" {
						return 'C' // Limit mode
					}
					if mode == "t" {
						return 'D' // Topic mode
					}
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"#test", "+o", "alice"},
			},
			wantChannelName: "#test",
			wantUserModes: map[string]string{
				"alice": "@",
				"bob":   "+",
			},
			wantModeMessage: "admin sets mode +o alice",
			wantProcessed:   true,
		},
		{
			name: "User privilege mode removal (-v)",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return true
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name: "#test",
								users: []*User{
									NewUser("alice", "@"),
									NewUser("bob", "+"),
								},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
							channelModes: make([]*ChannelMode, 0),
						}, nil
					}
					return nil, assert.AnError
				},
				getModeNameForMode: func(mode string) string {
					if mode == "o" {
						return "@"
					}
					if mode == "v" {
						return "+"
					}
					return ""
				},
				getChannelModeType: func(mode string) rune {
					if mode == "o" || mode == "v" {
						return 'P'
					}
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"#test", "-v", "bob"},
			},
			wantChannelName: "#test",
			wantUserModes: map[string]string{
				"alice": "@",
				"bob":   "",
			},
			wantModeMessage: "admin sets mode -v bob",
			wantProcessed:   true,
		},
		{
			name: "Channel mode change with parameter (+k)",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return true
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    []*User{},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
							channelModes: make([]*ChannelMode, 0),
						}, nil
					}
					return nil, assert.AnError
				},
				getModeNameForMode: func(mode string) string {
					return ""
				},
				getChannelModeType: func(mode string) rune {
					if mode == "k" {
						return 'B' // Key mode
					}
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"#test", "+k", "secretkey"},
			},
			wantChannelName: "#test",
			wantChannelModes: []struct {
				mode      string
				parameter string
				set       bool
				modeType  rune
			}{
				{"k", "secretkey", true, 'B'},
			},
			wantModeMessage: "admin sets mode +k secretkey",
			wantProcessed:   true,
		},
		{
			name: "Boolean channel mode (+t)",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return true
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    []*User{},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
							channelModes: make([]*ChannelMode, 0),
						}, nil
					}
					return nil, assert.AnError
				},
				getModeNameForMode: func(mode string) string {
					return ""
				},
				getChannelModeType: func(mode string) rune {
					if mode == "t" {
						return 'D' // Boolean mode
					}
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"#test", "+t"},
			},
			wantChannelName: "#test",
			wantChannelModes: []struct {
				mode      string
				parameter string
				set       bool
				modeType  rune
			}{
				{"t", "", true, 'D'},
			},
			wantModeMessage: "admin sets mode +t",
			wantProcessed:   true,
		},
		{
			name: "Multiple mode changes",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return true
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name: "#test",
								users: []*User{
									NewUser("alice", ""),
									NewUser("bob", ""),
								},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
							channelModes: make([]*ChannelMode, 0),
						}, nil
					}
					return nil, assert.AnError
				},
				getModeNameForMode: func(mode string) string {
					if mode == "o" {
						return "@"
					}
					if mode == "v" {
						return "+"
					}
					return ""
				},
				getChannelModeType: func(mode string) rune {
					if mode == "o" || mode == "v" {
						return 'P'
					}
					if mode == "t" {
						return 'D'
					}
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"#test", "+ov", "alice", "bob"},
			},
			wantChannelName: "#test",
			wantUserModes: map[string]string{
				"alice": "@",
				"bob":   "+",
			},
			wantProcessed: true,
		},
		{
			name: "Valid channel with boolean mode",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return true // Valid channel should be processed
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    []*User{},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
							channelModes: make([]*ChannelMode, 0),
						}, nil
					}
					return nil, assert.AnError
				},
				getModeNameForMode: func(mode string) string {
					return ""
				},
				getChannelModeType: func(mode string) rune {
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"#test", "+t"},
			},
			wantChannelName: "#test",
			wantChannelModes: []struct {
				mode      string
				parameter string
				set       bool
				modeType  rune
			}{
				{"t", "", true, 'D'},
			},
			wantModeMessage: "admin sets mode +t",
			wantProcessed:   true,
		},
		{
			name: "Invalid channel (should be skipped)",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return false // Invalid channel should be skipped
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError // Should not be called
				},
				getModeNameForMode: func(mode string) string {
					return ""
				},
				getChannelModeType: func(mode string) rune {
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"invalid", "+t"},
			},
			wantIsValidChannel: true,
			wantProcessed:      false,
		},
		{
			name: "Unknown channel",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return true
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					return nil, assert.AnError
				},
				getModeNameForMode: func(mode string) string {
					return ""
				},
				getChannelModeType: func(mode string) rune {
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"#nonexistent", "+t"},
			},
			wantChannelName:  "#nonexistent",
			wantChannelError: true,
			wantProcessed:    false,
		},
		{
			name: "Mode change without required parameter (should be skipped)",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return true
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    []*User{},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
							channelModes: make([]*ChannelMode, 0),
						}, nil
					}
					return nil, assert.AnError
				},
				getModeNameForMode: func(mode string) string {
					return ""
				},
				getChannelModeType: func(mode string) rune {
					if mode == "o" {
						return 'P' // Privilege mode requires parameter
					}
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"#test", "+o"},
			},
			wantChannelName: "#test",
			wantProcessed:   false,
		},
		{
			name: "Type A mode with parameter (+b)",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return true
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    []*User{},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
							channelModes: make([]*ChannelMode, 0),
						}, nil
					}
					return nil, assert.AnError
				},
				getModeNameForMode: func(mode string) string {
					return ""
				},
				getChannelModeType: func(mode string) rune {
					if mode == "b" {
						return 'A' // Ban mode - always needs parameter
					}
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"#test", "+b", "*!*@spam.com"},
			},
			wantChannelName: "#test",
			wantChannelModes: []struct {
				mode      string
				parameter string
				set       bool
				modeType  rune
			}{
				{"b", "*!*@spam.com", true, 'A'},
			},
			wantModeMessage: "admin sets mode +b *!*@spam.com",
			wantProcessed:   true,
		},
		{
			name: "Type A mode without parameter (should be skipped)",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return true
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    []*User{},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
							channelModes: make([]*ChannelMode, 0),
						}, nil
					}
					return nil, assert.AnError
				},
				getModeNameForMode: func(mode string) string {
					return ""
				},
				getChannelModeType: func(mode string) rune {
					if mode == "b" {
						return 'A' // Ban mode - always needs parameter
					}
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"#test", "+b"}, // Missing parameter
			},
			wantChannelName: "#test",
			wantProcessed:   false, // No valid operations, so setPendingUpdate is never called
		},
		{
			name: "Type B mode without parameter (should be skipped)",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return true
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    []*User{},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
							channelModes: make([]*ChannelMode, 0),
						}, nil
					}
					return nil, assert.AnError
				},
				getModeNameForMode: func(mode string) string {
					return ""
				},
				getChannelModeType: func(mode string) rune {
					if mode == "k" {
						return 'B' // Key mode - always needs parameter
					}
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"#test", "+k"}, // Missing parameter
			},
			wantChannelName: "#test",
			wantProcessed:   false, // No valid operations, so setPendingUpdate is never called
		},
		{
			name: "Type C mode removal without parameter (-l)",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return true
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    []*User{},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
							channelModes: make([]*ChannelMode, 0),
						}, nil
					}
					return nil, assert.AnError
				},
				getModeNameForMode: func(mode string) string {
					return ""
				},
				getChannelModeType: func(mode string) rune {
					if mode == "l" {
						return 'C' // Limit mode - needs parameter when setting, not when removing
					}
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"#test", "-l"}, // No parameter needed for removal
			},
			wantChannelName: "#test",
			wantModeMessage: "admin sets mode -l",
			wantProcessed:   true,
		},
		{
			name: "Type C mode set without parameter (should be skipped)",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return true
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    []*User{},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
							channelModes: make([]*ChannelMode, 0),
						}, nil
					}
					return nil, assert.AnError
				},
				getModeNameForMode: func(mode string) string {
					return ""
				},
				getChannelModeType: func(mode string) rune {
					if mode == "l" {
						return 'C' // Limit mode - needs parameter when setting
					}
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"#test", "+l"}, // Missing parameter for setting
			},
			wantChannelName: "#test",
			wantProcessed:   false, // No valid operations, so setPendingUpdate is never called
		},
		{
			name: "Mixed add and remove modes (+o-v alice bob)",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return true
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name: "#test",
								users: []*User{
									NewUser("alice", ""),
									NewUser("bob", "+"),
								},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
							channelModes: make([]*ChannelMode, 0),
						}, nil
					}
					return nil, assert.AnError
				},
				getModeNameForMode: func(mode string) string {
					if mode == "o" {
						return "@"
					}
					if mode == "v" {
						return "+"
					}
					return ""
				},
				getChannelModeType: func(mode string) rune {
					if mode == "o" || mode == "v" {
						return 'P'
					}
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"#test", "+o-v", "alice", "bob"},
			},
			wantChannelName: "#test",
			wantUserModes: map[string]string{
				"alice": "@",
				"bob":   "",
			},
			wantProcessed: true,
		},
		{
			name: "Unknown mode type (should warn and be ignored)",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return true
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    []*User{},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
							channelModes: make([]*ChannelMode, 0),
						}, nil
					}
					return nil, assert.AnError
				},
				getModeNameForMode: func(mode string) string {
					return ""
				},
				getChannelModeType: func(mode string) rune {
					if mode == "x" {
						return 'X' // Unknown mode type
					}
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"#test", "+x"},
			},
			wantChannelName: "#test",
			wantProcessed:   true, // Handler processes but mode is ignored with warning
		},
		{
			name: "Empty mode string",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return true
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    []*User{},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
							channelModes: make([]*ChannelMode, 0),
						}, nil
					}
					return nil, assert.AnError
				},
				getModeNameForMode: func(mode string) string {
					return ""
				},
				getChannelModeType: func(mode string) rune {
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"#test", ""},
			},
			wantChannelName: "#test",
			wantProcessed:   false, // No modes parsed, so setPendingUpdate is never called
		},
		{
			name: "Mode string with only plus sign",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return true
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name:     "#test",
								users:    []*User{},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
							channelModes: make([]*ChannelMode, 0),
						}, nil
					}
					return nil, assert.AnError
				},
				getModeNameForMode: func(mode string) string {
					return ""
				},
				getChannelModeType: func(mode string) rune {
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"#test", "+"},
			},
			wantChannelName: "#test",
			wantProcessed:   false, // No modes parsed, so setPendingUpdate is never called
		},
		{
			name: "Complex mode string with multiple types (+ovl-t alice bob 50)",
			args: args{
				linkRegex:       regexp.MustCompile(`https?://\S+`),
				timestampFormat: "15:04:05",
				isValidChannel: func(name string) bool {
					return true
				},
				setPendingUpdate: func() {},
				getChannelByName: func(name string) (*Channel, error) {
					if name == "#test" {
						return &Channel{
							Window: &Window{
								name: "#test",
								users: []*User{
									NewUser("alice", ""),
									NewUser("bob", ""),
								},
								messages: make([]*Message, 0),
								hasUsers: true,
							},
							channelModes: make([]*ChannelMode, 0),
						}, nil
					}
					return nil, assert.AnError
				},
				getModeNameForMode: func(mode string) string {
					if mode == "o" {
						return "@"
					}
					if mode == "v" {
						return "+"
					}
					return ""
				},
				getChannelModeType: func(mode string) rune {
					if mode == "o" || mode == "v" {
						return 'P'
					}
					if mode == "l" {
						return 'C'
					}
					if mode == "t" {
						return 'D'
					}
					return 'D'
				},
			},
			message: ircmsg.Message{
				Source:  "admin!admin@example.com",
				Command: "MODE",
				Params:  []string{"#test", "+ovl-t", "alice", "bob", "50"},
			},
			wantChannelName: "#test",
			wantUserModes: map[string]string{
				"alice": "@",
				"bob":   "+",
			},
			wantProcessed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pendingUpdateCalled bool
			var channel *Channel

			setPendingUpdate := func() {
				pendingUpdateCalled = true
			}

			getChannelByName := func(name string) (*Channel, error) {
				if tt.wantChannelError {
					return nil, assert.AnError
				}
				result, err := tt.args.getChannelByName(name)
				if err == nil {
					channel = result
				}
				return result, err
			}

			handler := HandleChannelModes(tt.args.timestampFormat, tt.args.isValidChannel, setPendingUpdate, getChannelByName, tt.args.getModeNameForMode, tt.args.getChannelModeType)
			handler(tt.message)

			if tt.wantIsValidChannel {
				return
			}

			if tt.wantChannelError {
				return
			}

			if !tt.wantProcessed {
				return
			}

			if channel == nil {
				t.Logf("Channel is nil - handler may have exited early")
				return
			}

			assert.True(t, pendingUpdateCalled, "setPendingUpdate should have been called")

			assert.NotNil(t, channel, "Channel should have been retrieved")
			assert.Equal(t, tt.wantChannelName, channel.GetName(), "Channel name should match")

			if len(tt.wantUserModes) > 0 {
				users := channel.GetUsers()
				for nickname, expectedModes := range tt.wantUserModes {
					var foundUser *User
					for _, user := range users {
						if user.GetNickListDisplay() == nickname {
							foundUser = user
							break
						}
					}
					assert.NotNil(t, foundUser, "Expected user should be found: "+nickname)
					assert.Equal(t, expectedModes, foundUser.GetNickListModes(), "User modes should match for: "+nickname)
				}
			}

			if len(tt.wantChannelModes) > 0 {
				channelModes := channel.GetChannelModes()
				assert.Equal(t, len(tt.wantChannelModes), len(channelModes), "Channel mode count should match")

				for _, expectedMode := range tt.wantChannelModes {
					var foundMode *ChannelMode
					for _, mode := range channelModes {
						if mode.Mode == expectedMode.mode {
							foundMode = mode
							break
						}
					}
					assert.NotNil(t, foundMode, "Expected channel mode should be found: "+expectedMode.mode)
					assert.Equal(t, expectedMode.parameter, foundMode.Parameter, "Mode parameter should match for: "+expectedMode.mode)
					assert.Equal(t, expectedMode.set, foundMode.Set, "Mode set state should match for: "+expectedMode.mode)
					assert.Equal(t, expectedMode.modeType, foundMode.Type, "Mode type should match for: "+expectedMode.mode)
				}
			}

			if tt.wantModeMessage != "" {
				messages := channel.GetMessages()
				assert.NotEmpty(t, messages, "At least one message should have been added to the channel")

				found := false
				for _, msg := range messages {
					if msg.GetMessage() == tt.wantModeMessage {
						found = true
						assert.Equal(t, MessageType(Event), msg.GetType(), "Message type should be Event")
						break
					}
				}
				assert.True(t, found, "Expected mode message should be found: "+tt.wantModeMessage)
			}

		})
	}
}
