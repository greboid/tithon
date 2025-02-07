package irc

import (
	"github.com/ergochat/irc-go/ircevent"
)

const (
	v3TimestampFormat = "2006-01-02T15:04:05.000Z"
)

type Server struct {
	Hostname     string
	Port         int
	TLS          bool
	SaslMech     string
	Saslusername string
	Saslpassword string
}

type Profile struct {
	Nick     string
	User     string
	Realname string
}

type Client struct {
	ircevent.Connection
}
