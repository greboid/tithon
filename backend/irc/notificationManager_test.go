package irc

import (
	"fmt"
	"github.com/greboid/tithon/config"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationManager_sortTriggers(t *testing.T) {
	tests := []struct {
		name     string
		triggers []config.NotificationTrigger
		want     []config.NotificationTrigger
	}{
		{
			name: "Sort by specificity",
			triggers: []config.NotificationTrigger{
				{
					Network: ".*",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   false,
					Popup:   false,
				},
				{
					Network: "Network",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   false,
					Popup:   false,
				},
			},
			want: []config.NotificationTrigger{
				{
					Network: "Network",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   false,
					Popup:   false,
				},
				{
					Network: ".*",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   false,
					Popup:   false,
				},
			},
		},
		{
			name: "Sort by sound when specificity is equal",
			triggers: []config.NotificationTrigger{
				{
					Network: ".*",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   false,
					Popup:   false,
				},
				{
					Network: ".*",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   true,
					Popup:   false,
				},
			},
			want: []config.NotificationTrigger{
				{
					Network: ".*",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   true,
					Popup:   false,
				},
				{
					Network: ".*",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   false,
					Popup:   false,
				},
			},
		},
		{
			name: "Sort by popup when specificity and sound are equal",
			triggers: []config.NotificationTrigger{
				{
					Network: ".*",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   false,
					Popup:   false,
				},
				{
					Network: ".*",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   false,
					Popup:   true,
				},
			},
			want: []config.NotificationTrigger{
				{
					Network: ".*",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   false,
					Popup:   true,
				},
				{
					Network: ".*",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   false,
					Popup:   false,
				},
			},
		},
		{
			name: "more complicated sort with different specificities, sound and popup",
			triggers: []config.NotificationTrigger{
				{
					Network: ".*",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   false,
					Popup:   false,
				},
				{
					Network: "Network",
					Source:  "#channel",
					Nick:    "user",
					Message: "hello",
					Sound:   false,
					Popup:   false,
				},
				{
					Network: "Network",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   true,
					Popup:   false,
				},
				{
					Network: ".*",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   false,
					Popup:   true,
				},
			},
			want: []config.NotificationTrigger{
				{
					Network: "Network",
					Source:  "#channel",
					Nick:    "user",
					Message: "hello",
					Sound:   false,
					Popup:   false,
				},
				{
					Network: "Network",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   true,
					Popup:   false,
				},
				{
					Network: ".*",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   false,
					Popup:   true,
				},
				{
					Network: ".*",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   false,
					Popup:   false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SortNotificationTriggers(tt.triggers)
			assert.Equal(t, tt.want, got, "sortTriggers() returned unexpected result")
		})
	}
}

func TestNotificationManager_compileRegex(t *testing.T) {
	tests := []struct {
		name    string
		regex   string
		want    string
		wantErr bool
	}{
		{
			name:    "Empty string",
			regex:   "",
			want:    ".*",
			wantErr: false,
		},
		{
			name:    "Valid regex",
			regex:   "test.*",
			want:    "test.*",
			wantErr: false,
		},
		{
			name:    "Another valid regex",
			regex:   "[a-z]+",
			want:    "[a-z]+",
			wantErr: false,
		},
		{
			name:    "Invalid regex",
			regex:   "[",
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nm := &DesktopNotificationManager{}
			require.NotNil(t, nm, "NotificationManager should not be nil")
			got, err := CompileNotificationRegex(tt.regex)
			if tt.wantErr {
				assert.Error(t, err, "compileRegex() should return an error")
			} else {
				assert.NoError(t, err, "compileRegex() should not return an error")
				assert.Equal(t, tt.want, got.String(), "compileRegex() returned unexpected result")
			}
		})
	}
}

func TestNotificationManager_AddNotification(t *testing.T) {
	tests := []struct {
		name    string
		network string
		source  string
		nick    string
		message string
		sound   bool
		popup   bool
		want    *Trigger
		wantErr bool
	}{
		{
			name:    "All valid regex patterns",
			network: "network.*",
			source:  "source.*",
			nick:    "nick.*",
			message: "message.*",
			sound:   true,
			popup:   true,
			want: &Trigger{
				Network: regexp.MustCompile("network.*"),
				Source:  regexp.MustCompile("source.*"),
				Nick:    regexp.MustCompile("nick.*"),
				Message: regexp.MustCompile("message.*"),
				Sound:   true,
				Popup:   true,
			},
			wantErr: false,
		},
		{
			name:    "All empty patterns",
			network: "",
			source:  "",
			nick:    "",
			message: "",
			sound:   false,
			popup:   false,
			want: &Trigger{
				Network: regexp.MustCompile(".*"),
				Source:  regexp.MustCompile(".*"),
				Nick:    regexp.MustCompile(".*"),
				Message: regexp.MustCompile(".*"),
				Sound:   false,
				Popup:   false,
			},
			wantErr: false,
		},
		{
			name:    "Invalid network regex",
			network: "[",
			source:  "source.*",
			nick:    "nick.*",
			message: "message.*",
			sound:   true,
			popup:   true,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Invalid source regex",
			network: "network.*",
			source:  "[",
			nick:    "nick.*",
			message: "message.*",
			sound:   true,
			popup:   true,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Invalid nick regex",
			network: "network.*",
			source:  "source.*",
			nick:    "[",
			message: "message.*",
			sound:   true,
			popup:   true,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Invalid message regex",
			network: "network.*",
			source:  "source.*",
			nick:    "nick.*",
			message: "[",
			sound:   true,
			popup:   true,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Sound on, popup off",
			network: "network.*",
			source:  "source.*",
			nick:    "nick.*",
			message: "message.*",
			sound:   true,
			popup:   false,
			want: &Trigger{
				Network: regexp.MustCompile("network.*"),
				Source:  regexp.MustCompile("source.*"),
				Nick:    regexp.MustCompile("nick.*"),
				Message: regexp.MustCompile("message.*"),
				Sound:   true,
				Popup:   false,
			},
			wantErr: false,
		},
		{
			name:    "Sound off, popup on",
			network: "network.*",
			source:  "source.*",
			nick:    "nick.*",
			message: "message.*",
			sound:   false,
			popup:   true,
			want: &Trigger{
				Network: regexp.MustCompile("network.*"),
				Source:  regexp.MustCompile("source.*"),
				Nick:    regexp.MustCompile("nick.*"),
				Message: regexp.MustCompile("message.*"),
				Sound:   false,
				Popup:   true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateNotification(tt.network, tt.source, tt.nick, tt.message, tt.sound, tt.popup, 5*time.Second)

			if tt.wantErr {
				assert.Error(t, err, "AddNotification() should return an error")
				assert.Nil(t, got, "AddNotification() should return nil when there's an error")
				return
			}

			require.NotNil(t, got, "AddNotification() should not return nil")
			assert.NoError(t, err, "AddNotification() should not return an error")

			assert.Equal(t, tt.want.Network.String(), got.Network.String(), "AddNotification() Network mismatch")
			assert.Equal(t, tt.want.Source.String(), got.Source.String(), "AddNotification() Source mismatch")
			assert.Equal(t, tt.want.Nick.String(), got.Nick.String(), "AddNotification() Nick mismatch")
			assert.Equal(t, tt.want.Message.String(), got.Message.String(), "AddNotification() Message mismatch")
			assert.Equal(t, tt.want.Sound, got.Sound, "AddNotification() Sound mismatch")
			assert.Equal(t, tt.want.Popup, got.Popup, "AddNotification() Popup mismatch")
		})
	}
}

func TestNotificationManager_convertFromConfig(t *testing.T) {
	tests := []struct {
		name     string
		triggers []config.NotificationTrigger
		want     []Trigger
	}{
		{
			name: "All valid triggers",
			triggers: []config.NotificationTrigger{
				{
					Network: "network1.*",
					Source:  "source1.*",
					Nick:    "nick1.*",
					Message: "message1.*",
					Sound:   true,
					Popup:   true,
				},
				{
					Network: "network2.*",
					Source:  "source2.*",
					Nick:    "nick2.*",
					Message: "message2.*",
					Sound:   false,
					Popup:   false,
				},
			},
			want: []Trigger{
				{
					Network: regexp.MustCompile("network1.*"),
					Source:  regexp.MustCompile("source1.*"),
					Nick:    regexp.MustCompile("nick1.*"),
					Message: regexp.MustCompile("message1.*"),
					Sound:   true,
					Popup:   true,
				},
				{
					Network: regexp.MustCompile("network2.*"),
					Source:  regexp.MustCompile("source2.*"),
					Nick:    regexp.MustCompile("nick2.*"),
					Message: regexp.MustCompile("message2.*"),
					Sound:   false,
					Popup:   false,
				},
			},
		},
		{
			name: "Some invalid triggers",
			triggers: []config.NotificationTrigger{
				{
					Network: "network1.*",
					Source:  "source1.*",
					Nick:    "nick1.*",
					Message: "message1.*",
					Sound:   true,
					Popup:   true,
				},
				{
					Network: "[",
					Source:  "source2.*",
					Nick:    "nick2.*",
					Message: "message2.*",
					Sound:   false,
					Popup:   false,
				},
				{
					Network: "network3.*",
					Source:  "source3.*",
					Nick:    "nick3.*",
					Message: "message3.*",
					Sound:   true,
					Popup:   false,
				},
			},
			want: []Trigger{
				{
					Network: regexp.MustCompile("network1.*"),
					Source:  regexp.MustCompile("source1.*"),
					Nick:    regexp.MustCompile("nick1.*"),
					Message: regexp.MustCompile("message1.*"),
					Sound:   true,
					Popup:   true,
				},
				{
					Network: regexp.MustCompile("network3.*"),
					Source:  regexp.MustCompile("source3.*"),
					Nick:    regexp.MustCompile("nick3.*"),
					Message: regexp.MustCompile("message3.*"),
					Sound:   true,
					Popup:   false,
				},
			},
		},
		{
			name:     "Empty triggers",
			triggers: []config.NotificationTrigger{},
			want:     []Trigger{},
		},
		{
			name: "Ensure sorting is applied",
			triggers: []config.NotificationTrigger{
				{
					Network: ".*",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   false,
					Popup:   false,
				},
				{
					Network: "network.*",
					Source:  ".*",
					Nick:    ".*",
					Message: ".*",
					Sound:   false,
					Popup:   false,
				},
			},
			want: []Trigger{
				{
					Network: regexp.MustCompile("network.*"),
					Source:  regexp.MustCompile(".*"),
					Nick:    regexp.MustCompile(".*"),
					Message: regexp.MustCompile(".*"),
					Sound:   false,
					Popup:   false,
				},
				{
					Network: regexp.MustCompile(".*"),
					Source:  regexp.MustCompile(".*"),
					Nick:    regexp.MustCompile(".*"),
					Message: regexp.MustCompile(".*"),
					Sound:   false,
					Popup:   false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertNotificationsFromConfig(tt.triggers)

			assert.Equal(t, len(tt.want), len(got), "convertFromConfig() returned unexpected number of triggers")

			for i := range got {
				assert.Equal(t, tt.want[i].Network.String(), got[i].Network.String(), "convertFromConfig() trigger[%d].Network mismatch", i)
				assert.Equal(t, tt.want[i].Source.String(), got[i].Source.String(), "convertFromConfig() trigger[%d].Source mismatch", i)
				assert.Equal(t, tt.want[i].Nick.String(), got[i].Nick.String(), "convertFromConfig() trigger[%d].Nick mismatch", i)
				assert.Equal(t, tt.want[i].Message.String(), got[i].Message.String(), "convertFromConfig() trigger[%d].Message mismatch", i)
				assert.Equal(t, tt.want[i].Sound, got[i].Sound, "convertFromConfig() trigger[%d].Sound mismatch", i)
				assert.Equal(t, tt.want[i].Popup, got[i].Popup, "convertFromConfig() trigger[%d].Popup mismatch", i)
			}
		})
	}
}

func TestNotificationManager_CheckAndNotify(t *testing.T) {
	tests := []struct {
		name               string
		notifications      []Trigger
		network            string
		source             string
		nick               string
		message            string
		expectNotification bool
		expectedTitle      string
		expectedText       string
		expectedSound      bool
		expectedPopup      bool
	}{
		{
			name: "Match all",
			notifications: []Trigger{
				{
					Network: regexp.MustCompile("testnet"),
					Source:  regexp.MustCompile("#testchannel"),
					Nick:    regexp.MustCompile("testnick"),
					Message: regexp.MustCompile("testmessage"),
					Sound:   true,
					Popup:   true,
				},
			},
			network:            "testnet",
			source:             "#testchannel",
			nick:               "testnick",
			message:            "testmessage",
			expectNotification: true,
			expectedTitle:      "testnick (#testchannel)",
			expectedText:       "testmessage",
			expectedSound:      true,
			expectedPopup:      true,
		},
		{
			name: "Match with wildcards",
			notifications: []Trigger{
				{
					Network: regexp.MustCompile(".*"),
					Source:  regexp.MustCompile(".*"),
					Nick:    regexp.MustCompile("user.*"),
					Message: regexp.MustCompile(".*hello.*"),
					Sound:   false,
					Popup:   true,
				},
			},
			network:            "anynetwork",
			source:             "#anychannel",
			nick:               "user123",
			message:            "saying hello world",
			expectNotification: true,
			expectedTitle:      "user123 (#anychannel)",
			expectedText:       "saying hello world",
			expectedSound:      false,
			expectedPopup:      true,
		},
		{
			name: "wrong network",
			notifications: []Trigger{
				{
					Network: regexp.MustCompile("testnet"),
					Source:  regexp.MustCompile("#testchannel"),
					Nick:    regexp.MustCompile("testnick"),
					Message: regexp.MustCompile("testmessage"),
					Sound:   true,
					Popup:   true,
				},
			},
			network:            "wrongnet",
			source:             "#testchannel",
			nick:               "testnick",
			message:            "testmessage",
			expectNotification: false,
			expectedPopup:      false,
			expectedSound:      false,
		},
		{
			name: "wrong source",
			notifications: []Trigger{
				{
					Network: regexp.MustCompile("testnet"),
					Source:  regexp.MustCompile("#testchannel"),
					Nick:    regexp.MustCompile("testnick"),
					Message: regexp.MustCompile("testmessage"),
					Sound:   true,
					Popup:   true,
				},
			},
			network:            "testnet",
			source:             "#wrongchannel",
			nick:               "testnick",
			message:            "testmessage",
			expectNotification: false,
			expectedPopup:      false,
			expectedSound:      false,
		},
		{
			name: "wrong nick",
			notifications: []Trigger{
				{
					Network: regexp.MustCompile("testnet"),
					Source:  regexp.MustCompile("#testchannel"),
					Nick:    regexp.MustCompile("testnick"),
					Message: regexp.MustCompile("testmessage"),
					Sound:   true,
					Popup:   true,
				},
			},
			network:            "testnet",
			source:             "#testchannel",
			nick:               "wrongnick",
			message:            "testmessage",
			expectNotification: false,
			expectedPopup:      false,
			expectedSound:      false,
		},
		{
			name: "No match - wrong message",
			notifications: []Trigger{
				{
					Network: regexp.MustCompile("testnet"),
					Source:  regexp.MustCompile("#testchannel"),
					Nick:    regexp.MustCompile("testnick"),
					Message: regexp.MustCompile("testmessage"),
					Sound:   true,
					Popup:   true,
				},
			},
			network:            "testnet",
			source:             "#testchannel",
			nick:               "testnick",
			message:            "wrongmessage",
			expectNotification: false,
			expectedPopup:      false,
			expectedSound:      false,
		},
		{
			name: "Multiple triggers - first",
			notifications: []Trigger{
				{
					Network: regexp.MustCompile("testnet"),
					Source:  regexp.MustCompile("#testchannel"),
					Nick:    regexp.MustCompile("testnick"),
					Message: regexp.MustCompile("testmessage"),
					Sound:   true,
					Popup:   false,
				},
				{
					Network: regexp.MustCompile("othernet"),
					Source:  regexp.MustCompile("#otherchannel"),
					Nick:    regexp.MustCompile("othernick"),
					Message: regexp.MustCompile("othermessage"),
					Sound:   false,
					Popup:   true,
				},
			},
			network:            "testnet",
			source:             "#testchannel",
			nick:               "testnick",
			message:            "testmessage",
			expectNotification: true,
			expectedTitle:      "testnick (#testchannel)",
			expectedText:       "testmessage",
			expectedSound:      true,
			expectedPopup:      false,
		},
		{
			name: "Multiple triggers - second",
			notifications: []Trigger{
				{
					Network: regexp.MustCompile("testnet"),
					Source:  regexp.MustCompile("#testchannel"),
					Nick:    regexp.MustCompile("testnick"),
					Message: regexp.MustCompile("testmessage"),
					Sound:   true,
					Popup:   false,
				},
				{
					Network: regexp.MustCompile("othernet"),
					Source:  regexp.MustCompile("#otherchannel"),
					Nick:    regexp.MustCompile("othernick"),
					Message: regexp.MustCompile("othermessage"),
					Sound:   false,
					Popup:   true,
				},
			},
			network:            "othernet",
			source:             "#otherchannel",
			nick:               "othernick",
			message:            "othermessage",
			expectNotification: true,
			expectedTitle:      "othernick (#otherchannel)",
			expectedText:       "othermessage",
			expectedSound:      false,
			expectedPopup:      true,
		},
		{
			name:               "No triggers",
			notifications:      []Trigger{},
			network:            "testnet",
			source:             "#testchannel",
			nick:               "testnick",
			message:            "testmessage",
			expectNotification: false,
			expectedPopup:      false,
			expectedSound:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notificationChan := make(chan Notification, 1)
			nm := &DesktopNotificationManager{
				notifications:         tt.notifications,
				pendingNotifications:  notificationChan,
				lastNotificationTimes: make(map[string]time.Time),
			}
			require.NotNil(t, nm, "NotificationManager should not be nil")
			nm.CheckAndNotify(tt.network, tt.source, tt.nick, tt.message)

			var receivedNotification Notification
			select {
			case receivedNotification = <-notificationChan:
				assert.True(t, tt.expectNotification, "CheckAndNotify() sent a notification when none was expected")
				assert.Equal(t, tt.expectedTitle, receivedNotification.Title, "CheckAndNotify() notification title mismatch")
				assert.Equal(t, tt.expectedText, receivedNotification.Text, "CheckAndNotify() notification text mismatch")
				assert.Equal(t, tt.expectedSound, receivedNotification.Sound, "CheckAndNotify() notification sound mismatch")
				assert.Equal(t, tt.expectedPopup, receivedNotification.Popup, "CheckAndNotify() notification popup mismatch")
			default:
				assert.False(t, tt.expectNotification, "CheckAndNotify() did not send a notification when one was expected")
			}
		})
	}
}

func TestNotificationManager_CheckAndNotify_Debouncing(t *testing.T) {
	tests := []struct {
		name             string
		debounceDuration time.Duration
		calls            []struct {
			network            string
			source             string
			nick               string
			message            string
			expectNotification bool
			sleepBefore        time.Duration
		}
	}{
		{
			name:             "First notification sent, second debounced",
			debounceDuration: 100 * time.Millisecond,
			calls: []struct {
				network            string
				source             string
				nick               string
				message            string
				expectNotification bool
				sleepBefore        time.Duration
			}{
				{
					network:            "testnet",
					source:             "#testchannel",
					nick:               "testnick",
					message:            "first message",
					expectNotification: true,
					sleepBefore:        0,
				},
				{
					network:            "testnet",
					source:             "#testchannel",
					nick:               "testnick2",
					message:            "second message",
					expectNotification: false,
					sleepBefore:        50 * time.Millisecond,
				},
			},
		},
		{
			name:             "Both notifications sent after debounce period",
			debounceDuration: 50 * time.Millisecond,
			calls: []struct {
				network            string
				source             string
				nick               string
				message            string
				expectNotification bool
				sleepBefore        time.Duration
			}{
				{
					network:            "testnet",
					source:             "#testchannel",
					nick:               "testnick",
					message:            "first message",
					expectNotification: true,
					sleepBefore:        0,
				},
				{
					network:            "testnet",
					source:             "#testchannel",
					nick:               "testnick2",
					message:            "second message",
					expectNotification: true,
					sleepBefore:        100 * time.Millisecond,
				},
			},
		},
		{
			name:             "Different network-channel pairs don't affect each other",
			debounceDuration: 100 * time.Millisecond,
			calls: []struct {
				network            string
				source             string
				nick               string
				message            string
				expectNotification bool
				sleepBefore        time.Duration
			}{
				{
					network:            "testnet",
					source:             "#testchannel",
					nick:               "testnick",
					message:            "first message",
					expectNotification: true,
					sleepBefore:        0,
				},
				{
					network:            "othernet",
					source:             "#testchannel",
					nick:               "testnick2",
					message:            "second message",
					expectNotification: true,
					sleepBefore:        10 * time.Millisecond,
				},
				{
					network:            "testnet",
					source:             "#otherchannel",
					nick:               "testnick3",
					message:            "third message",
					expectNotification: true,
					sleepBefore:        10 * time.Millisecond,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notificationChan := make(chan Notification, 10)
			trigger := Trigger{
				Network:          regexp.MustCompile(".*"),
				Source:           regexp.MustCompile(".*"),
				Nick:             regexp.MustCompile(".*"),
				Message:          regexp.MustCompile(".*"),
				Sound:            true,
				Popup:            true,
				DebounceDuration: tt.debounceDuration,
			}

			nm := &DesktopNotificationManager{
				notifications:         []Trigger{trigger},
				pendingNotifications:  notificationChan,
				lastNotificationTimes: make(map[string]time.Time),
			}

			for i, call := range tt.calls {
				if call.sleepBefore > 0 {
					time.Sleep(call.sleepBefore)
				}

				nm.CheckAndNotify(call.network, call.source, call.nick, call.message)

				select {
				case notification := <-notificationChan:
					assert.True(t, call.expectNotification, "Call %d: CheckAndNotify() sent notification when none expected", i)
					assert.Equal(t, fmt.Sprintf("%s (%s)", call.nick, call.source), notification.Title, "Call %d: notification title mismatch", i)
					assert.Equal(t, call.message, notification.Text, "Call %d: notification text mismatch", i)
				default:
					assert.False(t, call.expectNotification, "Call %d: CheckAndNotify() did not send notification when one was expected", i)
				}
			}
		})
	}
}
