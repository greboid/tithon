package templates

import "github.com/greboid/ircclient/irc"

type Window struct {
	WindowInfo string
	Messages   []*irc.Message
	Users      []string
}
