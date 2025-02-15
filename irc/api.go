package irc

import (
	"context"
)

type ConfigServer struct {
	Server       string        `yaml:"server"`
	TLS          bool          `yaml:"tls"`
	SaslUsername string        `yaml:"saslUsername"`
	SaslPassword string        `yaml:"saslPassword"`
	Profile      ConfigProfile `yaml:"profile"`
}

type ConfigProfile struct {
	Nick string `yaml:"nick"`
	User string `yaml:"user"`
}

type Config struct {
	Servers []ConfigServer `yaml:"servers"`
}

type App struct {
	Ctx     context.Context
	EE      EventEmitter
	Servers []Connection
}

type NullEmitter struct{}

type EventEmitter interface {
	Emit(eventName string, data ...interface{})
}

type Connection interface {
	Init(ctx context.Context, server string, useTLS bool, SASLLogin string, SASLPassword string, PreferredNick string)
	CurrentNick() string
	GetChanTypes() string
	Connect()
	Loop()
	ID() string
	Name() string
}
