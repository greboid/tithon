package irc

import (
	"github.com/greboid/tithon/config"
	"reflect"
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
