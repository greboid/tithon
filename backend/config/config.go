package config

import (
	"log/slog"
	"time"
)

type Config struct {
	instance      Provider
	Servers       []Server      `yaml:"servers"`
	UISettings    UISettings    `yaml:"ui_settings"`
	Notifications Notifications `yaml:"notifications"`
}

func NewConfig(provider Provider) *Config {
	return &Config{
		instance: provider,
	}
}

type Server struct {
	Hostname     string  `yaml:"hostname"`
	Port         int     `yaml:"port"`
	TLS          bool    `yaml:"tls"`
	Password     string  `yaml:"password"`
	SASLLogin    string  `yaml:"sasl_login"`
	SASLPassword string  `yaml:"sasl_password"`
	Profile      Profile `yaml:"profile"`
	ID           string  `yaml:"id"`
	AutoConnect  bool    `yaml:"auto_connect"`
}

type UISettings struct {
	TimestampFormat string `yaml:"timestamp_format"`
	ShowNicklist    bool   `yaml:"show_nicklist"`
	Theme           string `yaml:"theme"`
}

type Profile struct {
	Nickname string `yaml:"nickname"`
}

type Notifications struct {
	Triggers []NotificationTrigger `yaml:"triggers"`
}

type NotificationTrigger struct {
	Network string `yaml:"network"`
	Source  string `yaml:"source"`
	Nick    string `yaml:"nick"`
	Message string `yaml:"message"`
	Sound   bool   `yaml:"sound"`
	Popup   bool   `yaml:"popup"`
}

func (c *Config) Load() error {
	slog.Debug("Loading config")
	if err := c.instance.Load(c); err != nil {
		return err
	}
	if c.UISettings.TimestampFormat == "" {
		c.UISettings.TimestampFormat = time.TimeOnly
	}
	return nil
}

func (c *Config) Save() error {
	return c.instance.Save(c)
}
