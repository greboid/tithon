package config

import (
	"github.com/csmith/config"
	"log/slog"
)

type Config struct {
	instance *config.Config
	Servers  []Server `yaml:"servers"`
}

type Server struct {
	Hostname     string  `yaml:"hostname"`
	Port         int     `yaml:"port"`
	TLS          bool    `yaml:"tls"`
	Password     string  `yaml:"password"`
	SASLLogin    string  `yaml:"sasl_login"`
	SASLPassword string  `yaml:"sasl_password"`
	Profile      Profile `yaml:"profile"`
}

type Profile struct {
	Nickname string `yaml:"nickname"`
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
