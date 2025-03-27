package templates

import "github.com/greboid/ircclient/irc"

type Index struct {
	Connections  []*irc.Connection
	ActiveWindow *irc.Window
	WindowInfo   string
	Messages     []*irc.Message
	Users        []*irc.User
}
