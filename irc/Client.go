package irc

import (
	"github.com/ergochat/irc-go/ircevent"
)

const (
	v3TimestampFormat = "2006-01-02T15:04:05.000Z"
)

type Server struct {
	Name string `json:"name"`
}

type ConnectableServer struct {
	Server       string             `json:"server"`
	TLS          bool               `json:"tls"`
	SaslMech     string             `json:"saslMech,omitempty"`
	Saslusername string             `json:"saslUsername,omitempty"`
	Saslpassword string             `json:"saslPassword,omitempty"`
	Profile      ConnectableProfile `json:"profile"`
}

type ConnectableProfile struct {
	Nick     string `json:"nick"`
	User     string `json:"user,omitempty"`
	Realname string `json:"realname,omitempty"`
}

type Client struct {
	ircevent.Connection
}
