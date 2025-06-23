package irc

import (
	uniqueid "github.com/albinj12/unique-id"
)

type Query struct {
	*Window
}

func NewQuery(connection ServerInterface, name string) *Query {
	s, _ := uniqueid.Generateid("a", 5, "p")
	query := &Query{
		Window: &Window{
			id:         s,
			name:       name,
			messages:   make([]*Message, 0),
			connection: connection,
			isQuery:    true,
		},
	}
	query.Window.tabCompleter = NewQueryTabCompleter(query)
	return query
}
