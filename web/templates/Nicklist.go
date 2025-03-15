package templates

import "github.com/greboid/ircclient/irc"

type Nicklist struct {
	Users []*irc.User
}
