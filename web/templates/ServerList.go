package templates

import "github.com/greboid/ircclient/irc"

type ServerList struct {
	Connections   []*irc.Connection
	ActiveServer  *irc.Connection
	ActiveChannel *irc.Channel
}
