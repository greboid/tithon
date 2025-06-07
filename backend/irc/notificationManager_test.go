package irc

import (
	"github.com/greboid/tithon/config"
	"reflect"
	"regexp"
	"testing"
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
			nm := &NotificationManager{}
			got := nm.sortTriggers(tt.triggers)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortTriggers() = %v, want %v", got, tt.want)
			}
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
			nm := &NotificationManager{}
			got, err := nm.compileRegex(tt.regex)
			if (err != nil) != tt.wantErr {
				t.Errorf("compileRegex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if got.String() != tt.want {
					t.Errorf("compileRegex() got = %v, want %v", got.String(), tt.want)
				}
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
			nm := &NotificationManager{}
			got, err := nm.AddNotification(tt.network, tt.source, tt.nick, tt.message, tt.sound, tt.popup)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddNotification() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got.Network.String() != tt.want.Network.String() {
				t.Errorf("AddNotification() Network = %v, want %v", got.Network.String(), tt.want.Network.String())
			}
			if got.Source.String() != tt.want.Source.String() {
				t.Errorf("AddNotification() Source = %v, want %v", got.Source.String(), tt.want.Source.String())
			}
			if got.Nick.String() != tt.want.Nick.String() {
				t.Errorf("AddNotification() Nick = %v, want %v", got.Nick.String(), tt.want.Nick.String())
			}
			if got.Message.String() != tt.want.Message.String() {
				t.Errorf("AddNotification() Message = %v, want %v", got.Message.String(), tt.want.Message.String())
			}
			if got.Sound != tt.want.Sound {
				t.Errorf("AddNotification() Sound = %v, want %v", got.Sound, tt.want.Sound)
			}
			if got.Popup != tt.want.Popup {
				t.Errorf("AddNotification() Popup = %v, want %v", got.Popup, tt.want.Popup)
			}
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
			nm := &NotificationManager{}
			got := nm.convertFromConfig(tt.triggers)

			if len(got) != len(tt.want) {
				t.Errorf("convertFromConfig() got %d triggers, want %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i].Network.String() != tt.want[i].Network.String() {
					t.Errorf("convertFromConfig() trigger[%d].Network = %v, want %v", i, got[i].Network.String(), tt.want[i].Network.String())
				}
				if got[i].Source.String() != tt.want[i].Source.String() {
					t.Errorf("convertFromConfig() trigger[%d].Source = %v, want %v", i, got[i].Source.String(), tt.want[i].Source.String())
				}
				if got[i].Nick.String() != tt.want[i].Nick.String() {
					t.Errorf("convertFromConfig() trigger[%d].Nick = %v, want %v", i, got[i].Nick.String(), tt.want[i].Nick.String())
				}
				if got[i].Message.String() != tt.want[i].Message.String() {
					t.Errorf("convertFromConfig() trigger[%d].Message = %v, want %v", i, got[i].Message.String(), tt.want[i].Message.String())
				}
				if got[i].Sound != tt.want[i].Sound {
					t.Errorf("convertFromConfig() trigger[%d].Sound = %v, want %v", i, got[i].Sound, tt.want[i].Sound)
				}
				if got[i].Popup != tt.want[i].Popup {
					t.Errorf("convertFromConfig() trigger[%d].Popup = %v, want %v", i, got[i].Popup, tt.want[i].Popup)
				}
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
			nm := &NotificationManager{
				notifications:        tt.notifications,
				pendingNotifications: notificationChan,
			}
			nm.CheckAndNotify(tt.network, tt.source, tt.nick, tt.message)

			var receivedNotification Notification
			select {
			case receivedNotification = <-notificationChan:
				if !tt.expectNotification {
					t.Errorf("CheckAndNotify() sent a notification when none was expected")
				}
				if receivedNotification.Title != tt.expectedTitle {
					t.Errorf("CheckAndNotify() notification title = %v, want %v", receivedNotification.Title, tt.expectedTitle)
				}
				if receivedNotification.Text != tt.expectedText {
					t.Errorf("CheckAndNotify() notification text = %v, want %v", receivedNotification.Text, tt.expectedText)
				}
				if receivedNotification.Sound != tt.expectedSound {
					t.Errorf("CheckAndNotify() notification sound = %v, want %v", receivedNotification.Sound, tt.expectedSound)
				}
				if receivedNotification.Popup != tt.expectedPopup {
					t.Errorf("CheckAndNotify() notification popup = %v, want %v", receivedNotification.Popup, tt.expectedPopup)
				}
			default:
				if tt.expectNotification {
					t.Errorf("CheckAndNotify() did not send a notification when one was expected")
				}
			}
		})
	}
}
