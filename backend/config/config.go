package config

import (
	"github.com/csmith/config"
	"log/slog"
)

type Config struct {
	instance *config.Config
	Servers  []Server `json:"servers"`
}

type Server struct {
	Hostname     string  `json:"hostname"`
	Port         int     `json:"port"`
	TLS          bool    `json:"tls"`
	Password     string  `json:"password"`
	SASLLogin    string  `json:"sasllogin"`
	SASLPassword string  `json:"saslpassword"`
	Profile      Profile `json:"profile"`
}

type Profile struct {
	Nickname string `json:"nickname"`
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
	return nil
}

func (c *Config) Save() error {
	return c.instance.Save(c)
}
