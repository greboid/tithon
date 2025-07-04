package services

import (
	"fmt"
	"github.com/greboid/tithon/config"
	semver "github.com/hashicorp/go-version"
	"runtime/debug"
	"strings"
)

type SettingsData struct {
	Version         string
	TimestampFormat string
	ShowNicklist    bool
	Servers         []config.Server
	Notifications   []config.NotificationTrigger
	Theme           string
}

// SettingsService manages UI settings and configuration
type SettingsService struct {
	conf         *config.Config
	settingsData *SettingsData
}

func NewSettingsService(conf *config.Config) *SettingsService {
	return &SettingsService{
		conf: conf,
		settingsData: &SettingsData{
			Version:         getVersion(),
			TimestampFormat: conf.UISettings.TimestampFormat,
			ShowNicklist:    conf.UISettings.ShowNicklist,
			Servers:         conf.Servers,
			Notifications:   conf.Notifications.Triggers,
			Theme:           conf.UISettings.Theme,
		},
	}
}

func (ss *SettingsService) GetFromConfig() *SettingsData {
	ss.settingsData = &SettingsData{
		Version:         getVersion(),
		TimestampFormat: ss.conf.UISettings.TimestampFormat,
		ShowNicklist:    ss.conf.UISettings.ShowNicklist,
		Servers:         make([]config.Server, len(ss.conf.Servers)),
		Notifications:   make([]config.NotificationTrigger, len(ss.conf.Notifications.Triggers)),
		Theme:           ss.conf.UISettings.Theme,
	}
	copy(ss.settingsData.Servers, ss.conf.Servers)
	copy(ss.settingsData.Notifications, ss.conf.Notifications.Triggers)
	return ss.settingsData
}

func (ss *SettingsService) GetSettingsData() *SettingsData {
	return ss.settingsData
}

func (ss *SettingsService) SaveSettingsToConfig() error {
	ss.conf.UISettings.TimestampFormat = ss.settingsData.TimestampFormat
	ss.conf.UISettings.ShowNicklist = ss.settingsData.ShowNicklist
	ss.conf.UISettings.Theme = ss.settingsData.Theme
	ss.conf.Notifications.Triggers = make([]config.NotificationTrigger, len(ss.settingsData.Notifications))
	copy(ss.conf.Notifications.Triggers, ss.settingsData.Notifications)
	ss.conf.Servers = make([]config.Server, len(ss.settingsData.Servers))
	copy(ss.conf.Servers, ss.settingsData.Servers)
	return ss.conf.Save()
}

func getVersion() string {
	var versionString string
	if info, ok := debug.ReadBuildInfo(); ok {
		versionString = info.Main.Version
		if version, err := semver.NewVersion(versionString); err == nil {
			versionString = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(version.Segments()[0:3])), "."), "[]")
			if version.Prerelease() != "" {
				versionString = versionString + "-dev"
			}
		} else {
			versionString = "err"
		}
	} else {
		versionString = "unknown"
	}
	return versionString
}
