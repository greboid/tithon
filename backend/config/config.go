package config

import (
	"github.com/csmith/config"
	"log/slog"
	"time"
)

type Config struct {
	instance      *config.Config
	Servers       []Server      `yaml:"servers"`
	UISettings    UISettings    `yaml:"ui_settings"`
	Notifications Notifications `yaml:"notifications"`
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
}

type UISettings struct {
	TimestampFormat string `yaml:"timestamp_format"`
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
}

func (c *Config) Load() error {
	slog.Debug("Loading config")
	conf, err := config.New(config.DirectoryName("tithon"), config.FileName("config.yaml"))
	if err != nil {
		return err
	}
	err = conf.Load(c)
	if err != nil {
		return err
	}
	c.instance = conf
	if c.UISettings.TimestampFormat == "" {
		c.UISettings.TimestampFormat = time.TimeOnly
	}
	return nil
}

func (c *Config) Save() error {
	return c.instance.Save(c)
}
