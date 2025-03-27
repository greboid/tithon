package templates

import "github.com/greboid/ircclient/irc"

type ServerList struct {
	Connections  []*irc.Connection
	ActiveWindow *irc.Window
}
