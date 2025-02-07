package irc

import (
	"github.com/ergochat/irc-go/ircevent"
	"newirc/gui"
)

const (
	v3TimestampFormat = "2006-01-02T15:04:05.000Z"
)

type Client struct {
	ircevent.Connection
	App *gui.App
}
