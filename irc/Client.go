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
	Hostname     string `json:"hostname"`
	Port         int    `json:"port"`
	TLS          bool   `json:"tls"`
	SaslMech     string `json:"saslMech,omitempty"`
	Saslusername string `json:"saslUsername,omitempty"`
	Saslpassword string `json:"saslPassword,omitempty"`
}

type ConnectableProfile struct {
	Nick     string `json:"nick"`
	User     string `json:"user,omitempty"`
	Realname string `json:"readlname,omitempty"`
}

type Client struct {
	ircevent.Connection
}
